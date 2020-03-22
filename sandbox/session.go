package sandbox

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"syscall"
	"time"
)

const (
	RunnerStatusOK = iota // successful run

	RunnerStatusTLE // time limit exceeded
	RunnerStatusMLE // memory limit exceeded
	RunnerStatusOLE // output limit exceeded

	RunnerStatusILL // illegal syscall
	RunnerStatusRTE // runtime error
	RunnerStatusISE // internal server error
)

type RLimit struct {
	Type int
	Cur  uint64
	Max  uint64
}

type RunnerResult struct {
	Status int
	Error  string
}

type RunnerSessionResult struct {
	Status     int
	ExitCode   int
	Error      string
	TimeUsed   time.Duration
	MemoryUsed int64
}

type RunnerSession struct {
	// Channel to stream result back (init)
	ResultChan chan RunnerSessionResult

	// Internal result stream (init)
	InternalResultChan chan RunnerResult

	// Pid of child
	Pid  int
	Pgid int

	// Execveat (init)
	ExecFile uintptr
	ExecArgs []string
	ExecEnv  []string

	// Whether or not the initial exec was called
	ExecUsed bool

	// Whether or not the process has exited
	ProcExited bool

	// File descriptors to set: [newfd]oldfd (init)
	Files map[int]uintptr

	// Folder where file is executed
	Workspace string

	// Resource limits with rlimit
	RLimits []RLimit

	// Hard timeout, includes time spent preparing sandbox, done by goroutine -> kill (init)
	HardTimeout time.Duration

	// Soft timeout, done by process (init)
	TimeLimit time.Duration

	// Maximum memory, in bytes (init)
	MemoryLimit uint64

	// Maximum size of new files a process can create (init)
	FSizeLimit int64

	// Maximum number of processes that can be created (init)
	NProcLimit int64

	// Whether or not the process should be sandboxed with seccomp + ptrace (init)
	SandboxWithSeccomp bool

	// Seccomp profile (init)
	SeccompProfile util.SandboxProfile

	// Exit code
	ExitCode int

	// Start time
	StartTime time.Time

	// Max memory allocated at a point (kb)
	MemoryUsed int64
}

// enforce a hard timeout
func (session *RunnerSession) Timeout() {
	time.Sleep(session.HardTimeout)
	if !session.ProcExited {
		session.InternalResultChan <- RunnerResult{
			Status: RunnerStatusTLE,
			Error:  "hard timeout",
		}
	}
}

func (session *RunnerSession) Start() {
	// start hard timeout
	go session.Timeout()

	// configure rlimits
	session.InitRLimits()

	// listen on channel
	go session.WaitForStatus()

	// start child process
	err := session.ForkExec()
	if err != nil {
		session.InternalResultChan <- RunnerResult{
			Status: RunnerStatusISE,
			Error:  err.Error(),
		}
		return
	}

	// check for process state change
	if session.SandboxWithSeccomp && shared.SANDBOX {
		go session.Trace()
	} else {
		session.StartTime = time.Now()
		go session.WaitProcState()
	}
}

func (session *RunnerSession) Kill() {
	unix.Kill(session.Pid, syscall.SIGTERM)
	unix.Kill(session.Pid, unix.SIGKILL) // heh why
	var wstatus unix.WaitStatus
	unix.Wait4(session.Pid, &wstatus, unix.WALL|unix.WNOHANG, nil) // collect zombie
}

func (session *RunnerSession) WaitForStatus() {
	// receives when child is finished
	res := <-session.InternalResultChan

	if util.IsPidRunning(session.Pid) {
		session.Kill()
	}

	session.ProcExited = true

	// send result to result channel
	session.ResultChan <- RunnerSessionResult{
		Status:     res.Status,
		ExitCode:   session.ExitCode,
		Error:      res.Error,
		TimeUsed:   time.Since(session.StartTime),
		MemoryUsed: session.MemoryUsed,
	}
}
