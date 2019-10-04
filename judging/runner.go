package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/adapters"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
)

func TestSolution(req *pb.TestSolutionRequest, stream pb.SmoothieRunnerAPI_TestSolutionServer, res chan shared.JudgeStatus) {

	// attempt to compile user submitted code
	runCommand, err := adapters.CompileAndGetRunCommand(req.GetSolution().GetLanguage(), req.GetSolution().GetCode())
	if err != nil {
		res <- shared.JudgeStatus{Err: err}

		// send compile error back
		err = stream.Send(&pb.TestSolutionResponse{
			TestCaseResult:   nil,
			CompletedTesting: false,
			CompileError:     err.Error(),
		})
		return
	}

	// loop over test batches and cases
	for _, batch := range req.Solution.Problem.TestBatches {
		for _, batchCase := range batch.Cases {
			batchRes := make(chan pb.TestCaseResult)
			judgeCase(runCommand, batchCase, batchRes)

			// send case result
			result := <-batchRes
			err = stream.Send(&pb.TestSolutionResponse{
				TestCaseResult:   &result,
				CompletedTesting: false,
				CompileError:     "",
			})

			// exit if whole batch fails
			if result.Result != shared.OUTCOME_AC && !req.TestBatchEvenIfFailed {
				break
			}
		}
	}


}