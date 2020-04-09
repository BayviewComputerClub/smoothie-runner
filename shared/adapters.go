package shared

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	"os/exec"
)

type SmoothieAdapter interface {
	GetName() string
	Compile(session *JudgeSession) (*exec.Cmd, error) // return command in the workspace
	JudgeFinished(tcr *pb.TestCaseResult)
}
