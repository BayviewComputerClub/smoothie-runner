package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/adapters"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
)

func TestSolution(req *pb.TestSolutionRequest, res chan shared.JudgeStatus, cancelled *bool) {

	// attempt to compile user submitted code
	runCommand, err := adapters.CompileAndGetRunCommand(req.GetSolution().GetLanguage(), req.GetSolution().GetCode())
	if err != nil {

		// send compile error back
		res <- shared.JudgeStatus{
			Err: err,
			Res: pb.TestSolutionResponse{
				TestCaseResult:   nil,
				CompletedTesting: true,
				CompileError:     shared.OUTCOME_CE + ": " + err.Error(),
			},
		}

		return
	}

	// loop over test batches and cases
	for _, batch := range req.Solution.Problem.TestBatches {
		for _, batchCase := range batch.Cases {
			if *cancelled { // exit if cancelled
				return
			}

			batchRes := make(chan pb.TestCaseResult)
			judgeCase(runCommand, batchCase, batchRes)

			// wait for case result
			result := <-batchRes

			// send result
			res <- shared.JudgeStatus{
				Err: nil,
				Res: pb.TestSolutionResponse{
					TestCaseResult:   &result,
					CompletedTesting: false,
					CompileError:     "",
				},
			}

			// exit if whole batch fails
			if result.Result != shared.OUTCOME_AC && !req.TestBatchEvenIfFailed {
				break
			}
		}
	}

	// return successful judging
	res <- shared.JudgeStatus{
		Err: nil,
		Res: pb.TestSolutionResponse{
			TestCaseResult:   nil,
			CompletedTesting: true,
			CompileError:     "",
		},
	}

}