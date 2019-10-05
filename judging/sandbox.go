package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"log"
	"math"
	"syscall"
)

// could possibly switch to seccomp instead of ptrace

// https://filippo.io/linux-syscall-table/
func isBlockedSyscall(call uint64) bool {
	return false
	//allowedCalls := [4]uint64{4, 20, 231, 60}
	/*allowedCalls := [4]uint64{unix.SYS_READ, unix.SYS_WRITE, unix.SYS_EXIT, unix.SYS_RT_SIGRETURN}

	found := false
	for _, a := range allowedCalls {
		if a == call {
			found = true
		}
	}
	return !found*/
}

func sandboxWait4(pgid int, done chan CaseReturn) bool {
	// initialize and get status
	var ws unix.WaitStatus
	wpid, err := unix.Wait4(-1*pgid, &ws, syscall.WALL, nil)
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


func sandboxProcess(pid *int, done chan CaseReturn) {

	pgid, err := unix.Getpgid(*pid)
	if err != nil {
		util.Warn("getpgid: " + err.Error())
		done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	log.Printf("%d %d\n", *pid, pgid) // TODO

	err = unix.PtraceSetOptions(*pid, unix.PTRACE_O_EXITKILL)
	if err != nil {
		util.Warn("ptracesetoptions: " + err.Error())
		done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	//defer println("I LEFT") // TODO
	for { // scan through each syscall
		//println("WTFU") // TODO
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
		pregs := unix.PtraceRegs{}
		err = unix.PtraceGetRegs(*pid, &pregs)
		if err != nil {
			util.Warn("ptracegetregs: " + err.Error())
			done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		blocked := false
		// map syscall to nothing if syscall is blocked
		//println(strconv.Itoa(int(pregs.Orig_rax)) + " " + strconv.Itoa(*pid)) // TODO
		if blocked = isBlockedSyscall(pregs.Orig_rax); blocked {
			pregs.Orig_rax = uint64(math.Inf(0)) // TODO
			err = unix.PtraceSetRegs(*pid, &pregs)
			if err != nil {
				util.Warn("ptracesetregs: " + err.Error())
			}
		}

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

		if blocked {
			pregs.Rax = uint64(math.Inf(0)) // TODO
		}
	}
}

