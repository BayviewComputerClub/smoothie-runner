package judging

import (
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/adapters"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"github.com/rs/xid"
	"os"
	"sync/atomic"
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
	atomic.AddInt64(shared.TasksToBeDone, 1)
	atomic.AddInt64(shared.TasksInQueue, 1)
	workQueue <- job
}

// start one worker that will use a thread
func StartQueueWorker(num int) {
	for {
		job := <-workQueue
		atomic.AddInt64(shared.TasksInQueue, -1)
		util.Info(fmt.Sprintf("Worker %d has picked up job for %s in %s.", num, job.Req.Problem.ProblemID, job.Req.Solution.Language))

		TestSolution(job.Req, job.Res, job.Cancelled)
		atomic.AddInt64(shared.TasksToBeDone, -1)
		util.Info(fmt.Sprintf("Worker %v has completed job %v in %v", num, job.Req.Problem.ProblemID, job.Req.Solution.Language))
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
		Limit:			 shared.Rlimits{
			CpuTime: uint64(req.Problem.TimeLimit), // only second precision :/
			Fsize:   1e9, // 1e9 bytes -> 1 gigabyte
			Memory:  uint64(req.Problem.MemLimit*1e6), // MB -> bytes
		},
	}

	// remove workspace when exit
	if shared.CLEANUP_SESSIONS {
		defer os.RemoveAll(session.Workspace)
	}

	// create session workspace
	err := os.Mkdir(session.Workspace, 0755)
	if err != nil {
		panic(err)
	}

	// attempt to compile user submitted code
	session.RunCommand, err = adapters.CompileAndGetRunCommand(&session)
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
	for i, batch := range req.Problem.TestBatches {

		shared.Debug(fmt.Sprintf("Batch #%v", i))

		batchFailed := false
		for j, batchCase := range batch.Cases {
			shared.Debug(fmt.Sprintf("Judging case #%v", j))
			if *cancelled { // exit if cancelled
				return
			}

			// if whole batch had failed, skip
			if !req.TestBatchEvenIfFailed && batchFailed {
				res <- shared.JudgeStatus{
					Err: nil,
					Res: pb.TestSolutionResponse{
						TestCaseResult: &pb.TestCaseResult{
							BatchNumber:          uint64(i),
							CaseNumber:           uint64(j),
							Result:               shared.OUTCOME_SKIP,
							ResultInfo:           "",
						},
						CompletedTesting:     false,
						CompileError:         "",
					},
				}
				continue
			}

			// judge the case and get the result
			result := JudgeCase(uint64(i), uint64(j), &session, res, batchCase)

			if result.Result != shared.OUTCOME_AC {
				batchFailed = true
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

func JudgeCase(batchNum uint64, caseNum uint64, session *shared.JudgeSession, res chan shared.JudgeStatus, batchCase *pb.ProblemBatchCase) pb.TestCaseResult {
	batchRes := make(chan pb.TestCaseResult)

	// do judging
	gradingSession := GradeSession{
		CaseNum: 		caseNum,
		BatchNum: 		batchNum,
		JudgingSession: session,
		Problem:        session.OriginalRequest.Problem,
		Solution:       session.OriginalRequest.Solution,
		Limit: 			&session.Limit,
		CurrentBatch:   batchCase,
		Stderr:         "",
		ExitCode:       0,
		StreamResult:   batchRes,
		StreamDone:     make(chan CaseReturn),
		StreamProcEnd:  make(chan bool),
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
