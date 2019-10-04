package adapters

import "os/exec"

type Cpp11Adapter struct {}

func (adapter Cpp11Adapter) GetName() string {
	return "c++11"
}

func (adapter Cpp11Adapter) Compile(code string) (*exec.Cmd, error) {
	return nil, nil
}