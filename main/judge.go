package main

import (
	"bufio"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"
)

const (
	OUTCOME_AC  = "AC"  // answer correct
	OUTCOME_WA  = "WA"  // wrong answer
	OUTCOME_RTE = "RTE" // run time error
	OUTCOME_CE  = "CE"  // compile error
	OUTCOME_TLE = "TLE" // time limit exceeded
	OUTCOME_MLE = "MLE" // memory limit exceeded
	OUTCOME_ILL = "ILL" // illegal operation
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
 * set timeout
 */

func judgeStdoutListener(reader *io.ReadCloser, done chan CaseReturn, expectedAnswer *string) {
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

		// read rune to parse
		c, _, err := buff.ReadRune()

		// check for errors to return
		if err != nil {
			if err == io.EOF { // no more output
				if expectingEnd { // expected no more text
					done <- CaseReturn{
						Result: OUTCOME_AC,
					}
				} else { // did not finish giving full answer
					done <- CaseReturn{
						Result: OUTCOME_WA,
					}
				}
				return
			} else { // other errors
				done <- CaseReturn{
					Result:     OUTCOME_RTE,
					ResultInfo: err.Error(), // TODO
				}
				return
			}
		}

		// validate obtained rune

		if expectingEnd { // if expecting no more text

			if c != '\n' && c != ' ' { // if not new line and space
				done <- CaseReturn{
					Result:     OUTCOME_WA,
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
					Result: OUTCOME_WA,
				}
				return
			}

		} else if expectedLineIndex >= len(expectedLine) || c != expectedLine[expectedLineIndex] { // if character did not match expected character

			if !(expectedLineIndex >= ignoreSpacesIndex && c == ' ') { // if not waiting for space or new line, or character is not space or new line
				done <- CaseReturn{
					Result: OUTCOME_WA,
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
		warn("Stderr: " + err.Error()) // TODO
	} else {
		done <- CaseReturn{
			Result: OUTCOME_RTE,
			ResultInfo: string(str),
		}
	}
}

func judgeStdinFeeder(writer *io.WriteCloser, done chan CaseReturn, feed *string) {
	buff := bufio.NewWriter(*writer)
	_, err := buff.WriteString(*feed)
	if err != nil {
		done <- CaseReturn{
			Result: OUTCOME_RTE,
		}
		warn("Stdin: " + err.Error()) // TODO
		return
	}
}

func judgeCase(c *exec.Cmd, batchCase *pb.ProblemBatchCase) *pb.TestCaseResult {
	done := make(chan CaseReturn)

	t := time.Now()

	// initialize pipes
	stderrPipe, err := c.StderrPipe()
	if err != nil {
		warn(err.Error())
		return nil
	}
	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		warn(err.Error())
		return nil
	}
	stdinPipe, err := c.StdinPipe()
	if err != nil {
		warn(err.Error())
		return nil
	}

	// start listener pipes
	go judgeStdoutListener(&stdoutPipe, done, &batchCase.ExpectedAnswer)
	go judgeStderrListener(&stderrPipe, done)

	// start process
	err = c.Start()
	if err != nil {
		warn(err.Error())
		return nil
	}

	// start sandboxing
	pid := c.Process.Pid
	go sandboxProcess(pid, done)

	// feed input to process
	go judgeStdinFeeder(&stdinPipe, done, &batchCase.Input)

	// wait for judging to finish
	response := <-done
	return &pb.TestCaseResult{
		Result:      response.Result,
		ResultInfo:  response.ResultInfo,
		Time:        time.Since(t).Seconds(),
		MemUsage:    0, // TODO
	}
}

