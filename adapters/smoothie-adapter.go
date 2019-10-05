package adapters

import (
	"errors"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"os/exec"
)

type SmoothieAdapter interface {
	GetName() string
	Compile(session shared.JudgeSession) (*exec.Cmd, error)
}

var (
	adapters = make(map[string]SmoothieAdapter)
)

func init() {
	adapters["c++11"] = Cpp11Adapter{}
	adapters["c11"] = C11Adapter{}
	adapters["java11"] = Java11Adapter{}
}

func CompileAndGetRunCommand(session shared.JudgeSession) (*exec.Cmd, error) {
	if adapters[session.Language] == nil {
		return nil, errors.New("language not supported")
	}

	return adapters[session.Language].Compile(session)
}