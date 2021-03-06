package adapters

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"io/ioutil"
	"os/exec"
)

type Python3Adapter struct {}

func (adapter Python3Adapter) GetName() string {
	return "python3"
}

func (adapter Python3Adapter) JudgeFinished(tcr *pb.TestCaseResult) {}

func (adapter Python3Adapter) Compile(session *shared.JudgeSession) (*exec.Cmd, error) {
	err := ioutil.WriteFile(session.Workspace + "/main.py", []byte(session.Code), 0644)
	if err != nil {
		return nil, err
	}

	c := exec.Command("python3", "-BSu", "main.py")
	c.Dir = session.Workspace
	c.Env = append(c.Env, "")
	return c, nil
}