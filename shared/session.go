package shared

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"os"
)

type JudgeSession struct {
	Workspace string
	Code string
	Language string
	OriginalRequest *pb.TestSolutionRequest

	// streams during judging
	OutputBuffer *os.File
	ErrorBuffer *os.File
	OutputStream *os.File
	ErrorStream *os.File
	InputStream *os.File

	// per judge session
	Stderr string // error dumped here
	ExitCode int
}

func (session *JudgeSession) CloseStreams() {
	if session.OutputStream != nil {
		session.OutputStream.Close()
	}
	if session.OutputBuffer != nil {
		session.OutputBuffer.Close()
	}
	if session.ErrorStream != nil {
		session.ErrorStream.Close()
	}
	if session.ErrorBuffer != nil {
		session.ErrorBuffer.Close()
	}
	if session.InputStream != nil {
		session.InputStream.Close()
	}
}