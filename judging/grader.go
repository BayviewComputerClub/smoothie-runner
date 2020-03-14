package judging

import (
	"bufio"
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

var (
	graders = make(map[string]Grader)
)

func init() {
	graders["strict"] = StrictGrader{}
	graders["endtrim"] = EndTrimGrader{}
}

type Grader interface {
	// must wait on session.StreamProcEnd and send output to done
	CompareStream(session *GradeSession, expectedAnswer *os.File, done chan CaseReturn)
}

func StartGrader(session *GradeSession) {
	if grader, ok := graders[session.JudgingSession.OriginalRequest.Problem.Grader.Type]; ok {
		grader.CompareStream(session, session.CurrentCase.Output, session.StreamDone)
	} else {
		session.StreamDone <- CaseReturn{
			Result:     shared.OUTCOME_ISE,
			ResultInfo: "Grader not found.",
		}
	}
}

// ***** Strict Grader *****
// Requires exact output, including whitespace
// ignores extra new line at end

type StrictGrader struct {}

func (grader StrictGrader) CompareStream(session *GradeSession, expectedAnswerFile *os.File, done chan CaseReturn) {
	expectedAnswer, err := ioutil.ReadAll(expectedAnswerFile)
	if err != nil {
		done <- CaseReturn{
			Result:     shared.OUTCOME_ISE,
			ResultInfo: "could not read answer file",
		}
		return
	}

	// move index for reading the file to beginning
	_, _ = session.OutputStream.Seek(0, 0)

	// read from output stream
	buff := bufio.NewReader(session.OutputStream)
	expectingEnd := false
	ansIndex := 0
	ans := []rune(strings.ReplaceAll(string(expectedAnswer), "\r", ""))

	for {
		c, _, err := buff.ReadRune()
		if err != nil {
			shared.Debug("readrune: " + err.Error())
			if true /*err != io.EOF*/ {
				if expectingEnd { // expected no more text
					done <- CaseReturn{
						Result: shared.OUTCOME_AC,
					}
				} else { // did not finish giving full answer
					done <- CaseReturn{
						Result: shared.OUTCOME_WA,
						ResultInfo: "Ended early",
					}
				}
				break
			}
			continue
		}

		shared.Debug(string(c) + " (" + fmt.Sprint(c) + ")")

		// if wrong character or expecting no output
		// ignore new line at end
		if !(expectingEnd && c == '\n') && ((expectingEnd && c != '\n') || (ansIndex < len(ans) && c != ans[ansIndex])) {
			done <- CaseReturn{
				Result:     shared.OUTCOME_WA,
				ResultInfo: "Wrong char",
			}
			break
		}

		// expecting end when reach the end of the file
		if ansIndex >= len(ans) - 1 || (ans[len(ans)-1] == '\n' && ansIndex >= len(ans) - 2) {
			expectingEnd = true
		}

		ansIndex++
	}
}

// ***** EndTrim Grader *****
// Ignores whitespace at the end of a line, and new lines characters at the end

type EndTrimGrader struct {}

func (grader EndTrimGrader) CompareStream(session *GradeSession, expectedAnswerFile *os.File, done chan CaseReturn) {
	expectedAnswer, err := ioutil.ReadAll(expectedAnswerFile)
	if err != nil {
		done <- CaseReturn{
			Result:     shared.OUTCOME_ISE,
			ResultInfo: "could not read answer file",
		}
	}

	buff := bufio.NewReader(session.OutputStream)

	expectedScanner := bufio.NewScanner(strings.NewReader(string(expectedAnswer)))
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
		if !util.IsPidRunning(session.RunnerSession.Pid) { // possibly should move this after all runes are read
			if expectingEnd { // expected no more text
				done <- CaseReturn{
					Result: shared.OUTCOME_AC,
				}
			} else { // did not finish giving full answer
				done <- CaseReturn{
					Result: shared.OUTCOME_WA,
					ResultInfo: "Ended early",
				}
			}
			break
		}

		if buff.Size() == 0 {
			continue
		}

		// read rune to parse
		c, _, err := buff.ReadRune()

		shared.Debug(string(c)) // TODO
		if err != nil {
			if err != io.EOF {
				//util.Warn("readrune: " + err.Error())
			}
			continue
		}

		// validate obtained rune

		if expectingEnd { // if expecting no more text

			if c != '\n' && c != ' ' { // if not new line and space
				done <- CaseReturn{
					Result:     shared.OUTCOME_WA,
					ResultInfo: "Wrong char",
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
