package judging

import (
	"bufio"
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"io/ioutil"
	"os"
	"strings"
	"strconv"
	"container/list"
)

var (
	graders = make(map[string]Grader)
)

func init() {
	graders["strict"] = StrictGrader{}
	graders["endtrim"] = EndTrimGrader{}
	graders["double"] = DoubleGrader{}
	graders["field"] = FieldGrader{}
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
// ignores extra new lines at end

type StrictGrader struct{}

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
			readRuneError(err, expectingEnd, done)
			break
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
		if ansIndex >= len(ans)-1 || (ans[len(ans)-1] == '\n' && ansIndex >= len(ans)-2) {
			expectingEnd = true
		}

		ansIndex++
	}
}

// ***** EndTrim Grader *****
// Ignores whitespace at the end of a line, and new lines characters at the end

type EndTrimGrader struct{}

func (grader EndTrimGrader) CompareStream(session *GradeSession, expectedAnswerFile *os.File, done chan CaseReturn) {
	// move index for reading the file to beginning
	_, _ = session.OutputStream.Seek(0, 0)

	// read from output stream and answer stream simultaneously per line
	outputScan := bufio.NewScanner(session.OutputStream)
	answerScan := bufio.NewScanner(expectedAnswerFile)

	outputHasNext := outputScan.Scan()
	answerHasNext := answerScan.Scan()

	for outputHasNext && answerHasNext {
		outLine := strings.ReplaceAll(strings.TrimRight(outputScan.Text(), " "), "\r", "")
		ansLine := strings.ReplaceAll(strings.TrimRight(answerScan.Text(), " "), "\r", "")

		if outLine != ansLine {
			shared.Debug("compare: " + outLine + " and " + ansLine)
			done <- CaseReturn{
				Result:     shared.OUTCOME_WA,
				ResultInfo: "Wrong char",
			}
			return
		}

		outputHasNext = outputScan.Scan()
		answerHasNext = answerScan.Scan()
	}

	// check trailing characters
	if outputHasNext {
		first := true
		for first || outputScan.Scan() {
			text := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(outputScan.Text(), " ", ""), "\n", ""), "\r", "")
			if text != "" {
				done <- CaseReturn{
					Result:     shared.OUTCOME_WA,
					ResultInfo: "Wrong char",
				}
				return
			}
			first = false
		}
	}

	// check if answer is empty
	if answerHasNext {
		first := true
		for first || answerScan.Scan() {
			text := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(answerScan.Text(), " ", ""), "\n", ""), "\r", "")
			if text != "" {
				done <- CaseReturn{
					Result:     shared.OUTCOME_WA,
					ResultInfo: "Ended early",
				}
				return
			}
			first = false
		}
	}

	// AC if it completes successfully
	done <- CaseReturn{
		Result: shared.OUTCOME_AC,
	}
}

// ***** Field Grader *****
// Judges submissions word by word

type FieldGrader struct{}

func (grader FieldGrader) CompareStream(session *GradeSession, expectedAnswerFile *os.File, done chan CaseReturn) {
	// move index for reading the file to beginning
	_, _ = session.OutputStream.Seek(0, 0)
	
	// Lists for words
	outputList := list.New()
	answerList := list.New()
	outputLength := 0
	answerLength := 0
	
	// read from output stream
	outputScan := bufio.NewScanner(session.OutputStream)
	for outputScan.Scan() {
		outputs := strings.Fields(outputScan.Text())
		for _, word := range outputs {
			outputList.PushBack(word)
			outputLength++
		}
	}
	
	// read from answer stream
	answerScan := bufio.NewScanner(expectedAnswerFile)
	for answerScan.Scan() {
		answers := strings.Fields(answerScan.Text())
		for _, word := range answers {
			answerList.PushBack(word)
			answerLength++
		}
	}
	
	// Diff # words is bad
	if answerLength != outputLength {
		done <- CaseReturn{
			Result:     shared.OUTCOME_WA,
			ResultInfo: "Different number of words",
		}
		return
	}
	
	// Go through output and answer lists
	outputIterator := outputList.Front()
	answerIterator := answerList.Front()
	for outputIterator != nil {
		outputWord := (outputIterator.Value).(string)
		answerWord := (answerIterator.Value).(string)
		if outputWord != answerWord {
			done <- CaseReturn{
				Result:     shared.OUTCOME_WA,
				ResultInfo: "Wrong char",
			}
			return
		}
		// advance iterator
		outputIterator = outputIterator.Next()
		answerIterator = answerIterator.Next()
	}
	
	// AC if it completes successfully
	done <- CaseReturn{
		Result: shared.OUTCOME_AC,
	}
}

// ***** Double Grader *****
// Field Grader but with double uncertainty of 0.000005

type DoubleGrader struct{}

func (grader DoubleGrader) CompareStream(session *GradeSession, expectedAnswerFile *os.File, done chan CaseReturn) {
	// move index for reading the file to beginning
	_, _ = session.OutputStream.Seek(0, 0)
	
	//Lists for words
	outputList := list.New()
	answerList := list.New()
	outputLength := 0
	answerLength := 0
	
	// read from output stream
	outputScan := bufio.NewScanner(session.OutputStream)
	for outputScan.Scan() {
		outputs := strings.Fields(outputScan.Text())
		for _, word := range outputs {
			outputList.PushBack(word)
			outputLength++
		}
	}
	
	// read from answer stream
	answerScan := bufio.NewScanner(expectedAnswerFile)
	for answerScan.Scan() {
		answers := strings.Fields(answerScan.Text())
		for _, word := range answers {
			answerList.PushBack(word)
			answerLength++
		}
	}
	
	// Diff # words is bad
	if outputLength != answerLength {
		done <- CaseReturn{
			Result:     shared.OUTCOME_WA,
			ResultInfo: "Different number of words",
		}
		return
	}
	
	// Go through output and answer lists
	outputIterator := outputList.Front()
	answerIterator := answerList.Front()
	for outputIterator != nil {
		outputWord := (outputIterator.Value).(string)
		answerWord := (answerIterator.Value).(string)
		// not correct word
		if outputWord != answerWord {
			// parse floats and check if outputs really are floats
			outputFloat, outputFloatError := strconv.ParseFloat(outputWord, 64)
			answerFloat, answerFloatError := strconv.ParseFloat(answerWord, 64)
			outputIsFloat := outputFloatError == nil
			answerIsFloat := answerFloatError == nil

			// answer's not a float
			if !answerIsFloat {
				done <- CaseReturn{
					Result:     shared.OUTCOME_WA,
					ResultInfo: "Wrong char",
				}
				return
			}
			// output's not a float
			if !outputIsFloat {
				done <- CaseReturn{
					Result:     shared.OUTCOME_WA,
					ResultInfo: "Wrong char",
				}
				return
			}
			// they are both floats
			max := answerFloat + 0.000005
			min := answerFloat - 0.000005
			// not within the margin of error >:(
			if outputFloat > max || outputFloat < min {
				done <- CaseReturn{
					Result:     shared.OUTCOME_WA,
					ResultInfo: "Wrong double",
				}
				return
			}
		}
		// advance iterator
		outputIterator = outputIterator.Next()
		answerIterator = answerIterator.Next()
	}
	
	// AC if it completes successfully
	done <- CaseReturn{
		Result: shared.OUTCOME_AC,
	}
}

func readRuneError(err error, expectingEnd bool, done chan CaseReturn) {
	shared.Debug("readrune: " + err.Error())
	if expectingEnd { // expected no more text
		done <- CaseReturn{
			Result: shared.OUTCOME_AC,
		}
	} else { // did not finish giving full answer
		done <- CaseReturn{
			Result:     shared.OUTCOME_WA,
			ResultInfo: "Ended early",
		}
	}
}
