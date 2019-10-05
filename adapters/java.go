package adapters

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"os/exec"
)

type Java11Adapter struct {}

func (adapter Java11Adapter) GetName() string {
	return "java11"
}

func (adapter Java11Adapter) Compile(session shared.JudgeSession) (*exec.Cmd, error) {
	return nil, nil
}