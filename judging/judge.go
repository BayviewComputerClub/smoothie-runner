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
	Result     string
	ResultInfo string
}

/*
 * run process as noperm user (or just run the whole program with no perm)
 */

func judgeStderrListener(session *shared.JudgeSession, pid int) {
	buff := bufio.NewReader(session.ErrorBuffer)
	for {
		if !util.IsPidRunning(pid) { // if the program has ended
			break
		}

		c, _, err := buff.ReadRune()
		if err != nil {
			continue
		}
		session.Stderr += string(c) // append to stderr
	}
}

// check for TLE

func judgeCheckTimeout(c *exec.Cmd, d time.Duration, done chan CaseReturn) {
	hasExited := false
	time.Sleep(d)

	go func() { // wait for output by other processes
		<-done
		hasExited = true
	}()

	if !hasExited {
		done <- CaseReturn{Result: shared.OUTCOME_TLE}
	}
}

// pipe test input to buffer

func initInputStream(c *exec.Cmd, session *shared.JudgeSession, input string) *os.File {
	// possibly use fake file instead so code can't access
	inputFileLoc := session.Workspace + "/" + strconv.FormatInt(time.Now().Unix(), 10) + ".in"
	err := ioutil.WriteFile(inputFileLoc, []byte(input), 0644)
	if err != nil {
		util.Warn("inputstream: " + err.Error())
		return nil
	}
	inputFile, err := os.Open(inputFileLoc)
	if err != nil {
		util.Warn("inputstream: " + err.Error())
		return nil
	}
	c.Stdin = inputFile
	return inputFile
}

// judge individual batch case

func judgeCase(c *exec.Cmd, session *shared.JudgeSession, batchCase *pb.ProblemBatchCase, result chan pb.TestCaseResult) {
	runtime.LockOSThread() // https://github.com/golang/go/issues/7699
	defer runtime.UnlockOSThread()

	done := make(chan CaseReturn)

	t := time.Now()

	if shared.SANDBOX {
		// enable ptrace
		c.SysProcAttr = &unix.SysProcAttr{Ptrace: true}
	}

	var err error
	// initialize pipes
	session.InputStream = initInputStream(c, session, batchCase.Input)
	if session.InputStream == nil {
		result <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: ""}
		return
	}

	// stdout buffer
	session.OutputBuffer, session.OutputStream, err = os.Pipe()
	if err != nil {
		util.Warn("stdoutpipeinit: " + err.Error())
		result <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}
	c.Stdout = session.OutputStream

	// stderr buffer
	session.ErrorBuffer, session.ErrorStream, err = os.Pipe()
	if err != nil {
		util.Warn("stderrpipeinit: " + err.Error())
		result <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}
	c.Stderr = session.ErrorStream

	// start process
	err = c.Start()
	if err != nil {
		util.Warn("RTE: " + err.Error())
		result <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	// start time watch (convert to seconds)
	go judgeCheckTimeout(c, time.Duration(batchCase.TimeLimit)*time.Second, done)

	if shared.SANDBOX {
		c.Wait() // pause execution on first instruction for sandbox to start
	}

	// start listener pipes
	go StartGrader(session, c.Process.Pid, &batchCase.ExpectedAnswer, done)
	go judgeStderrListener(session, c.Process.Pid)

	go func() {
		err = c.Wait() // make sure exit code is retrieved to prevent zombie process in nonsandbox environment

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
		go judgeWaitForResponse(session, c, t, done, result)

		// start sandboxing
		// must run on this thread because all ptrace calls have to come from one thread
		sandboxProcess(session, &c.Process.Pid, done)
	} else {
		judgeWaitForResponse(session, c, t, done, result)
	}
}

// receive response from judging processes

func judgeWaitForResponse(session *shared.JudgeSession, c *exec.Cmd, t time.Time, done chan CaseReturn, result chan pb.TestCaseResult) {
	// wait for judging to finish
	response := <-done

	// kill process if still running
	if util.IsPidRunning(c.Process.Pid) {
		unix.Kill(c.Process.Pid, syscall.SIGTERM)
		unix.Kill(c.Process.Pid, syscall.SIGKILL) // extra assurance
		err := c.Process.Signal(syscall.SIGKILL)
		if err != nil && err.Error() != "os: process already finished" {
			util.Warn("pkill fail: " + err.Error())
		}
	}

	// send result back to runner
	if session.ExitCode != 0 && session.ExitCode != -1 { // if the program did not exit successfully
		result <- pb.TestCaseResult{
			Result:     shared.OUTCOME_RTE,
			ResultInfo: fmt.Sprintf("Exit code: %v: %v", session.ExitCode, session.Stderr),
			Time:       time.Since(t).Seconds(),
			MemUsage:   0,
		}
	} else { // if the program exited successfully
		result <- pb.TestCaseResult{
			Result:     response.Result,
			ResultInfo: response.ResultInfo,
			Time:       time.Since(t).Seconds(),
			MemUsage:   0, // TODO
		}
	}
}
