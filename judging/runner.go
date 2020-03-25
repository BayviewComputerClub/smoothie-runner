package judging

import (
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/adapters"
	"github.com/BayviewComputerClub/smoothie-runner/cache"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"github.com/rs/xid"
	"os"
	"path"
	"sync/atomic"
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
		util.Info(fmt.Sprintf("Worker %d has picked up job for %s in %s.", num, job.Req.Problem.ProblemId, job.Req.Solution.Language))

		TestSolution(job.Req, job.Res, job.Cancelled)
		atomic.AddInt64(shared.TasksToBeDone, -1)
		util.Info(fmt.Sprintf("Worker %v has completed job %v in %v", num, job.Req.Problem.ProblemId, job.Req.Solution.Language))
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
		FSizeLimit: 1e9, // TODO output limit
	}

	// resolve full path for workspace
	if len(shared.TESTING_DIR) == 0 || shared.TESTING_DIR[0] == '.' {
		p, err := os.Getwd()
		if err != nil {
			sendISE(err, res)
			return
		}
		session.Workspace = path.Clean(p + "/" + session.Workspace)
	}

	// get test data
	testData, err := cache.GetTestData(session.OriginalRequest.Problem.ProblemId)
	if err != nil {
		sendISE(err, res)
		return
	}
	defer testData.Cleanup() // close streams

	// remove workspace when exit
	if shared.CLEANUP_SESSIONS {
		defer os.RemoveAll(session.Workspace)
	}

	// create session workspace
	err = os.Mkdir(session.Workspace, 0755)
	if err != nil {
		panic(err)
	}

	util.Info("Compiling request in " + session.Language + " for " + session.Workspace + ".")
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
	util.Info("Compile complete for " + session.Workspace + ".")

	var f *os.File
	// get exec command pointers
	f, err = util.GetPtrsFromCmd(session.RunCommand)
	if err != nil {
		util.Warn("getptrsfromcmd: " + err.Error())
		sendISE(err, res)
		return
	}
	defer f.Close() // must close AFTER the testing is finished
	session.CommandFd = f.Fd()

	// loop over test batches and cases
	for _, batch := range testData.Batches {

		shared.Debug(fmt.Sprintf("Batch #%v", batch.BatchNum))

		batchFailed := false
		for _, batchCase := range batch.Cases {
			shared.Debug(fmt.Sprintf("Judging case #%v", batchCase.CaseNum))
			if *cancelled { // exit if cancelled
				return
			}

			// if whole batch had failed, skip
			if !req.TestBatchEvenIfFailed && batchFailed {
				res <- shared.JudgeStatus{
					Err: nil,
					Res: pb.TestSolutionResponse{
						TestCaseResult: &pb.TestCaseResult{
							BatchNumber:          uint64(batchCase.BatchNum),
							CaseNumber:           uint64(batchCase.CaseNum),
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
			result := JudgeCase(uint64(batchCase.BatchNum), uint64(batchCase.CaseNum), &session, res, &batchCase)

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

func JudgeCase(batchNum uint64, caseNum uint64, session *shared.JudgeSession, res chan shared.JudgeStatus, batchCase *cache.CachedTestDataCase) pb.TestCaseResult {
	batchRes := make(chan pb.TestCaseResult)

	// do judging
	gradingSession := GradeSession{
		JudgingSession: session,

		Problem:     session.OriginalRequest.Problem,
		Solution:    session.OriginalRequest.Solution,
		CurrentCase: batchCase,

		BatchNum: 		batchNum,
		CaseNum: 		caseNum,

		Stderr:         "",

		StreamResult:   batchRes,
		StreamDone:     make(chan CaseReturn),

		Command:  session.RunCommand,
		ExecFile: session.CommandFd,

		SeccompProfile: util.SANDBOX_DEFAULT_PROFILE, // TODO
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
