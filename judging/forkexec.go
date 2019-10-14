package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"syscall"
	"unsafe"
)

// thank you https://github.com/criyle/go-sandbox/tree/master/pkg/forkexec
// or else this would be bad

var empty = [...]byte{0}

func prepareExec(Args, Env []string) (*byte, []*byte, []*byte, error) {
	// make exec args0
	argv0, err := syscall.BytePtrFromString(Args[0])
	if err != nil {
		return nil, nil, nil, err
	}
	// make exec args
	argv, err := syscall.SlicePtrFromStrings(Args)
	if err != nil {
		return nil, nil, nil, err
	}
	// make env
	envv, err := syscall.SlicePtrFromStrings(Env)
	if err != nil {
		return nil, nil, nil, err
	}
	return argv0, argv, envv, nil
}

func (tracer *PTracer) ForkExec() {
	var (
		err1, err2 unix.Errno
		r1 uintptr
	)

	_, argv, envv, err := prepareExec(tracer.Session.Command.Args, []string{})
	if err != nil {
		util.Warn("forkexec prepareexec: " + err.Error())
		tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	// create pipe
	p, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)
	if err != nil {
		util.Warn("forkexec socketpair: " + err.Error())
		tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	pipe := p[1]

	syscall.ForkLock.Lock()

	pid, _, err1 := unix.Syscall6(syscall.SYS_CLONE, uintptr(unix.SIGCHLD), 0, 0, 0, 0, 0)
	if err1 != 0 || pid != 0 {
		syscall.ForkLock.Unlock()

		// in parent process

		unix.Close(p[1])

		if err1 != 0 {
			util.Warn("forkexec clone: " + err.Error())
			tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		// child returned error code
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
			tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		syscall.RawSyscall(syscall.SYS_WRITE, uintptr(p[0]), uintptr(unsafe.Pointer(&err1)), unsafe.Sizeof((err1)))

		tracer.Pid = int(pid)
		return
	}

	// now in the child OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO
	// no more golang calls

	if _, _, err1 = syscall.RawSyscall(syscall.SYS_CLOSE, uintptr(p[0]), 0, 0); err1 != 0 {
		goto justleave
	}

	pid, _, err = syscall.RawSyscall(unix.SYS_GETPID, 0, 0, 0)

	_, _, err1 = syscall.RawSyscall(syscall.SYS_SETPGID, 0, 0, 0)
	if err1 != 0 {
		goto justleave
	}

	// set stdin, stdout, stderr file descriptors
	_, _, err1 = syscall.RawSyscall(syscall.SYS_DUP2, tracer.Session.InputStream.Fd(), 0, 0)
	if err1 != 0 {
		goto justleave
	}
	_, _, err1 = syscall.RawSyscall(syscall.SYS_DUP2, tracer.Session.OutputStream.Fd(), 1, 0)
	if err1 != 0 {
		goto justleave
	}
	_, _, err1 = syscall.RawSyscall(syscall.SYS_DUP2, tracer.Session.ErrorStream.Fd(), 2, 0)
	if err1 != 0 {
		goto justleave
	}

	// sync
	err2 = 0
	r1, _, err1 = syscall.RawSyscall(syscall.SYS_WRITE, uintptr(pipe), uintptr(unsafe.Pointer(&err2)), unsafe.Sizeof(err2))
	if r1 == 0 || err1 != 0 {
		goto justleave
	}

	r1, _, err1 = syscall.RawSyscall(syscall.SYS_READ, uintptr(pipe), uintptr(unsafe.Pointer(&err2)), unsafe.Sizeof(err2))
	if r1 == 0 || err1 != 0 {
		goto justleave
	}

	if shared.SANDBOX {
		_, _, err1 = syscall.RawSyscall(syscall.SYS_PTRACE, uintptr(syscall.PTRACE_TRACEME), 0, 0)
		if err1 != 0 {
			goto justleave
		}
	}

	// TODO change working dir

	// execute process
	_, _, err1 = syscall.RawSyscall6(unix.SYS_EXECVEAT, tracer.ExecCommand, uintptr(unsafe.Pointer(&empty[0])), uintptr(unsafe.Pointer(&argv[0])), uintptr(unsafe.Pointer(&envv[0])), unix.AT_EMPTY_PATH, 0)

	justleave:
		// send error code on pipe
		syscall.RawSyscall(unix.SYS_WRITE, uintptr(pipe), uintptr(unsafe.Pointer(&err1)), unsafe.Sizeof(err1))
	for {
		syscall.RawSyscall(syscall.SYS_EXIT, uintptr(err1+err2), 0, 0)
	}
	// cannot reach this point
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