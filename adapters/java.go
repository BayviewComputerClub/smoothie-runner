package adapters

import (
	"errors"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

type Java11Adapter struct {}

func (adapter Java11Adapter) GetName() string {
	return "java11"
}

func (adapter Java11Adapter) Compile(session *shared.JudgeSession) (*exec.Cmd, error) {
	err := ioutil.WriteFile(session.Workspace + "/Main.java", []byte(session.Code), 0644)
	if err != nil {
		return nil, err
	}

	output, err := exec.Command("javac", session.Workspace+"/Main.java").CombinedOutput()
	if err != nil {
		return nil, errors.New(strings.ReplaceAll(string(output), session.Workspace+"/Main.java", ""))
	}

	c := exec.Command("java", "-Xmx" + strconv.Itoa(int(session.OriginalRequest.Problem.MemLimit)) + "M", "-Xss128m", "-XX:+UseSerialGC", "-XX:ErrorFile=crash.log", "-XX:MaxMetaspaceSize=128m", "Main")
	c.Env = append(c.Env, "MALLOC_ARENAS_MAX=1")
	c.Dir = session.Workspace

	// set memory limit to zero since it's enforced by the jvm
	session.OriginalRequest.Problem.MemLimit = 0

	return c, nil
}

// TODO scan stderr for outofmemoryexception and turn that into mle