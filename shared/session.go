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
	CommandArgs []*byte
	RunCommand *exec.Cmd
}