package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"math"
)

type PTracer struct {
	Session *GradeSession

	Pid int
	Pgid int

	StreamDone chan CaseReturn

	ExecCommand uintptr
	ExecArgs uintptr
}

func (tracer *PTracer) Wait4() bool {
	// initialize and get status
	var ws unix.WaitStatus
	wpid, err := unix.Wait4(-1*tracer.Pgid, &ws, unix.WALL, nil)
	if err != nil {
		util.Warn("wait4: " + err.Error())
		tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return true
	}

	if wpid == -1 {
		util.Warn("wpid = -1")
		tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: "wpid = -1"}
		return true
	}

	// check if segfault, or other stuff
	// http://people.cs.pitt.edu/~alanjawi/cs449/code/shell/UnixSignals.htm

	if isStopSignal(ws.StopSignal()) {
		tracer.Session.ExitCode = int(ws.StopSignal())
		// this object will be filled in by judge channel
		tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_RTE, ResultInfo: "",}
		tracer.Kill()
		return true
	}

	// if process has already exited, leave
	if ws.Exited() {
		tracer.Session.ExitCode = int(ws.StopSignal())
		return true
	}
	return false
}

func (tracer *PTracer) Syscall() bool {
	err := unix.PtraceSyscall(tracer.Pid, 0)
	if err != nil {
		util.Warn("ptracesyscall: " + err.Error())
		tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return true
	}
	return false
}

func (tracer *PTracer) Trace() {
	var err error

	tracer.Pgid, err = unix.Getpgid(tracer.Pid)
	if err != nil {
		util.Warn("getpgid: " + err.Error())
		tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	tracer.Wait4()

	err = unix.PtraceSetOptions(tracer.Pid, unix.PTRACE_O_EXITKILL|unix.PTRACE_O_TRACESYSGOOD|unix.PTRACE_O_TRACEEXIT|unix.PTRACE_O_TRACECLONE|unix.PTRACE_O_TRACEFORK|unix.PTRACE_O_TRACEVFORK)
	if err != nil {
		util.Warn("ptracesetoptions: " + err.Error())
		tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	for { // scan through each syscall
		if tracer.Syscall() {
			return
		}
		if tracer.Wait4() {
			return
		}

		// get system call
		pregs := unix.PtraceRegs{} // user regs struct
		err = unix.PtraceGetRegs(tracer.Pid, &pregs)
		if err != nil {
			util.Warn("ptracegetregs: " + err.Error())
			tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		// map syscall to nothing if syscall is blockedCall
		blockedCall := blockRestrictedCalls(&pregs, tracer.Pid)

		// run system call
		if tracer.Syscall() {
			return
		}
		if tracer.Wait4() {
			return
		}
		if blockedCall {
			pregs.Rax = uint64(math.Inf(0))
		}
	}

}

func (tracer *PTracer) Kill() {
	var ws unix.WaitStatus
	unix.Kill(tracer.Pid, unix.SIGKILL)
	_, err := unix.Wait4(tracer.Pid, &ws, 0, nil)
	for err == unix.EINTR {
		_, err = unix.Wait4(tracer.Pid, &ws, 0, nil)
	}
}