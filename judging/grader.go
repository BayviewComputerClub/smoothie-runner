package judging

import (
	"bufio"
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"io"
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
	CompareStream(session *shared.JudgeSession, pid int, expectedAnswer *string, done chan CaseReturn)
}

func StartGrader(session *shared.JudgeSession, pid int, expectedAnswer *string, done chan CaseReturn) {
	if grader, ok := graders[session.OriginalRequest.Solution.Problem.Grader.Type]; ok {
		grader.CompareStream(session, pid, expectedAnswer, done)
	} else {
		done <- CaseReturn{
			Result:     shared.OUTCOME_ISE,
			ResultInfo: "Grader not found.",
		}
	}
}

// ***** Strict Grader *****
// Requires exact output, including whitespace
// ignores extra new line at end

type StrictGrader struct {}

func (grader StrictGrader) CompareStream(session *shared.JudgeSession, pid int, expectedAnswer *string, done chan CaseReturn) {
	buff := bufio.NewReader(session.OutputBuffer)
	expectingEnd := false
	ansIndex := 0
	ans := []rune(strings.ReplaceAll(*expectedAnswer, "\r", ""))

	for {
		if !util.IsPidRunning(pid) { // if the program has ended
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

		// is the buffer empty
		if buff.Size() == 0 {
			continue
		}

		c, _, err := buff.ReadRune()
		if err != nil {
			shared.Debug("readrune: " + err.Error())
			// if err != io.EOF {}
			continue
		}

		shared.Debug(string(c) + " (" + fmt.Sprint(c) + ")")

		// if wrong character or expecting no output
		// ignore new line at end
		if (expectingEnd && c != '\n') || (ansIndex < len(ans) && c != ans[ansIndex]) {
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

func (grader EndTrimGrader) CompareStream(session *shared.JudgeSession, pid int, expectedAnswer *string, done chan CaseReturn) {
	buff := bufio.NewReader(session.OutputBuffer)

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
		if !util.IsPidRunning(pid) { // possibly should move this after all runes are read
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