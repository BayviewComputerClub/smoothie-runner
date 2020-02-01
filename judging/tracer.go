package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"runtime"
	"strconv"
)

// process that is sandboxed
type ForkProcess struct {
	Session *GradeSession

	Pid int
	Pgid int

	StreamDone chan CaseReturn

	ExecCommand uintptr
	ExecArgs uintptr
}

func (proc *ForkProcess) Syscall() bool {
	err := unix.PtraceSyscall(proc.Pid, 0)
	if err != nil {
		util.Warn("ptracesyscall: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return true
	}
	return false
}

// start tracer
func (proc *ForkProcess) Trace() {
	runtime.LockOSThread() // https://github.com/golang/go/issues/7699
	defer runtime.UnlockOSThread()

	defer func() {
		if err := recover(); err != nil {
			util.Warn("panic in tracer")
		}
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: "panic in tracer"}
	}()

	var err error

	proc.Pgid, err = unix.Getpgid(proc.Pid)
	if err != nil {
		util.Warn("getpgid: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	// set ptrace options
	var ws unix.WaitStatus

	_, err = unix.Wait4(-proc.Pgid, &ws, unix.WALL, nil)
	if err != nil {
		util.Warn("wait4: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	err = unix.PtraceSetOptions(proc.Pid, unix.PTRACE_O_TRACESECCOMP|unix.PTRACE_O_EXITKILL|unix.PTRACE_O_TRACEFORK|unix.PTRACE_O_TRACECLONE|unix.PTRACE_O_TRACEEXEC|unix.PTRACE_O_TRACEVFORK)
	if err != nil {
		util.Warn("ptracesetoptions: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	// trace each syscall
	for {
		pid, err := unix.Wait4(-proc.Pgid, &ws, unix.WALL, nil)
		if err != nil {
			util.Warn("wait4: " + err.Error())
			proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		switch {
		case ws.Exited():
			shared.Debug("tracer: process exited with " + strconv.Itoa(ws.ExitStatus()))
			return

		case ws.Signaled():
			err := unix.PtraceCont(pid, int(ws.Signal()))
			if err != nil {
				util.Warn("ptracecont signaled: " + err.Error())
				proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
				return
			}

		case ws.Stopped():
			if ws.StopSignal() == unix.SIGTRAP {
				switch ws.TrapCause() {
				case unix.PTRACE_EVENT_SECCOMP:
					shared.Debug("ptrace trap event_seccomp")
					err := handleTrap(pid, proc)
					if err != nil {
						proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ILL, ResultInfo: err.Error()}
						return
					}
				}
				err = unix.PtraceCont(pid, 0)
			} else {

				switch ws.StopSignal() {
				case unix.SIGXCPU:
					proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_TLE}
					return
				case unix.SIGXFSZ:
					proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_OLE}
					return
				}

				err = unix.PtraceCont(pid, int(ws.StopSignal()))
			}

			if err != nil {
				util.Warn("ptracecont signaled: " + err.Error())
				proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
				return
			}

		}
	}

}

func (proc *ForkProcess) Kill() {
	var ws unix.WaitStatus
	unix.Kill(proc.Pid, unix.SIGKILL)
	_, err := unix.Wait4(proc.Pid, &ws, 0, nil)
	for err == unix.EINTR {
		_, err = unix.Wait4(proc.Pid, &ws, 0, nil)
	}
}