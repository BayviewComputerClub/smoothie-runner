package sandbox

import "github.com/BayviewComputerClub/smoothie-runner/util"

type Rlimit struct {
	Type int
	Cur uint64
	Max uint64
}

type RunnerResult struct {

}

type RunnerSession struct {
	// Channel to stream result back
	ResultChan chan RunnerResult

	// Pid of child
	Pid int
	Pgid int

	// Execveat
	ExecFile uintptr
	ExecArgs []string
	ExecEnv []string

	// Whether or not the initial exec was called
	ExecUsed bool

	// File descriptors to set: [newfd]oldfd
	Files map[int]uintptr

	// Folder where file is executed
	Workspace string

	// Resource limits with rlimit
	RLimits []Rlimit

	// Seccomp profile
	SeccompProfile util.SandboxProfile
}

func (session *RunnerSession) Start() {

}