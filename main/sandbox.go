package main

import (
	"golang.org/x/sys/unix"
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
		warn(err.Error())
		done <- CaseReturn{Result: OUTCOME_RTE}
		return
	}

	if wpid == -1 {
		warn("wpid = -1")
		done <- CaseReturn{Result: OUTCOME_RTE}
		return
	}

	// if process has already exited, leave
	if ws.Exited() {
		return
	}
}


func sandboxProcess(pid int, done chan CaseReturn) {
	runtime.LockOSThread() // https://github.com/golang/go/issues/7699

	for { // scan through each syscall
		err := unix.PtraceSyscall(pid, 0)
		if err != nil {
			warn(err.Error())
			done <- CaseReturn{Result: OUTCOME_RTE}
			return
		}

		sandboxWait4(pid, done)

		// get system call
		pregs := unix.PtraceRegs{}
		err = unix.PtraceGetRegs(pid, &pregs)
		if err != nil {
			warn(err.Error())
			done <- CaseReturn{Result: OUTCOME_RTE}
			return
		}

		blocked := false
		// map syscall to nothing if syscall is blocked
		if blocked = isBlockedSyscall(pregs.Orig_rax); blocked {
			pregs.Orig_rax = -1
			err = unix.PtraceSetRegs(pid, &pregs)
		}

		// run system call
		err = unix.PtraceSyscall(pid, 0)
		if err != nil {
			warn(err.Error())
			done <- CaseReturn{Result: OUTCOME_RTE}
			return
		}

		sandboxWait4(pid, done)

		if blocked {
			pregs.Rax = -0x1
		}
	}
}
