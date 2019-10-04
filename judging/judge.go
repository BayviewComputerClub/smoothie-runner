package judging

import (
	"bufio"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"
)

type CaseReturn struct {
	Result     string
	ResultInfo string
}

/*
 * Tasks:
 * setrlimit to prevent fork() in processes (use golang.org/x/sys)
 * set nice value to be low
 * run process as noperm user (or just run the whole program with no perm)
 * use ptrace unix calls
 * can use syscall credentials
 * make sure thread is locked
 */

func judgeStdoutListener(cmd *exec.Cmd, reader *io.ReadCloser, done chan CaseReturn, expectedAnswer *string) {
	buff := bufio.NewReader(*reader)

	expectedScanner := bufio.NewScanner(strings.NewReader(*expectedAnswer))
	expectedScanner.Scan()

	expectedLine := []rune(strings.ReplaceAll(expectedScanner.Text(), "\r", ""))
	expectedLineIndex := 0
	expectingEnd := false
	ignoreSpacesIndex := 0

	// calculate when to start ignoring spaces (at the end of the line before \n or \r\n
	for i := len(expectedLine) - 1; i >= 0; i-- {
		if expectedLine[i] != '\n' && expectedLine[i] != ' ' {
			ignoreSpacesIndex = i + 1
			break
		}
	}

	// loop through to read rune by rune
	for {
		if cmd.ProcessState.Exited() {
			done <- CaseReturn{
				Result: shared.OUTCOME_WA,
			}
			break
		}

		// read rune to parse
		c, _, err := buff.ReadRune()

		// check for errors to return
		if err != nil {
			if err == io.EOF { // no more output
				if expectingEnd { // expected no more text
					done <- CaseReturn{
						Result: shared.OUTCOME_AC,
					}
				} else { // did not finish giving full answer
					done <- CaseReturn{
						Result: shared.OUTCOME_WA,
					}
				}
				return
			} else { // other errors
				done <- CaseReturn{
					Result:     shared.OUTCOME_RTE,
					ResultInfo: err.Error(), // TODO
				}
				return
			}
		}

		// validate obtained rune

		if expectingEnd { // if expecting no more text

			if c != '\n' && c != ' ' { // if not new line and space
				done <- CaseReturn{
					Result:     shared.OUTCOME_WA,
					ResultInfo: "bruh",
				}
				return
			}

		} else if c == '\n' { // if new line is detected

			if expectedLineIndex >= ignoreSpacesIndex { // if waiting for new line

				if expectedScanner.Scan() { // if there is more expected output
					// reset for next line of reading
					expectedLine = []rune(strings.ReplaceAll(expectedScanner.Text(), "\r", ""))
					ignoreSpacesIndex = 0
					expectedLineIndex = -1 // it adds at end

					// calculate when to start ignoring spaces (at the end of the line before \n or \r\n
					for i := len(expectedLine) - 1; i >= 0; i-- {
						if expectedLine[i] != '\n' && expectedLine[i] != ' ' {
							ignoreSpacesIndex = i + 1
							break
						}
					}
				} else { // if there isn't
					expectingEnd = true
				}

			} else { // if not waiting for new line
				done <- CaseReturn{
					Result: shared.OUTCOME_WA,
				}
				return
			}

		} else if expectedLineIndex >= len(expectedLine) || c != expectedLine[expectedLineIndex] { // if character did not match expected character

			if !(expectedLineIndex >= ignoreSpacesIndex && c == ' ') { // if not waiting for space or new line, or character is not space or new line
				done <- CaseReturn{
					Result: shared.OUTCOME_WA,
				}
				return
			}
		}

		expectedLineIndex++
	}
}

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

func judgeStdinFeeder(writer *io.WriteCloser, done chan CaseReturn, feed *string) {
	buff := bufio.NewWriter(*writer)
	_, err := buff.WriteString(*feed)
	if err != nil {
		done <- CaseReturn{
			Result: shared.OUTCOME_RTE,
		}
		util.Warn("Stdin: " + err.Error()) // TODO
		return
	}
}

func judgeCheckTimeout(c *exec.Cmd, d time.Duration, done chan CaseReturn) {
	time.Sleep(d)
	if !c.ProcessState.Exited() {
		done <- CaseReturn {Result: shared.OUTCOME_TLE}
	}
}

func judgeCase(c *exec.Cmd, batchCase *pb.ProblemBatchCase, result chan pb.TestCaseResult) {
	done := make(chan CaseReturn)

	t := time.Now()

	// initialize pipes
	stderrPipe, err := c.StderrPipe()
	if err != nil {
		util.Warn(err.Error())
		result <- pb.TestCaseResult{Result: shared.OUTCOME_ISE,}
		return
	}
	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		util.Warn(err.Error())
		result <- pb.TestCaseResult{Result: shared.OUTCOME_ISE,}
		return
	}
	stdinPipe, err := c.StdinPipe()
	if err != nil {
		util.Warn(err.Error())
		result <- pb.TestCaseResult{Result: shared.OUTCOME_ISE,}
		return
	}

	// start listener pipes
	go judgeStdoutListener(c, &stdoutPipe, done, &batchCase.ExpectedAnswer)
	go judgeStderrListener(&stderrPipe, done)

	// start process
	err = c.Start()
	if err != nil {
		util.Warn(err.Error())
		result <- pb.TestCaseResult{Result: shared.OUTCOME_ISE,}
		return
	}

	// start time watch (convert to seconds)
	go judgeCheckTimeout(c, time.Duration(batchCase.TimeLimit*1000000000) * time.Second, done)

	// start sandboxing
	pid := c.Process.Pid
	go sandboxProcess(pid, done) // this will reserve a thread

	// feed input to process
	go judgeStdinFeeder(&stdinPipe, done, &batchCase.Input)

	// wait for judging to finish
	response := <-done

	if !c.ProcessState.Exited() {
		err = c.Process.Kill()
		if err != nil {
			util.Warn(err.Error())
		}
	}

	result <- pb.TestCaseResult{
		Result:     response.Result,
		ResultInfo: response.ResultInfo,
		Time:       time.Since(t).Seconds(),
		MemUsage:   0, // TODO
	}
}
