package adapters

import (
	"errors"
	"github.com/BayviewComputerClub/smoothie-runner/sandbox"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
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

	se := util.SANDBOX_COMPILER_PROFILE
	se.AllowWrite = make(map[string]bool)
	for k, v := range util.SANDBOX_COMPILER_PROFILE.AllowWrite {
		se.AllowWrite[k] = v
	}
	se.AllowWrite[session.Workspace] = true
	se.AllowWrite["main.cpp"] = true

	rsr, err := sandboxCompileHelper(compileCmd, &sandbox.RunnerSession{SeccompProfile: se, SandboxWithSeccomp: false})
	if err != nil {
		return nil, err
	}

	// read stdout and stderr from compile (truncate at 4096 bytes to not make it too long)
	dat := make([]byte, 4096)
	f, err := os.Open(session.Workspace + "/compileout")
	if err != nil {
		return nil, err
	}
	io.ReadFull(f, dat)

	if rsr.Status != sandbox.RunnerStatusOK || rsr.ExitCode != 0 {
		return nil, errors.New(strings.ReplaceAll(string(dat), session.Workspace+"/main.cpp", "") + " : " + rsr.Error)
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

// c++11

type Cpp11Adapter struct{}

func (adapter Cpp11Adapter) GetName() string {
	return "c++11"
}

func (adapter Cpp11Adapter) Compile(session *shared.JudgeSession) (*exec.Cmd, error) {
	return CppHelper(session, "gnu++11")
}

// c++14

type Cpp14Adapter struct{}

func (adapter Cpp14Adapter) GetName() string {
	return "c++14"
}

func (adapter Cpp14Adapter) Compile(session *shared.JudgeSession) (*exec.Cmd, error) {
	return CppHelper(session, "gnu++14")
}

// c++17

type Cpp17Adapter struct{}

func (adapter Cpp17Adapter) GetName() string {
	return "c++17"
}

func (adapter Cpp17Adapter) Compile(session *shared.JudgeSession) (*exec.Cmd, error) {
	return CppHelper(session, "gnu++17")
}

