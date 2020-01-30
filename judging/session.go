package judging

import (
	"bufio"
	"fmt"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

type CaseReturn struct {
	Result string
	ResultInfo string
}

type GradeSession struct {
	JudgingSession *shared.JudgeSession

	Problem *pb.Problem
	Solution *pb.Solution
	CurrentBatch *pb.ProblemBatchCase

	BatchNum uint64
	CaseNum uint64

	// streams during judging
	ErrorBuffer  *os.File
	OutputStream *os.File
	ErrorStream  *os.File
	InputStream  *os.File

	// per judge session
	Stderr   string // error dumped here
	ExitCode int

	StreamResult chan pb.TestCaseResult // return result to runner
	StreamDone chan CaseReturn // end batch case with verdict
	StreamDoneUsed bool
	StartTime time.Time

	Command *exec.Cmd
	ExecCommand uintptr
	ExecArgs uintptr

	Pid int
}

/*
 * Initialize all streams before program starts
 */

func (session *GradeSession) InitStreams() {

	session.InitIOFiles()

	var err error
	// stderr buffer
	session.ErrorBuffer, session.ErrorStream, err = os.Pipe()
	if err != nil {
		util.Warn("stderrpipeinit: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}
}

/*
 * Initializes input stream
 */

func (session *GradeSession) InitIOFiles() {
	name := strconv.FormatInt(time.Now().Unix(), 10)

	outputFileLoc := session.JudgingSession.Workspace + "/" + name + ".out"
	inputFileLoc := session.JudgingSession.Workspace + "/" + name + ".in"

	err := ioutil.WriteFile(outputFileLoc, []byte(""), 0644)
	if err != nil {
		util.Warn("outputstream: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: ""}
		return
	}
	err = ioutil.WriteFile(inputFileLoc, []byte(session.CurrentBatch.Input), 0644)
	if err != nil {
		util.Warn("inputstream: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: ""}
		return
	}

	session.OutputStream, err = os.Open(outputFileLoc)
	if err != nil {
		util.Warn("outputstream: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: ""}
		return
	}

	session.InputStream, err = os.Open(inputFileLoc)
	if err != nil {
		util.Warn("inputstream: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: ""}
		return
	}

}

/*
 * Run to start judging
 */

func (session *GradeSession) StartJudging() {
	runtime.LockOSThread() // https://github.com/golang/go/issues/7699
	defer runtime.UnlockOSThread()

	// setup tracer
	tracer := PTracer{
		Session:     session,
		StreamDone:  session.StreamDone,
		ExecCommand: session.ExecCommand,
		ExecArgs:    session.ExecArgs,
	}

	// init pipes
	session.InitStreams()

	tracer.ForkExec()
	session.Pid = tracer.Pid

	// time
	session.StartTime = time.Now()
	go session.WaitTLE()
	go StartGrader(session, session.Pid, &session.CurrentBatch.ExpectedAnswer, session.StreamDone)
	go session.ListenStderr()

	if shared.SANDBOX {
		defer session.CloseStreams()

		// sandbox has to hog the thread, so move receiving to another one
		go session.WaitVerdict()

		// start sandboxing
		// must run on this thread because all ptrace calls have to come from one thread
		tracer.Trace()
	} else {
		session.CloseStreams()
		session.WaitVerdict()
	}

}

/*
 * Dump stderr to session
 */

func (session *GradeSession) ListenStderr() {
	buff := bufio.NewReader(session.ErrorBuffer)
	for {
		if !util.IsPidRunning(session.Pid) { // if the program has ended
			break
		}

		c, _, err := buff.ReadRune()
		if err != nil {
			continue
		}
		session.Stderr += string(c) // append to stderr
	}
}

/*
 * Timeout for TLE
 */

func (session *GradeSession) WaitTLE() {
	time.Sleep(time.Duration(session.CurrentBatch.TimeLimit)*time.Second)

	if !session.StreamDoneUsed {
		shared.Debug("TLE")
		session.StreamDone <- CaseReturn{Result: shared.OUTCOME_TLE}
	}
}

/*
 * Wait for result from other goroutines
 */

func (session *GradeSession) WaitVerdict() {
	defer func() { // prevent buffer from closing too early
		if session.ErrorBuffer != nil {
			session.ErrorBuffer.Close()
		}
	}()

	// wait for judging to finish
	response := <-session.StreamDone
	session.StreamDoneUsed = true

	// kill process if still running
	if util.IsPidRunning(session.Pid) {
		unix.Kill(session.Pid, syscall.SIGTERM)
		unix.Kill(session.Pid, syscall.SIGKILL) // extra assurance
		var wstatus unix.WaitStatus
		unix.Wait4(session.Pid, &wstatus, unix.WALL|unix.WNOHANG, nil) // collect zombie
	}

	// send result back to runner
	if response.Result == shared.OUTCOME_WA && session.Stderr != "" { //  if the outcome was wrong answer but there was an error
		session.StreamResult <- pb.TestCaseResult{
			Result:     shared.OUTCOME_RTE,
			ResultInfo: session.Stderr,
			Time:       time.Since(session.StartTime).Seconds(),
			MemUsage:   0, // TODO
			BatchNumber: session.BatchNum,
			CaseNumber: session.CaseNum,
		}
	} else if session.ExitCode != 0 && session.ExitCode != -1 { // if the program did not exit successfully
		session.StreamResult <- pb.TestCaseResult{
			Result:     shared.OUTCOME_RTE,
			ResultInfo: fmt.Sprintf("Exit code: %v: %v", session.ExitCode, session.Stderr),
			Time:       time.Since(session.StartTime).Seconds(),
			MemUsage:   0,
			BatchNumber: session.BatchNum,
			CaseNumber: session.CaseNum,
		}
	} else { // if the program exited successfully
		session.StreamResult <- pb.TestCaseResult{
			Result:     response.Result,
			ResultInfo: response.ResultInfo,
			Time:       time.Since(session.StartTime).Seconds(),
			MemUsage:   0, // TODO
			BatchNumber: session.BatchNum,
			CaseNumber: session.CaseNum,
		}
	}
}

/*
 * Cleanup streams
 */

func (session *GradeSession) CloseStreams() {
	if session.OutputStream != nil {
		session.OutputStream.Close()
	}
	/*if session.OutputBuffer != nil {
		session.OutputBuffer.Close() NO
	}*/
	if session.ErrorStream != nil {
		session.ErrorStream.Close()
	}
	//if session.ErrorBuffer != nil {
	//	session.ErrorBuffer.Close()
	//}
	if session.InputStream != nil {
		session.InputStream.Close()
	}
}