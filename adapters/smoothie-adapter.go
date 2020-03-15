package adapters

import (
	"errors"
	"github.com/BayviewComputerClub/smoothie-runner/sandbox"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"os/exec"
	"time"
)

type SmoothieAdapter interface {
	GetName() string
	Compile(session *shared.JudgeSession) (*exec.Cmd, error) // return command in the workspace
}

var (
	adapters = make(map[string]SmoothieAdapter)
)

func init() {
	adapters["c++98"] = Cpp98Adapter{}
	adapters["c++11"] = Cpp11Adapter{}
	adapters["c++14"] = Cpp14Adapter{}
	adapters["c++17"] = Cpp17Adapter{}
	adapters["c11"] = C11Adapter{}
	adapters["java11"] = Java11Adapter{}
	adapters["python3"] = Python3Adapter{}
}

func CompileAndGetRunCommand(session *shared.JudgeSession) (*exec.Cmd, error) {
	if adapters[session.Language] == nil {
		return nil, errors.New("language not supported (" + session.Language + ")")
	}

	return adapters[session.Language].Compile(session)
}

// run compile command with sandbox
func sandboxCompileHelper(compileCommand *exec.Cmd, sandboxProfile util.SandboxProfile) (*sandbox.RunnerSessionResult, error) {
	session := sandbox.RunnerSession{
		ResultChan:         make(chan sandbox.RunnerSessionResult),
		InternalResultChan: make(chan sandbox.RunnerResult),
		ExecFile:           0,
		ExecArgs:           compileCommand.Args,
		ExecEnv:            compileCommand.Env,
		ExecUsed:           false,
		Files:              make(map[int]uintptr),
		Workspace:          compileCommand.Dir,
		RLimits:            nil,
		TimeLimit:          30 * time.Second, // set static compile time limit to 30 seconds
		MemoryLimit:        1e9, // set static compile memory limit to 1GB
		SeccompProfile:     sandboxProfile,
	}

	f, err := util.GetPtrsFromCmd(compileCommand)
	if err != nil {
		return nil, err
	}
	session.ExecFile = f.Fd()
	defer f.Close()

	go session.Start()
	res := <-session.ResultChan
	return &res, nil
}