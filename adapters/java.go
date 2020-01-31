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

	//c := exec.Command("java", "-XX:MaxHeapSize=" + memLim + "m", "-XX:InitialHeapSize=" + memLim + "m", "-XX:CompressedClassSpaceSize=" + memLim + "m", "-XX:MaxMetaspaceSize=" + memLim + "m", "Main")
	c := exec.Command("java", "-Xmx" + strconv.Itoa(int(session.Limit.Memory/1e3)) + "K", "-Xss128m", "-XX:+UseSerialGC", "-XX:ErrorFile=crash.log", "-XX:MaxMetaspaceSize=128m", "Main")
	c.Env = append(c.Env, "MALLOC_ARENAS_MAX=1")
	c.Dir = session.Workspace

	// set memory limit to zero since it's enforced by the jvm
	session.Limit.Memory = 0

	return c, nil
}