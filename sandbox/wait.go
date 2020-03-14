package sandbox

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"golang.org/x/sys/unix"
	"time"
)

// status checker when sandbox is off
func (session *RunnerSession) WaitProcState() {
	var (
		wstatus unix.WaitStatus
		rusage  unix.Rusage
	)

	for {
		// wait for process change state
		_, err := unix.Wait4(session.Pid, &wstatus, 0, &rusage)
		println("WAIT4: ", wstatus) // TODO
		if err != nil {
			shared.Debug("wait4: " + err.Error())
			return err
		}

		// check tle
		if session.TimeLimit < time.Duration(rusage.Utime.Nano()) {
			shared.Debug("TLE")
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusTLE}
			return
		}

		// check memory
		// maxrss - KB, memlimit - MB
		if session.MemoryLimit < rusage.Maxrss/1e3 {
			shared.Debug("MLE")
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusMLE}
			return
		}

		switch {
		case wstatus.Exited(): // program exit
			session.ExitCode = wstatus.ExitStatus()
			// send exit status to grader
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusOK}
			return
		case wstatus.Signaled():
			sig := wstatus.Signal()
			session.ExitCode = int(wstatus.Signal())
			switch sig {
			case unix.SIGXCPU, unix.SIGKILL:
				session.InternalResultChan <- RunnerResult{Status: RunnerStatusTLE}
			case unix.SIGXFSZ:
				session.InternalResultChan <- RunnerResult{Status: RunnerStatusOLE}
			case unix.SIGSYS:
				session.InternalResultChan <- RunnerResult{Status: RunnerStatusIllegalSyscall}
			default:
				session.InternalResultChan <- RunnerResult{Status: RunnerStatusRuntimeError}
			}
			return
		}

	}
}