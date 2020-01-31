package judging

import (
	"bufio"
	"fmt"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"math"
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
	Limit *shared.Rlimits

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
	StreamProcEnd chan bool // wait until the process has stopped running
	StartTime time.Time

	Command *exec.Cmd
	ExecCommand uintptr
	ExecArgs uintptr

	Pid int
	MemoryUsage float64
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
	name := strconv.FormatInt(time.Now().UnixNano(), 10)

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

	session.OutputStream, err = os.OpenFile(outputFileLoc, os.O_RDWR, os.ModeAppend) // open with read/write fd
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

	// setup proc
	proc := ForkProcess{
		Session:     session,
		StreamDone:  session.StreamDone,
		ExecCommand: session.ExecCommand,
		ExecArgs:    session.ExecArgs,
	}

	// init pipes
	session.InitStreams()
	defer session.CloseStreams()

	proc.ForkExec()

	session.Pid = proc.Pid
	session.StartTime = time.Now()

	go session.WaitProcState()
	go StartGrader(session)
	go session.ListenStderr()

	session.WaitVerdict()
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

func (session *GradeSession) WaitProcState() {
	var (
		wstatus unix.WaitStatus
	    rusage unix.Rusage
		)

	// TODO proc timeout

	for {
		// wait for process change state
		_, err := unix.Wait4(session.Pid, &wstatus, 0, &rusage)
		println("WAIT4: ", wstatus) // TODO
		if err != nil {
			shared.Debug("wait4: " + err.Error())
			session.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
		}

		// check tle
		t := time.Now()
		if session.Limit.CpuTime != 0 && t.After(session.StartTime.Add(time.Duration(session.Limit.CpuTime*1e9))) {
			shared.Debug("TLE")
			session.StreamDone <- CaseReturn{Result: shared.OUTCOME_TLE}
			return
		}

		// update memory usage
		session.MemoryUsage = math.Max(session.MemoryUsage, float64(rusage.Maxrss)/1000)

		// check memory
		// maxrss - KB, memlimit - MB
		if session.Limit.Memory != 0 && rusage.Maxrss > int64(session.Limit.Memory*1000) {
			shared.Debug("MLE")
			session.StreamDone <- CaseReturn{Result: shared.OUTCOME_MLE}
			return
		}

		switch {
		case wstatus.Exited():
			session.ExitCode = wstatus.ExitStatus()
			// send to grader
			// grader is expected to send to session.StreamDone when finished grading
			session.StreamProcEnd <- true
			return
		case wstatus.Signaled():
			sig := wstatus.Signal()
			session.ExitCode = int(wstatus.Signal())
			switch sig {
			case unix.SIGXCPU, unix.SIGKILL:
				session.StreamDone <- CaseReturn{Result: shared.OUTCOME_TLE}
			case unix.SIGXFSZ:
				session.StreamDone <- CaseReturn{Result: shared.OUTCOME_OLE}
			case unix.SIGSYS:
				session.StreamDone <- CaseReturn{Result: shared.OUTCOME_ILL}
			default:
				session.StreamDone <- CaseReturn{Result: shared.OUTCOME_RTE}
			}
			return
		}

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

	// kill process if still running
	if util.IsPidRunning(session.Pid) {
		unix.Kill(session.Pid, syscall.SIGTERM)
		unix.Kill(session.Pid, syscall.SIGKILL) // extra assurance
		var wstatus unix.WaitStatus
		unix.Wait4(session.Pid, &wstatus, unix.WALL|unix.WNOHANG, nil) // collect zombie
	}

	// send result back to runner
	if response.Result != shared.OUTCOME_AC && session.Stderr != "" { //  if the outcome was wrong answer but there was an error
		session.StreamResult <- pb.TestCaseResult{
			Result:     shared.OUTCOME_RTE,
			ResultInfo: session.Stderr,
			Time:       time.Since(session.StartTime).Seconds(),
			MemUsage:   session.MemoryUsage,
			BatchNumber: session.BatchNum,
			CaseNumber: session.CaseNum,
		}
	} else if response.Result == shared.OUTCOME_RTE { // if the program did not exit successfully
		session.StreamResult <- pb.TestCaseResult{
			Result:     shared.OUTCOME_RTE,
			ResultInfo: fmt.Sprintf("Exit code: %v: %v", session.ExitCode, session.Stderr),
			Time:       time.Since(session.StartTime).Seconds(),
			MemUsage:   session.MemoryUsage,
			BatchNumber: session.BatchNum,
			CaseNumber: session.CaseNum,
		}
	} else { // if the program exited successfully
		session.StreamResult <- pb.TestCaseResult{
			Result:     response.Result,
			ResultInfo: response.ResultInfo,
			Time:       time.Since(session.StartTime).Seconds(),
			MemUsage:   session.MemoryUsage,
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
	if session.ErrorStream != nil {
		session.ErrorStream.Close()
	}
	if session.InputStream != nil {
		session.InputStream.Close()
	}
}