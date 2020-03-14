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
	Cur uint64
	Max uint64
}

type RunnerResult struct {
	Status int
	Error string
}

type RunnerSessionResult struct {
	Status int
	Error string
	TimeUsed time.Duration
	MemoryUsed int64
}

type RunnerSession struct {
	// Channel to stream result back (init)
	ResultChan chan RunnerSessionResult

	// Internal result stream (init)
	InternalResultChan chan RunnerResult

	// Pid of child
	Pid int
	Pgid int

	// Execveat (init)
	ExecFile uintptr
	ExecArgs []string
	ExecEnv []string

	// Whether or not the initial exec was called
	ExecUsed bool

	// File descriptors to set: [newfd]oldfd (init)
	Files map[int]uintptr

	// Folder where file is executed
	Workspace string

	// Resource limits with rlimit
	RLimits []RLimit

	// Timeout, in seconds (init)
	TimeLimit time.Duration

	// Maximum memory, in bytes (init)
	MemoryLimit int64

	// Seccomp profile (init)
	SeccompProfile util.SandboxProfile

	// Exit code
	ExitCode int

	// Start time
	StartTime time.Time
}

func (session *RunnerSession) Start() {

	// init rlimit
	session.RLimits = []RLimit{
		{
			Type: unix.RLIMIT_CPU,
			Cur: uint64(session.TimeLimit.Seconds()),
			Max: uint64(session.TimeLimit.Seconds()),
		},
		{
			Type: unix.RLIMIT_FSIZE,
			Cur: uint64(session.MemoryLimit),
			Max: uint64(session.MemoryLimit),
		},
		{
			Type: unix.RLIMIT_AS, // TODO output limit
			Cur: uint64(session.MemoryLimit),
			Max: uint64(session.MemoryLimit),
		},
	}

	go session.WaitForStatus()

	if shared.SANDBOX {
		go session.Trace()
	} else {
		session.StartTime = time.Now()
		go session.WaitProcState()
	}
}

func (session *RunnerSession) WaitForStatus() {
	res := <-session.InternalResultChan

	if util.IsPidRunning(session.Pid) {
		unix.Kill(session.Pid, syscall.SIGTERM)
		unix.Kill(session.Pid, syscall.SIGKILL) // extra assurance
		var wstatus unix.WaitStatus
		unix.Wait4(session.Pid, &wstatus, unix.WALL|unix.WNOHANG, nil) // collect zombie
	}

	session.ResultChan <- RunnerSessionResult{
		Status:     res.Status,
		Error:      res.Error,
		TimeUsed:   time.Since(session.StartTime),
		MemoryUsed: 0, // TODO
	}
}