package judging

import (
	"bufio"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"io"
	"os"
	"os/exec"
	"strings"
)

// compare expected answer with stream

func judgeStdoutListener(cmd *exec.Cmd, reader *os.File, done chan CaseReturn, expectedAnswer *string) {
	buff := bufio.NewReader(reader)

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
		if !util.IsPidRunning(cmd.Process.Pid) {
			if expectingEnd { // expected no more text
				done <- CaseReturn{
					Result: shared.OUTCOME_AC,
				}
			} else { // did not finish giving full answer
				done <- CaseReturn{
					Result: shared.OUTCOME_WA,
				}
			}
			break
		}

		if buff.Size() == 0 {
			continue
		}

		// read rune to parse
		c, _, err := buff.ReadRune()

		//println(string(c)) // TODO
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
