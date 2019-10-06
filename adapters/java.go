package adapters

import (
	"errors"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"io/ioutil"
	"os/exec"
	"strings"
)

type Java11Adapter struct {}

func (adapter Java11Adapter) GetName() string {
	return "java11"
}

func (adapter Java11Adapter) Compile(session shared.JudgeSession) (*exec.Cmd, error) {
	err := ioutil.WriteFile(session.Workspace + "/Main.java", []byte(session.Code), 0644)
	if err != nil {
		return nil, err
	}

	output, err := exec.Command("javac", session.Workspace+"/Main.java").CombinedOutput()
	if err != nil {
		return nil, errors.New(strings.ReplaceAll(string(output), session.Workspace+"/Main.java", ""))
	}

	c := exec.Command("/usr/bin/java", "Main")
	c.Dir = session.Workspace
	return c, nil
}