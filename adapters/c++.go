package adapters

import (
	"errors"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"io/ioutil"
	"os/exec"
	"strings"
)

func CppHelper(session shared.JudgeSession, std string) (*exec.Cmd, error) {

	err := ioutil.WriteFile(session.Workspace + "/main.cpp", []byte(session.Code), 0644)
	if err != nil {
		return nil, err
	}

	output, err := exec.Command("g++", "-std=" + std, session.Workspace+"/main.cpp", "-o", session.Workspace+"/main").CombinedOutput()
	if err != nil {
		return nil, errors.New(strings.ReplaceAll(string(output), session.Workspace+"/main.cpp", ""))
	}

	c := exec.Command(session.Workspace+"/main")
	c.Dir = session.Workspace
	return c, nil
}

// c++98

type Cpp98Adapter struct{}

func (adapter Cpp98Adapter) GetName() string {
	return "c++98"
}

func (adapter Cpp98Adapter) Compile(session shared.JudgeSession) (*exec.Cmd, error) {
	return CppHelper(session, "gnu++98")
}

// c++11

type Cpp11Adapter struct{}

func (adapter Cpp11Adapter) GetName() string {
	return "c++11"
}

func (adapter Cpp11Adapter) Compile(session shared.JudgeSession) (*exec.Cmd, error) {
	return CppHelper(session, "gnu++11")
}

// c++14

type Cpp14Adapter struct{}

func (adapter Cpp14Adapter) GetName() string {
	return "c++14"
}

func (adapter Cpp14Adapter) Compile(session shared.JudgeSession) (*exec.Cmd, error) {
	return CppHelper(session, "gnu++14")
}

// c++17

type Cpp17Adapter struct{}

func (adapter Cpp17Adapter) GetName() string {
	return "c++17"
}

func (adapter Cpp17Adapter) Compile(session shared.JudgeSession) (*exec.Cmd, error) {
	return CppHelper(session, "gnu++17")
}

