package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"math"
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

func (proc *ForkProcess) Wait4() bool {
	// initialize and get status
	var ws unix.WaitStatus
	wpid, err := unix.Wait4(-1*proc.Pgid, &ws, unix.WALL, nil)
	if err != nil {
		util.Warn("wait4: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return true
	}

	if wpid == -1 {
		util.Warn("wpid = -1")
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: "wpid = -1"}
		return true
	}

	// check if segfault, or other stuff
	// http://people.cs.pitt.edu/~alanjawi/cs449/code/shell/UnixSignals.htm

	if isStopSignal(ws.StopSignal()) {
		proc.Session.ExitCode = int(ws.StopSignal())
		// this object will be filled in by judge channel
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_RTE, ResultInfo: "",}
		proc.Kill()
		return true
	}

	// if process has already exited, leave
	if ws.Exited() {
		proc.Session.ExitCode = int(ws.StopSignal())
		return true
	}
	return false
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

func (proc *ForkProcess) Trace() {
	var err error

	proc.Pgid, err = unix.Getpgid(proc.Pid)
	if err != nil {
		util.Warn("getpgid: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	proc.Wait4()

	err = unix.PtraceSetOptions(proc.Pid, unix.PTRACE_O_EXITKILL|unix.PTRACE_O_TRACESYSGOOD|unix.PTRACE_O_TRACEEXIT|unix.PTRACE_O_TRACECLONE|unix.PTRACE_O_TRACEFORK|unix.PTRACE_O_TRACEVFORK)
	if err != nil {
		util.Warn("ptracesetoptions: " + err.Error())
		proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	for { // scan through each syscall
		if proc.Syscall() {
			return
		}
		if proc.Wait4() {
			return
		}

		// get system call
		pregs := unix.PtraceRegs{} // user regs struct
		err = unix.PtraceGetRegs(proc.Pid, &pregs)
		if err != nil {
			util.Warn("ptracegetregs: " + err.Error())
			proc.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		// map syscall to nothing if syscall is blockedCall
		blockedCall := blockRestrictedCalls(&pregs, proc.Pid)

		// run system call
		if proc.Syscall() {
			return
		}
		if proc.Wait4() {
			return
		}
		if blockedCall {
			pregs.Rax = uint64(math.Inf(0))
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