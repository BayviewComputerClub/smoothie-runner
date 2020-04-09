package adapters

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"io/ioutil"
	"os/exec"
)

func CppHelper(session *shared.JudgeSession, std string) (*exec.Cmd, error) {
	err := ioutil.WriteFile(session.Workspace + "/main.cpp", []byte(session.Code), 0644)
	if err != nil {
		return nil, err
	}

	// compile
	compileCmd := exec.Command("/usr/bin/g++", "-std=" + std, "main.cpp", "-o", "main")
	compileCmd.Dir = session.Workspace
	compileCmd.Env = append(compileCmd.Env, "PATH=$PATH:/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin")

	err = CCompileHelper(session, compileCmd, session.Workspace + "/main.cpp")
	if err != nil {
		return nil, err
	}

	// return exec command
	c := exec.Command(session.Workspace+"/main")
	c.Dir = session.Workspace
	return c, nil
}

// c++98

type Cpp98Adapter struct{}

func (adapter Cpp98Adapter) GetName() string {
	return "c++98"
}

func (adapter Cpp98Adapter) Compile(session *shared.JudgeSession) (*exec.Cmd, error) {
	return CppHelper(session, "gnu++98")
}

func (adapter Cpp98Adapter) JudgeFinished(tcr *pb.TestCaseResult) {}

// c++11

type Cpp11Adapter struct{}

func (adapter Cpp11Adapter) GetName() string {
	return "c++11"
}

func (adapter Cpp11Adapter) Compile(session *shared.JudgeSession) (*exec.Cmd, error) {
	return CppHelper(session, "gnu++11")
}

func (adapter Cpp11Adapter) JudgeFinished(tcr *pb.TestCaseResult) {}

// c++14

type Cpp14Adapter struct{}

func (adapter Cpp14Adapter) GetName() string {
	return "c++14"
}

func (adapter Cpp14Adapter) Compile(session *shared.JudgeSession) (*exec.Cmd, error) {
	return CppHelper(session, "gnu++14")
}

func (adapter Cpp14Adapter) JudgeFinished(tcr *pb.TestCaseResult) {}

// c++17

type Cpp17Adapter struct{}

func (adapter Cpp17Adapter) GetName() string {
	return "c++17"
}

func (adapter Cpp17Adapter) Compile(session *shared.JudgeSession) (*exec.Cmd, error) {
	return CppHelper(session, "gnu++17")
}

func (adapter Cpp17Adapter) JudgeFinished(tcr *pb.TestCaseResult) {}
