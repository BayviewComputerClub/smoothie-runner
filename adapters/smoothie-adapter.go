package adapters

import (
	"errors"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"os/exec"
)

type SmoothieAdapter interface {
	GetName() string
	Compile(session shared.JudgeSession) (*exec.Cmd, error) // return command in the workspace
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

func CompileAndGetRunCommand(session shared.JudgeSession) (*exec.Cmd, error) {
	if adapters[session.Language] == nil {
		return nil, errors.New("language not supported")
	}

	return adapters[session.Language].Compile(session)
}