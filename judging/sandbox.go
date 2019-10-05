package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"math"
	"runtime"
	"syscall"
)

// could possibly switch to seccomp instead of ptrace

// https://filippo.io/linux-syscall-table/
func isBlockedSyscall(call uint64) bool {
	allowedCalls := [4]uint64{unix.SYS_READ, unix.SYS_WRITE, unix.SYS_EXIT, unix.SYS_RT_SIGRETURN}

	found := false
	for _, a := range allowedCalls {
		if a == call {
			found = true
		}
	}
	return !found
}

func sandboxWait4(pgid int, done chan CaseReturn) {
	// initialize and get status
	var ws unix.WaitStatus
	wpid, err := unix.Wait4(-1*pgid, &ws, syscall.WALL, nil)
	if err != nil {
		util.Warn("wait4: " + err.Error())
		done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	if wpid == -1 {
		util.Warn("wpid = -1")
		done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: "wpid = -1"}
		return
	}

	// if process has already exited, leave
	if ws.Exited() {
		return
	}
}


func sandboxProcess(pid int, done chan CaseReturn) {
	runtime.LockOSThread() // https://github.com/golang/go/issues/7699
	defer runtime.LockOSThread()

	//log.Printf("%d\n", pid) // TODO

	err := syscall.PtraceSetOptions(pid, syscall.PTRACE_O_TRACESYSGOOD)
	if err != nil {
		util.Warn("ptracesetoptions: " + err.Error())
		done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	for { // scan through each syscall
		err := syscall.PtraceSyscall(pid, 0)
		if err != nil {
			util.Warn("ptracesyscall: " + err.Error())
			done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		sandboxWait4(pid, done)

		// get system call
		pregs := unix.PtraceRegs{}
		err = unix.PtraceGetRegs(pid, &pregs)
		if err != nil {
			util.Warn("ptracegetregs: " + err.Error())
			done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		blocked := false
		// map syscall to nothing if syscall is blocked
		if blocked = isBlockedSyscall(pregs.Orig_rax); blocked {
			pregs.Orig_rax = uint64(math.Inf(0)) // TODO
			err = unix.PtraceSetRegs(pid, &pregs)
		}

		// run system call
		err = unix.PtraceSyscall(pid, 0)
		if err != nil {
			util.Warn("ptracesyscall: " + err.Error())
			done <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		sandboxWait4(pid, done)

		if blocked {
			pregs.Rax = uint64(math.Inf(0)) // TODO
		}
	}
}
