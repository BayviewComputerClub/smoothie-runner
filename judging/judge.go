package judging

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

type CaseReturn struct {
	Result     string
	ResultInfo string
}

/*
 * run process as noperm user (or just run the whole program with no perm)
 */

func judgeStderrListener(reader *io.ReadCloser, done chan CaseReturn) {
	str, err := ioutil.ReadAll(*reader)

	if err != nil { // should terminate peacefully
		util.Warn("Stderr: " + err.Error()) // TODO
	} else {
		done <- CaseReturn{
			Result:     shared.OUTCOME_RTE,
			ResultInfo: string(str),
		}
	}
}

// check for TLE

func judgeCheckTimeout(c *exec.Cmd, d time.Duration, done chan CaseReturn) {
	time.Sleep(d)
	if util.IsPidRunning(c.Process.Pid) {
		done <- CaseReturn {Result: shared.OUTCOME_TLE}
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

func judgeCase(c *exec.Cmd, session shared.JudgeSession, batchCase *pb.ProblemBatchCase, result chan pb.TestCaseResult) {
	runtime.LockOSThread() // https://github.com/golang/go/issues/7699
	defer runtime.UnlockOSThread()

	done := make(chan CaseReturn)

	t := time.Now()

	if shared.SANDBOX {
		// enable ptrace
		c.SysProcAttr = &unix.SysProcAttr{Ptrace: true}
	}

	// initialize pipes

	/*stderrPipe, err := c.StderrPipe()
	if err != nil {
		util.Warn("stderrpipeinit: " + err.Error())
		result <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}*/

	f := initInputStream(c, &session, batchCase.Input)
	if f == nil {
		result <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: ""}
		return
	}
	defer f.Close()

	outputBuff, outStream, err := os.Pipe()
	if err != nil {
		panic(err) // TODO
	}
	c.Stdout = outStream
	defer outputBuff.Close()
	defer outStream.Close()

	// start process
	err = c.Start()
	if err != nil {
		util.Warn("RTE: " + err.Error())
		result <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		return
	}

	// start time watch (convert to seconds)
	go judgeCheckTimeout(c, time.Duration(batchCase.TimeLimit) * time.Second, done)

	c.Wait() // pause execution on first instruction

	// start listener pipes
	go StartGrader(&session, c.Process.Pid, outputBuff, &batchCase.ExpectedAnswer, done)
	//go judgeStderrListener(&stderrPipe, done)

	// sandbox has to hog the thread, so move receiving to another one
	go judgeWaitForResponse(c, t, done, result)

	if shared.SANDBOX {
		// start sandboxing
		// must run on this thread because all ptrace calls have to come from one thread
		sandboxProcess(&c.Process.Pid, done)
	}
}

// receive response from judging processes

func judgeWaitForResponse(c *exec.Cmd, t time.Time, done chan CaseReturn, result chan pb.TestCaseResult) {
	// wait for judging to finish
	response := <-done

	// kill process if still running
	if util.IsPidRunning(c.Process.Pid) {
		err := c.Process.Kill()
		if err != nil  && err.Error() != "os: process already finished" {
			util.Warn("pkill fail: " + err.Error())
		}
	}

	// send result back to runner
	result <- pb.TestCaseResult{
		Result:     response.Result,
		ResultInfo: response.ResultInfo,
		Time:       time.Since(t).Seconds(),
		MemUsage:   0, // TODO
	}
}