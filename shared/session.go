package shared

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	"os/exec"
)

type JudgeSession struct {
	Workspace string
	Code string
	Language string
	OriginalRequest *pb.TestSolutionRequest
	CommandFd uintptr
	RunCommand *exec.Cmd

	// limits set by adapters
	FSizeLimit int64
	NProcLimit int64
}