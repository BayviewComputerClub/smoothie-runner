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
			session.InternalResultChan <- RunnerResult{
				Status: RunnerStatusISE,
				Error:  err.Error(),
			}
			return
		}

		// check tle
		if session.TimeLimit < time.Duration(rusage.Utime.Nano()) {
			shared.Debug("TLE")
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusTLE}
			return
		}

		// check memory
		// maxrss - KB, memlimit - MB
		session.MemoryUsed = rusage.Maxrss
		if session.MemoryLimit > 0 && session.MemoryLimit*1024 < uint64(rusage.Maxrss) {
			shared.Debug("MLE")
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusMLE}
			return
		}

		// check program status
		switch {
		case wstatus.Exited(): // normal program exit
			session.ExitCode = wstatus.ExitStatus()
			// send exit status to grader
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusOK}
			return
		case wstatus.Signaled(): // not normal program exit
			sig := wstatus.Signal()
			session.ExitCode = int(wstatus.Signal())
			switch sig {
			case unix.SIGXCPU, unix.SIGKILL:
				session.InternalResultChan <- RunnerResult{Status: RunnerStatusTLE}
			case unix.SIGXFSZ:
				session.InternalResultChan <- RunnerResult{Status: RunnerStatusOLE}
			case unix.SIGSYS:
				session.InternalResultChan <- RunnerResult{Status: RunnerStatusILL}
			default:
				session.InternalResultChan <- RunnerResult{Status: RunnerStatusRTE}
			}
			return
		}

	}
}