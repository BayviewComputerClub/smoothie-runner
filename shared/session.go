package shared

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	"os/exec"
)

type Rlimits struct {
	CpuTime uint64 // seconds
	Fsize uint64 // bytes
	Memory uint64 // bytes
}

type JudgeSession struct {
	Workspace string
	Code string
	Language string
	OriginalRequest *pb.TestSolutionRequest
	CommandFd uintptr
	CommandArgs []*byte
	RunCommand *exec.Cmd
	Limits
}

