package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"math"
)

// could possibly switch to seccomp instead of ptrace

func sandboxWait4(pgid int, done chan CaseReturn) bool {
	// initialize and get status
	var ws unix.WaitStatus
	wpid, err := unix.Wait4(-1*pgid, &ws, unix.WALL, nil)
	if err != nil {
		util.Warn("wait4: " + err.Error())
		done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return true
	}

	if wpid == -1 {
		util.Warn("wpid = -1")
		done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: "wpid = -1"}
		return true
	}

	// if process has already exited, leave
	if ws.Exited() {
		return true
	}
	return false
}

// do sandboxing of application using ptrace

func sandboxProcess(pid *int, done chan CaseReturn) {

	pgid, err := unix.Getpgid(*pid)
	if err != nil {
		util.Warn("getpgid: " + err.Error())
		done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	err = unix.PtraceSetOptions(*pid, unix.PTRACE_O_EXITKILL | unix.PTRACE_O_TRACESYSGOOD | unix.PTRACE_O_TRACEEXIT | unix.PTRACE_O_TRACECLONE | unix.PTRACE_O_TRACEFORK | unix.PTRACE_O_TRACEVFORK)
	if err != nil {
		util.Warn("ptracesetoptions: " + err.Error())
		done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	for { // scan through each syscall
		err := unix.PtraceSyscall(*pid, 0)
		if err != nil {
			util.Warn("ptracesyscall1: " + err.Error())
			done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		if sandboxWait4(pgid, done) {
			return
		}

		// get system call
		pregs := unix.PtraceRegs{} // user regs struct
		err = unix.PtraceGetRegs(*pid, &pregs)
		if err != nil {
			util.Warn("ptracegetregs: " + err.Error())
			done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		// map syscall to nothing if syscall is blockedCall
		//log.Println(pregs.Orig_rax) // TODO

		blockedCall := blockRestrictedCalls(&pregs, *pid)

		// run system call
		err = unix.PtraceSyscall(*pid, 0)
		if err != nil {
			util.Warn("ptracesyscall2: " + err.Error())
			done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		if sandboxWait4(pgid, done) {
			return
		}

		if blockedCall {
				pregs.Rax = uint64(math.Inf(0))
		}
	}
}

