package adapters

import (
	"errors"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"io/ioutil"
	"os/exec"
	"strings"
)

type Cpp11Adapter struct{}

func (adapter Cpp11Adapter) GetName() string {
	return "c++11"
}

func (adapter Cpp11Adapter) Compile(session shared.JudgeSession) (*exec.Cmd, error) {

	err := ioutil.WriteFile(session.Workspace + "/main.cpp", []byte(session.Code), 0644)
	if err != nil {
		return nil, err
	}

	output, err := exec.Command("g++", "-std=c++11", session.Workspace+"/main.cpp", "-o", session.Workspace+"/main").CombinedOutput()
	if err != nil {
		return nil, errors.New(strings.ReplaceAll(string(output), session.Workspace+"/main.cpp", ""))
	}

	return exec.Command(session.Workspace + "/main"), nil
}
