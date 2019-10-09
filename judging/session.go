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

	// streams during judging
	OutputBuffer *os.File
	ErrorBuffer  *os.File
	OutputStream *os.File
	ErrorStream  *os.File
	InputStream  *os.File

	// per judge session
	Stderr   string // error dumped here
	ExitCode int

	StreamResult chan pb.TestCaseResult // return result to runner
	StreamDone chan CaseReturn // end batch case with verdict
	StartTime time.Time

	Command *exec.Cmd
}

/*
 * Initialize all streams before program starts
 */

func (session *GradeSession) InitStreams() {
	// init input
	session.StartInput()

	var err error
	// stdout buffer
	session.OutputBuffer, session.OutputStream, err = os.Pipe()
	if err != nil {
		util.Warn("stdoutpipeinit: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}
	session.Command.Stdout = session.OutputStream

	// stderr buffer
	session.ErrorBuffer, session.ErrorStream, err = os.Pipe()
	if err != nil {
		util.Warn("stderrpipeinit: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}
	session.Command.Stderr = session.ErrorStream
}

/*
 * Initializes input stream
 */

func (session *GradeSession) StartInput() {

	// possibly use fake file instead so code can't access
	inputFileLoc := session.JudgingSession.Workspace + "/" + strconv.FormatInt(time.Now().Unix(), 10) + ".in"
	err := ioutil.WriteFile(inputFileLoc, []byte(session.CurrentBatch.Input), 0644)
	if err != nil {
		util.Warn("inputstream: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: ""}
		return
	}
	inputFile, err := os.Open(inputFileLoc)
	if err != nil {
		util.Warn("inputstream: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: ""}
		return
	}
	session.Command.Stdin = inputFile
	session.InputStream = inputFile
}

/*
 * Run to start judging
 */

func (session *GradeSession) StartJudging() {
	runtime.LockOSThread() // https://github.com/golang/go/issues/7699
	defer runtime.UnlockOSThread()

	if shared.SANDBOX {
		// enable ptrace
		session.Command.SysProcAttr = &unix.SysProcAttr{Ptrace: true}
	}

	// init pipes
	session.InitStreams()

	// time
	session.StartTime = time.Now()
	go session.WaitTLE()

	err := session.Command.Start()
	if err != nil {
		util.Warn("RTE: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	if shared.SANDBOX {
		session.Command.Wait() // pause execution on first instruction for ptrace to start
	}

	go StartGrader(session, session.Command.Process.Pid, &session.CurrentBatch.ExpectedAnswer, session.StreamDone)
	go session.ListenStderr()

	go func() {
		err = session.Command.Wait() // make sure exit code is retrieved to prevent zombie process in nonsandbox environment

		if !shared.SANDBOX {

			session.ExitCode = 0
			// save exit code
			if err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						session.ExitCode = int(status.Signal())
					}
				}
			}

			// close streams so that grader will finish grading
			session.CloseStreams()
		}
	}()

	if shared.SANDBOX {
		defer session.CloseStreams() // sandbox ends when process ends

		// sandbox has to hog the thread, so move receiving to another one
		go session.WaitVerdict()

		// start sandboxing
		// must run on this thread because all ptrace calls have to come from one thread
		sandboxProcess(session, &session.Command.Process.Pid, session.StreamDone)
	} else {
		session.WaitVerdict()
	}

}

/*
 * Dump stderr to session
 */

func (session *GradeSession) ListenStderr() {
	buff := bufio.NewReader(session.ErrorBuffer)
	for {
		if !util.IsPidRunning(session.Command.Process.Pid) { // if the program has ended
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
	hasExited := false
	time.Sleep(time.Duration(session.CurrentBatch.TimeLimit)*time.Second)

	go func() { // wait for output by other processes
		<-session.StreamDone
		hasExited = true
	}()

	if !hasExited {
		session.StreamDone <- CaseReturn{Result: shared.OUTCOME_TLE}
	}
}

/*
 * Wait for result from other goroutines
 */

func (session *GradeSession) WaitVerdict() {
	// wait for judging to finish
	response := <-session.StreamDone

	// kill process if still running
	if util.IsPidRunning(session.Command.Process.Pid) {
		unix.Kill(session.Command.Process.Pid, syscall.SIGTERM)
		unix.Kill(session.Command.Process.Pid, syscall.SIGKILL) // extra assurance
		err := session.Command.Process.Signal(syscall.SIGKILL)
		if err != nil && err.Error() != "os: process already finished" {
			util.Warn("pkill fail: " + err.Error())
		}
	}

	// send result back to runner
	if session.ExitCode != 0 && session.ExitCode != -1 { // if the program did not exit successfully
		session.StreamResult <- pb.TestCaseResult{
			Result:     shared.OUTCOME_RTE,
			ResultInfo: fmt.Sprintf("Exit code: %v: %v", session.ExitCode, session.Stderr),
			Time:       time.Since(session.StartTime).Seconds(),
			MemUsage:   0,
		}
	} else { // if the program exited successfully
		session.StreamResult <- pb.TestCaseResult{
			Result:     response.Result,
			ResultInfo: response.ResultInfo,
			Time:       time.Since(session.StartTime).Seconds(),
			MemUsage:   0, // TODO
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
	if session.OutputBuffer != nil {
		session.OutputBuffer.Close()
	}
	if session.ErrorStream != nil {
		session.ErrorStream.Close()
	}
	if session.ErrorBuffer != nil {
		session.ErrorBuffer.Close()
	}
	if session.InputStream != nil {
		session.InputStream.Close()
	}
}