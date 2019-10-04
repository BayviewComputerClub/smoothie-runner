package adapters

import "os/exec"

type C11Adapter struct {}

func (adapter C11Adapter) GetName() string {
	return "c11"
}

func (adapter C11Adapter) Compile(code string) (*exec.Cmd, error) {
	return nil, nil
}
