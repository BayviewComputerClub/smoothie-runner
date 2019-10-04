package adapters

import "os/exec"

type Java11Adapter struct {}

func (adapter Java11Adapter) GetName() string {
	return "java11"
}

func (adapter Java11Adapter) Compile(code string) (*exec.Cmd, error) {
	return nil, nil
}