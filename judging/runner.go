package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/adapters"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
)

func TestSolution(req *pb.TestSolutionRequest, stream *pb.SmoothieRunnerAPI_TestSolutionServer, res chan shared.JudgeStatus) {
	runCommand, err := adapters.CompileAndGetRunCommand(req.GetSolution().GetLanguage(), req.GetSolution().GetCode())
	if err != nil {
		res <- shared.JudgeStatus{Err: err}
		return
	}

	for _, pb.ProblemBatch := range req.Solution.Problem.TestBatches {

	}
}