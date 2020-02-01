package judging

import (
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
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

	var err error

	proc.Pgid, err = unix.Getpgid(proc.Pid)
	if err != nil {
		util.Warn("getpgid: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	var (
		ws unix.WaitStatus
		rusage unix.Rusage
		)

	// init tracing
	pid, err := unix.Wait4(-proc.Pgid, &ws, unix.WALL, nil)
	if err != nil {
		util.Warn("wait4 error: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}
	shared.Debug(fmt.Sprintf("wait4: pickup ptrace %v %v", ws.Stopped(), pid))

	err = unix.PtraceSetOptions(proc.Pid, unix.PTRACE_O_TRACESECCOMP|unix.PTRACE_O_EXITKILL|unix.PTRACE_O_TRACEFORK|unix.PTRACE_O_TRACECLONE|unix.PTRACE_O_TRACEEXEC|unix.PTRACE_O_TRACEVFORK)
	if err != nil {
		util.Warn("ptracesetoptions: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	unix.PtraceCont(pid, int(ws.StopSignal())) // restart process

	// trace each syscall
	for {
		pid, err := unix.Wait4(-proc.Pgid, &ws, unix.WALL, nil)
		if err != nil {
			util.Warn("wait4 error: " + err.Error())
			proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		// check judging
		if proc.Session.CheckProcState(&ws, &rusage) {
			return
		}

		// check process status
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
					err := handleTrap(pid, proc) // handle syscall
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