package sandbox

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"github.com/tklauser/go-sysconf"
	"golang.org/x/sys/unix"
	"syscall"
	"unsafe"
)

// thank you https://github.com/criyle/go-sandbox/tree/master/pkg/forkexec
// or else this would be bad

var empty = [...]byte{0}

func prepareExec(args, env []string) (*byte, []*byte, []*byte, error) {
	// make exec args0
	argv0, err := syscall.BytePtrFromString(args[0])
	if err != nil {
		return nil, nil, nil, err
	}
	// make exec args
	argv, err := syscall.SlicePtrFromStrings(args)
	if err != nil {
		return nil, nil, nil, err
	}
	// make env
	envv, err := syscall.SlicePtrFromStrings(env)
	if err != nil {
		return nil, nil, nil, err
	}
	return argv0, argv, envv, nil
}

func (session *RunnerSession) ForkExec() {
	var (
		err1, err2 unix.Errno
		r1 uintptr
	)

	_, argv, envv, err := prepareExec(session.ExecArgs, session.Env)
	if err != nil {
		util.Warn("forkexec prepareexec: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	// create pipe
	p, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)
	if err != nil {
		util.Warn("forkexec socketpair: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	pipe := p[1]

	syscall.ForkLock.Lock()

	pid, _, err1 := unix.Syscall6(syscall.SYS_CLONE, uintptr(unix.SIGCHLD), 0, 0, 0, 0, 0)

	if err1 != 0 || pid != 0 {
		// -=-=- in parent process -=-=-

		syscall.ForkLock.Unlock()

		unix.Close(p[1])

		if err1 != 0 {
			util.Warn("forkexec clone: " + err.Error())
			proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		// child returned error code, and sync
		r1, _, err1 := unix.RawSyscall(syscall.SYS_READ, uintptr(p[0]), uintptr(unsafe.Pointer(&err2)), unsafe.Sizeof(err2))
		if r1 != unsafe.Sizeof(err2) || err2 != 0 || err1 != 0 {
			unix.Close(p[0])
			if r1 == unsafe.Sizeof(err2) {
				err = err2
			}
			if err == nil {
				err = syscall.EPIPE
			}
			handleChildFailed(pid)
			util.Warn("forkexec execread: " + err.Error())
			proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}
		unix.RawSyscall(syscall.SYS_WRITE, uintptr(p[0]), uintptr(unsafe.Pointer(&err1)), unsafe.Sizeof(err1))

		proc.Pid = int(pid)
		unix.Close(p[0])

		return
	}

	// -=-=- child forked process -=-=-

	if err := unix.Close(p[0]); err != nil {
		forkLeaveError(pipe, err)
		return
	}

	pid, _, err = syscall.RawSyscall(unix.SYS_GETPID, 0, 0, 0)

	_, _, err1 = syscall.RawSyscall(syscall.SYS_SETPGID, 0, 0, 0)
	if err1 != 0 {
		forkLeaveError(pipe, err1)
		return
	}

	// change workspace
	err = unix.Chdir(proc.Session.Command.Dir)
	if err != nil {
		forkLeaveError(pipe, err)
		return
	}

	// set stdin, stdout, stderr file descriptors
	if err := unix.Dup2(int(proc.Session.InputStream.Fd()), 0) ; err != nil {
		forkLeaveError(pipe, err)
		return
	}
	if err := unix.Dup2(int(proc.Session.OutputStream.Fd()), 1) ; err != nil {
		forkLeaveError(pipe, err)
		return
	}
	if err := unix.Dup2(int(proc.Session.ErrorStream.Fd()), 2) ; err != nil {
		forkLeaveError(pipe, err)
		return
	}

	// close all file descriptors in a brutal fashion :)
	for i := 3; i < sysconf.SC_OPEN_MAX; i++ {
		_ = unix.Close(i)
	}

	// set resource limits
	err = proc.SetRlimits()
	if err != nil {
		forkLeaveError(pipe, err)
		return
	}

	// sync with parent
	err2 = 0
	r1, _, err1 = syscall.RawSyscall(syscall.SYS_WRITE, uintptr(pipe), uintptr(unsafe.Pointer(&err2)), unsafe.Sizeof(err2))
	if r1 == 0 || err1 != 0 {
		forkLeaveError(pipe, err1)
		return
	}

	r1, _, err1 = syscall.RawSyscall(syscall.SYS_READ, uintptr(pipe), uintptr(unsafe.Pointer(&err2)), unsafe.Sizeof(err2))
	if r1 == 0 || err1 != 0 {
		forkLeaveError(pipe, err1)
		return
	}
	if shared.SANDBOX {
		// ptrace
		_, _, err1 = unix.RawSyscall(unix.SYS_PTRACE, uintptr(unix.PTRACE_TRACEME), 0, 0)
		if err1 != 0 {
			forkLeaveError(pipe, err1)
			return
		}

		// wait for tracer
		_,_, err1 = unix.RawSyscall(unix.SYS_KILL, pid, uintptr(unix.SIGSTOP), 0)
		if err1 != 0 {
			forkLeaveError(pipe, err1)
			return
		}

		// seccomp
		err = proc.LoadSeccompFilter() // calls prctl set no privs as well
		if err != nil {
			forkLeaveError(pipe, err)
			return
		}
	}

	// execute process, now replaced by new process
	_, _, err1 = syscall.RawSyscall6(unix.SYS_EXECVEAT, proc.ExecCommand, uintptr(unsafe.Pointer(&empty[0])), uintptr(unsafe.Pointer(&argv[0])), uintptr(unsafe.Pointer(&envv[0])), unix.AT_EMPTY_PATH, 0)
}

func forkLeaveError(pipe int, err error) {
	util.Warn("child: " + err.Error())
	syscall.RawSyscall(unix.SYS_WRITE, uintptr(pipe), uintptr(unsafe.Pointer(&err)), unsafe.Sizeof(err))
	for {
		unix.Exit(0)
	}
}

func handleChildFailed(pid uintptr) {
	var wstatus syscall.WaitStatus
	// make sure not blocked
	syscall.Kill(int(pid), syscall.SIGKILL)
	// child failed; wait for it to exit, to make sure the zombies don't accumulate
	_, err := syscall.Wait4(int(pid), &wstatus, 0, nil)
	for err == syscall.EINTR {
		_, err = syscall.Wait4(int(pid), &wstatus, 0, nil)
	}
}