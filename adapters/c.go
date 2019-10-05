package adapters

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"os/exec"
)

type C11Adapter struct {}

func (adapter C11Adapter) GetName() string {
	return "c11"
}

func (adapter C11Adapter) Compile(session shared.JudgeSession) (*exec.Cmd, error) {
	return nil, nil
}
