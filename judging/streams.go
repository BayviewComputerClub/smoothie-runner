package judging

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

func createStreamFile(loc string) (*os.File, error) {
	err := ioutil.WriteFile(loc, []byte(""), 0644)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(loc, os.O_RDWR, os.ModeAppend) // open with read/write fd
	if err != nil {
		return nil, err
	}
	return file, nil
}

/*
 * Initialize all streams before program starts
 */

func (session *GradeSession) InitIOFiles() {
	name := strconv.FormatInt(time.Now().UnixNano(), 10)

	// os pipes are nice but have a size cap soo

	outputFileLoc := session.JudgingSession.Workspace + "/" + name + ".out"
	errFileLoc := session.JudgingSession.Workspace + "/" + name + ".err"

	var err error

	// create empty file for output
	session.OutputStream, err = createStreamFile(outputFileLoc)
	if err != nil {
		util.Warn("outputstream: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: ""}
		return
	}

	// create empty file for errors
	session.ErrorStream, err = createStreamFile(errFileLoc)
	if err != nil {
		util.Warn("errorstream: " + err.Error())
		session.StreamResult <- pb.TestCaseResult{Result: shared.OUTCOME_ISE, ResultInfo: ""}
		return
	}

	// use opened input file from cache
	session.InputStream = session.CurrentCase.Input
}

/*
 * Cleanup streams
 */

func (session *GradeSession) CloseStreams() {
	// remove output and error files (not input, since that is cached)
	if session.OutputStream != nil {
		session.OutputStream.Close()
		os.Remove(session.OutputStream.Name())
	}
	if session.ErrorStream != nil {
		session.ErrorStream.Close()
		os.Remove(session.ErrorStream.Name())
	}
	if session.InputStream != nil {
		session.InputStream.Close()
	}
}
