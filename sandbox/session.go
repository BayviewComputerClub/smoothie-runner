package sandbox

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"time"
)

const (
	RunnerStatusOK = iota // successful run

	RunnerStatusTLE // time limit exceeded
	RunnerStatusMLE // memory limit exceeded
	RunnerStatusOLE // output limit exceeded

	RunnerStatusIllegalSyscall
	RunnerStatusRuntimeError
	RunnerStatusInternalServerError
)

type Rlimit struct {
	Type int
	Cur uint64
	Max uint64
}

type RunnerResult struct {
	Status int
	Error string
}

type RunnerSession struct {
	// Channel to stream result back (init)
	ResultChan chan RunnerResult

	// Internal result stream
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
	RLimits []Rlimit

	// Timeout, in seconds (init)
	TimeLimit time.Duration

	// Maximum memory, in bytes (init)
	MemoryLimit int64

	// Seccomp profile (init)
	SeccompProfile util.SandboxProfile

	// Exit code
	ExitCode int
}

func (session *RunnerSession) Start() {

	// init rlimit
	session.RLimits = []Rlimit {
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

	} else {
		go session.WaitProcState()
	}
}

func (session *RunnerSession) WaitForStatus() {
	res := <-session.InternalResultChan



	session.ResultChan <- res
}