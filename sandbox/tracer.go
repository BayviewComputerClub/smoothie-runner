package sandbox

import (
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"
)

// status checker when sandbox is on
func (session *RunnerSession) Trace() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var (
		err    error
		traced = make(map[int]bool)
	)

	defer func() {
		shared.Debug("left tracer")
		if err := recover(); err != nil {
			util.Warn("trace panic recover: " + string(debug.Stack()))
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusISE, Error: string(debug.Stack()),}
		}
	}()

	// get process pgid
	session.Pgid, err = unix.Getpgid(session.Pid)
	if err != nil {
		util.Warn("getpgid: " + err.Error())
		session.InternalResultChan <- RunnerResult{Status: RunnerStatusISE, Error: err.Error()}
		return
	}

	var (
		ws     unix.WaitStatus
		rusage unix.Rusage
	)

	shared.Debug(fmt.Sprintf("wait4: now tracing %v %v", ws.Stopped(), session.Pid))

	// trace each syscall
	for {
		pid, err := unix.Wait4(-session.Pgid, &ws, unix.WALL, &rusage)
		if err != nil {
			util.Warn("wait4 error: " + err.Error())
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusISE, Error: err.Error()}
			return
		}

		if session.ExecUsed {
			if session.traceOnce(pid, ws, rusage, traced) {
				return
			}
		} else {
			if session.traceOncePreExec(pid, ws, rusage) {
				return
			}
		}
	}

}

// check status before sandboxed process starts
func (session *RunnerSession) traceOncePreExec(pid int, ws unix.WaitStatus, rusage unix.Rusage) bool {
	var err error

	switch {
	case ws.Exited(): // after signal or stop
		session.InternalResultChan <- RunnerResult{Status: RunnerStatusISE, Error: "Child process exited before exec"}
		return true
	case ws.Signaled():
		err := unix.PtraceCont(pid, int(ws.Signal()))
		if err != nil {
			util.Warn("ptracecont signaled: " + err.Error())
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusISE, Error: err.Error()}
			return true
		}
	case ws.Stopped():
		switch ws.StopSignal() {
		case unix.SIGXCPU:
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusTLE}
			return true
		case unix.SIGXFSZ:
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusOLE}
			return true
		}

		err = unix.PtraceCont(pid, int(ws.StopSignal()))
		if err != nil {
			util.Warn("ptracecont signaled: " + err.Error())
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusISE, Error: err.Error()}
			return true
		}
	}
	return false
}

// check status after sandboxed process starts
func (session *RunnerSession) traceOnce(pid int, ws unix.WaitStatus, rusage unix.Rusage, traced map[int]bool) bool {
	var err error

	// check tle
	if session.TimeLimit < time.Duration(rusage.Utime.Nano()) {
		shared.Debug("TLE")
		session.InternalResultChan <- RunnerResult{Status: RunnerStatusTLE}
		return true
	}

	// check memory
	// maxrss - KB, memlimit - MB
	if session.MemoryLimit < rusage.Maxrss/1e3 {
		shared.Debug("MLE")
		session.InternalResultChan <- RunnerResult{Status: RunnerStatusMLE}
		return true
	}

	// check process status
	switch {
	case ws.Exited(): // process exit (after signal or stop)
		delete(traced, pid) // remove from traced processes

		shared.Debug("tracer: process exited with " + strconv.Itoa(ws.ExitStatus()))
		session.ExitCode = ws.ExitStatus()
		// send exit status to grader
		session.InternalResultChan <- RunnerResult{Status: RunnerStatusOK}
		return true

	case ws.Signaled():
		delete(traced, pid) // remove from traced processes

		// check exit status
		sig := ws.Signal()
		session.ExitCode = int(ws.Signal())
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

	case ws.Stopped():

		// Set option if the process is newly forked
		if !traced[pid] {
			// set ptrace options
			err = unix.PtraceSetOptions(session.Pid, unix.PTRACE_O_TRACESECCOMP|unix.PTRACE_O_EXITKILL|unix.PTRACE_O_TRACEFORK|unix.PTRACE_O_TRACECLONE|unix.PTRACE_O_TRACEEXEC|unix.PTRACE_O_TRACEVFORK)
			if err != nil {
				util.Warn("ptracesetoptions: " + err.Error())
				session.InternalResultChan <- RunnerResult{Status: RunnerStatusISE, Error: err.Error()}
				return true
			}
			traced[pid] = true
		}

		if ws.StopSignal() == unix.SIGTRAP {
			switch ws.TrapCause() {
			case unix.PTRACE_EVENT_SECCOMP:
				shared.Debug("ptrace trap event_seccomp")
				err := session.handleTrap(pid) // handle syscall
				if err != nil {
					session.InternalResultChan <- RunnerResult{Status: RunnerStatusILL, Error: err.Error()}
					return true
				}
			case unix.PTRACE_EVENT_EXEC:
				session.StartTime = time.Now()
			}
			err = unix.PtraceCont(pid, 0)
		} else {

			switch ws.StopSignal() {
			case unix.SIGXCPU:
				session.InternalResultChan <- RunnerResult{Status: RunnerStatusTLE}
				return true
			case unix.SIGXFSZ:
				session.InternalResultChan <- RunnerResult{Status: RunnerStatusOLE}
				return true
			}

			err = unix.PtraceCont(pid, int(ws.StopSignal()))
		}

		if err != nil {
			util.Warn("ptracecont signaled: " + err.Error())
			session.InternalResultChan <- RunnerResult{Status: RunnerStatusISE, Error: err.Error()}
			return true
		}

	}
	return false
}
