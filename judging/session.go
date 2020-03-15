package judging

import (
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/cache"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	"github.com/BayviewComputerClub/smoothie-runner/sandbox"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type CaseReturn struct {
	Result     string
	ResultInfo string
}

type GradeSession struct {
	JudgingSession *shared.JudgeSession

	// case information
	Problem     *pb.Problem
	Solution    *pb.Solution
	CurrentCase *cache.CachedTestDataCase
	BatchNum    uint64
	CaseNum     uint64

	// streams during judging
	OutputStream *os.File
	ErrorStream  *os.File
	InputStream  *os.File

	Stderr string // error dumped here

	StreamResult chan pb.TestCaseResult // return result to runner
	StreamDone   chan CaseReturn        // end batch case with verdict
	DoneSent     bool                   // whether or not the done channel was used

	// for forkexec
	Command  *exec.Cmd
	ExecFile uintptr

	// seccomp & ptrace
	SeccompProfile util.SandboxProfile

	// sandbox session
	RunnerSession sandbox.RunnerSession
	RunnerResult  sandbox.RunnerSessionResult
}


/*
 * Run to start judging
 */

func (session *GradeSession) StartJudging() {
	runtime.LockOSThread() // https://github.com/golang/go/issues/7699
	defer runtime.UnlockOSThread()

	// init pipes
	session.InitIOFiles()
	defer session.CloseStreams()

	// runner session
	session.RunnerSession = sandbox.RunnerSession{
		ResultChan:         make(chan sandbox.RunnerSessionResult),
		InternalResultChan: make(chan sandbox.RunnerResult),
		ExecFile:           session.ExecFile,
		ExecArgs:           session.Command.Args,
		ExecEnv:            session.Command.Env,
		ExecUsed:           false,
		Files:              make(map[int]uintptr),
		Workspace:          session.Command.Dir,
		TimeLimit:          time.Duration(session.Problem.TimeLimit) * time.Second,
		MemoryLimit:        int64(session.Problem.MemLimit)*1000000,
		SeccompProfile:     session.SeccompProfile,
	}

	// initialize file descriptors
	session.RunnerSession.Files[0] = session.InputStream.Fd()
	session.RunnerSession.Files[1] = session.OutputStream.Fd()
	session.RunnerSession.Files[2] = session.ErrorStream.Fd()

	go session.RunnerSession.Start()
	go session.WaitVerdict()
	go session.Timeout()

	// wait until child processes finish
	session.RunnerResult = <-session.RunnerSession.ResultChan

	// read stderr after process runs
	stderr, _ := ioutil.ReadFile(session.ErrorStream.Name())
	session.Stderr = string(stderr)

	if session.RunnerResult.Status == sandbox.RunnerStatusOK && session.Stderr == "" {
		// run the grading session if the runner ran successfully
		// it will send AC or WA
		// must run on this goroutine (streams need to remain open for grading)
		StartGrader(session)
	} else {
		// return status otherwise
		ret := CaseReturn{
			Result:     "",
			ResultInfo: session.RunnerResult.Error,
		}

		// deal with any output in stderr
		if session.Stderr != "" {
			ret = CaseReturn{
				Result:     shared.OUTCOME_RTE,
				ResultInfo: session.Stderr,
			}
		}

		switch session.RunnerResult.Status {
		case sandbox.RunnerStatusISE:
			ret.Result = shared.OUTCOME_ISE
		case sandbox.RunnerStatusILL:
			ret.Result = shared.OUTCOME_ILL
		case sandbox.RunnerStatusMLE:
			ret.Result = shared.OUTCOME_MLE
		case sandbox.RunnerStatusTLE:
			ret.Result = shared.OUTCOME_TLE
		case sandbox.RunnerStatusOLE:
			ret.Result = shared.OUTCOME_OLE
		case sandbox.RunnerStatusRTE:
			ret.Result = shared.OUTCOME_RTE
		}
		session.StreamDone <- ret
	}
}

func (session *GradeSession) Timeout() {
	time.Sleep(time.Duration(session.Problem.TimeLimit)*time.Second*2 + 10*time.Second)
	if !session.DoneSent {
		session.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: "timeout"}
	}
}

/*
 * Wait for result from other goroutines
 */

func (session *GradeSession) WaitVerdict() {
	defer session.CloseStreams()

	// wait for judging to finish
	response := <-session.StreamDone
	session.DoneSent = true

	// send result back to runner
	if response.Result != shared.OUTCOME_AC && session.Stderr != "" { //  if the outcome was wrong answer but there was an error
		session.StreamResult <- pb.TestCaseResult{
			Result:      shared.OUTCOME_RTE,
			ResultInfo:  session.Stderr,
			Time:        session.RunnerResult.TimeUsed.Seconds(),
			MemUsage:    float64(session.RunnerResult.MemoryUsed) / 1000,
			BatchNumber: session.BatchNum,
			CaseNumber:  session.CaseNum,
		}
	} else if response.Result == shared.OUTCOME_RTE { // if the program did not exit successfully
		session.StreamResult <- pb.TestCaseResult{
			Result:      shared.OUTCOME_RTE,
			ResultInfo:  fmt.Sprintf("Exit code: %v: %v", session.RunnerSession.ExitCode, session.Stderr),
			Time:        session.RunnerResult.TimeUsed.Seconds(),
			MemUsage:    float64(session.RunnerResult.MemoryUsed) / 1000,
			BatchNumber: session.BatchNum,
			CaseNumber:  session.CaseNum,
		}
	} else { // if the program exited successfully
		session.StreamResult <- pb.TestCaseResult{
			Result:      response.Result,
			ResultInfo:  response.ResultInfo,
			Time:        session.RunnerResult.TimeUsed.Seconds(),
			MemUsage:    float64(session.RunnerResult.MemoryUsed) / 1000,
			BatchNumber: session.BatchNum,
			CaseNumber:  session.CaseNum,
		}
	}
}
