package shared

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
)

type JudgeSession struct {
	Workspace string
	Code string
	Language string
	OriginalRequest *pb.TestSolutionRequest
}
