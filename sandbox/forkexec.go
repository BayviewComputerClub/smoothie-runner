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

type ForkExecContext struct {
	Pid  uintptr
	Pipe [2]int

	ArgV []*byte
	EnvV []*byte
}

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

func (session *RunnerSession) ForkExec() error {
	var (
		err     error
		err1    unix.Errno
		context ForkExecContext
	)

	// prepare for execveat
	_, context.ArgV, context.EnvV, err = prepareExec(session.ExecArgs, session.ExecEnv)
	if err != nil {
		util.Warn("forkexec prepareexec: " + err.Error())
		return err
	}

	// create pipe
	context.Pipe, err = syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)
	if err != nil {
		util.Warn("forkexec socketpair: " + err.Error())
		return err
	}

	syscall.ForkLock.Lock()

	// fork
	context.Pid, _, err1 = unix.Syscall6(syscall.SYS_CLONE, uintptr(unix.SIGCHLD), 0, 0, 0, 0, 0)

	if err1 != 0 || context.Pid != 0 {
		if err1 != 0 {
			util.Warn("forkexec clone: " + err1.Error())
			return err1
		}
		// -=-=- in parent process -=-=-
		return session.ForkExecParent(context)
	}

	// -=-=- in child process -=-=-
	session.ForkExecChild(context)
	util.Fatal("Fork exec in child error, leaving child...")
	return nil // should not be able to reach
}

func (session *RunnerSession) ForkExecParent(context ForkExecContext) error {
	var (
		err        error
		err1, err2 unix.Errno
		r1         uintptr
	)

	syscall.ForkLock.Unlock()

	unix.Close(context.Pipe[1])

	// child returned error code, and sync
	r1, _, err1 = unix.RawSyscall(syscall.SYS_READ, uintptr(context.Pipe[0]), uintptr(unsafe.Pointer(&err2)), unsafe.Sizeof(err2))
	if r1 != unsafe.Sizeof(err2) || err2 != 0 || err1 != 0 {
		unix.Close(context.Pipe[0])
		if r1 == unsafe.Sizeof(err2) {
			err = err2
		}
		if err == nil {
			err = syscall.EPIPE
		}
		util.Warn("forkexec execread: " + err.Error())
		handleChildFailed(context.Pid)
		return err
	}
	unix.RawSyscall(syscall.SYS_WRITE, uintptr(context.Pipe[0]), uintptr(unsafe.Pointer(&err1)), unsafe.Sizeof(err1))

	session.Pid = int(context.Pid)

	unix.Close(context.Pipe[0])
	return nil
}

func (session *RunnerSession) ForkExecChild(context ForkExecContext) {
	// -=-=- child forked process -=-=-

	var (
		err        error
		err1, err2 unix.Errno
		r1         uintptr
	)

	pipe := context.Pipe[1]

	if err := unix.Close(context.Pipe[0]); err != nil {
		forkLeaveError(pipe, err)
		return
	}

	pid, _, err := syscall.RawSyscall(unix.SYS_GETPID, 0, 0, 0)

	_, _, err1 = syscall.RawSyscall(syscall.SYS_SETPGID, 0, 0, 0)
	if err1 != 0 {
		forkLeaveError(pipe, err1)
		return
	}

	// change workspace
	err = unix.Chdir(session.Workspace)
	if err != nil {
		forkLeaveError(pipe, err)
		return
	}

	// set specified file descriptors (ex. stdin - 0, stdout - 1, stderr - 2)
	for k, v := range session.Files {
		if err := unix.Dup2(int(v), k); err != nil {
			forkLeaveError(pipe, err)
			return
		}
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

	if session.SandboxWithSeccomp && shared.SANDBOX {
		// ptrace
		_, _, err1 = unix.RawSyscall(unix.SYS_PTRACE, uintptr(unix.PTRACE_TRACEME), 0, 0)
		if err1 != 0 {
			forkLeaveError(pipe, err1)
			return
		}

		//f, _ := os.Create(strconv.Itoa(int(time.Now().UnixNano())) + "debug")

		// wait for tracer
		_, _, err1 = unix.RawSyscall(unix.SYS_KILL, pid, uintptr(unix.SIGSTOP), 0)
		if err1 != 0 {
			forkLeaveError(pipe, err1)
			return
		}

		_, _, err1 = syscall.RawSyscall6(syscall.SYS_PRCTL, unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0, 0)
		if err1 != 0 {
			forkLeaveError(pipe, err1)
			return
		}

		// seccomp
		// calls prctl PR_SET_NO_NEW_PRIVS as well
		_, _, err1 = unix.RawSyscall(unix.SYS_SECCOMP, 1 /* SECCOMP_SET_MODE_FILTER */, 1 /* SECCOMP_FILTER_FLAG_TSYNC */, uintptr(unsafe.Pointer(session.Seccomp)))
		if err1 != 0 {
			forkLeaveError(pipe, err1)
			return
		}

		// close all file descriptors in a brutal fashion :)
		for i := 0; i < sysconf.SC_OPEN_MAX; i++ {
			// check if file descriptor is not used
			if _, ok := session.Files[i]; !ok {
				_ = unix.Close(i)
			}
		}
	}

	// set resource limits - rlimit is restricted once exeveat is called
	if shared.RLIMITS {
		err = session.SetRlimits()
		if err != nil {
			forkLeaveError(pipe, err)
			return
		}
	}

	// execute process, now replaced by new process
	_, _, err1 = syscall.RawSyscall6(unix.SYS_EXECVEAT, session.ExecFile, uintptr(unsafe.Pointer(&empty[0])), uintptr(unsafe.Pointer(&context.ArgV[0])), uintptr(unsafe.Pointer(&context.EnvV[0])), unix.AT_EMPTY_PATH, 0)
}

func forkLeaveError(pipe int, err error) {
	util.Warn("child: " + err.Error())
	syscall.RawSyscall(unix.SYS_WRITE, uintptr(pipe), uintptr(unsafe.Pointer(&err)), unsafe.Sizeof(err))
}

func handleChildFailed(pid uintptr) {
	// make sure not blocked
	syscall.Kill(int(pid), syscall.SIGKILL)

	// child failed; wait for it to exit, to make sure the zombies don't accumulate
	var wstatus syscall.WaitStatus
	_, err := syscall.Wait4(int(pid), &wstatus, 0, nil)
	for err == syscall.EINTR {
		_, err = syscall.Wait4(int(pid), &wstatus, 0, nil)
	}
}
