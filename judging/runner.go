package judging

import (
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/adapters"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"github.com/rs/xid"
	"os"
	"syscall"
	"unsafe"
)

var (
	workQueue = make(chan JudgeJob)
)

type JudgeJob struct {
	Req *pb.TestSolutionRequest
	Res chan shared.JudgeStatus
	Cancelled *bool
}

// add a judging job to queue
func AddToQueue(job JudgeJob) {
	workQueue <- job
}

// start one worker that will use a thread
func StartQueueWorker(num int) {
	for {
		job := <-workQueue
		util.Info(fmt.Sprintf("Worker %d has picked up job for %d in %s.", num, job.Req.Solution.Problem.ProblemID, job.Req.Solution.Language))
		TestSolution(job.Req, job.Res, job.Cancelled)
	}
}

func emptyTcr() *pb.TestCaseResult {
	return &pb.TestCaseResult{
		BatchNumber: 0,
		CaseNumber:  0,
		Result:      "",
		ResultInfo:  "",
		Time:        0,
		MemUsage:    0,
	}
}

func sendISE(err error, res chan shared.JudgeStatus) {
	res <- shared.JudgeStatus{
		Err: err,
		Res: pb.TestSolutionResponse{
			TestCaseResult:   emptyTcr(),
			CompletedTesting: true,
			CompileError:     shared.OUTCOME_ISE,
		},
	}
}

func TestSolution(req *pb.TestSolutionRequest, res chan shared.JudgeStatus, cancelled *bool) {
	if *cancelled {
		return
	}

	// create judgesession object
	session := shared.JudgeSession{
		Workspace:       shared.TESTING_DIR + "/" + xid.New().String(),
		Code:            req.Solution.Code,
		Language:        req.Solution.Language,
		OriginalRequest: req,
	}

	// remove workspace when exit
	defer os.RemoveAll(session.Workspace)

	// create session workspace
	err := os.Mkdir(session.Workspace, 0755)
	if err != nil {
		panic(err)
	}

	// attempt to compile user submitted code
	session.RunCommand, err = adapters.CompileAndGetRunCommand(session)
	if err != nil {
		// send compile error back
		res <- shared.JudgeStatus{
			Err: err,
			Res: pb.TestSolutionResponse{
				TestCaseResult:   emptyTcr(),
				CompletedTesting: true,
				CompileError:     shared.OUTCOME_CE + ": " + err.Error(),
			},
		}
		return
	}

	// get exec command pointers
	f, err := os.Open(session.RunCommand.Path)
	if err != nil {
		util.Warn("commandfileopen: " + err.Error())
		sendISE(err, res)
		return
	}
	defer f.Close() // must close AFTER the testing is finished
	session.CommandFd = f.Fd()

	// command args ptr
	session.CommandArgs, err = syscall.SlicePtrFromStrings(append(session.RunCommand.Args, "NULL"))
	if err != nil {
		util.Warn("commandbyteparse: " + err.Error())
		sendISE(err, res)
		return
	}

	// loop over test batches and cases
	for _, batch := range req.Solution.Problem.TestBatches {
		for _, batchCase := range batch.Cases {
			if *cancelled { // exit if cancelled
				return
			}

			// judge the case and get the result
			result := JudgeCase(&session, res, batchCase)

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
			TestCaseResult:   emptyTcr(),
			CompletedTesting: true,
			CompileError:     "",
		},
	}

}

func JudgeCase(session *shared.JudgeSession, res chan shared.JudgeStatus, batchCase *pb.ProblemBatchCase) pb.TestCaseResult {
	batchRes := make(chan pb.TestCaseResult)

	// do judging
	gradingSession := GradeSession{
		JudgingSession: session,
		Problem:        session.OriginalRequest.Solution.Problem,
		Solution:       session.OriginalRequest.Solution,
		CurrentBatch:   batchCase,
		Stderr:         "",
		ExitCode:       0,
		StreamResult:   batchRes,
		StreamDone:     make(chan CaseReturn),
		Command:        session.RunCommand,
		ExecCommand:    session.CommandFd,
		ExecArgs:       uintptr(unsafe.Pointer(&session.CommandArgs)),
	}
	go gradingSession.StartJudging()

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

	return result
}
