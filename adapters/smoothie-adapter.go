package adapters

import (
	"errors"
	"os/exec"
)

type SmoothieAdapter interface {
	GetName() string
	Compile(string) (exec.Cmd, error)
}

var (
	adapters = make(map[string]SmoothieAdapter)
)

func init() {
	adapters["c++11"] = Cpp11Adapter{}
	adapters["c11"] = C11Adapter{}
	adapters["java11"] = Java11Adapter{}
}

func CompileAndGetRunCommand(language string, code string) (exec.Cmd, error) {
	if adapters[language] == nil {
		return "", errors.New("language not supported")
	}

	return adapters[language].Compile(code)
}