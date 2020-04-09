package adapters

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"io/ioutil"
	"os/exec"
)

type C11Adapter struct {}

func (adapter C11Adapter) GetName() string {
	return "c11"
}

func (adapter C11Adapter) JudgeFinished(tcr *pb.TestCaseResult) {}

func (adapter C11Adapter) Compile(session *shared.JudgeSession) (*exec.Cmd, error) {
	err := ioutil.WriteFile(session.Workspace + "/main.c", []byte(session.Code), 0644)
	if err != nil {
		return nil, err
	}

	// compile
	compileCmd := exec.Command("/usr/bin/gcc", "-std=c11", "main.c", "-o", "main")
	compileCmd.Dir = session.Workspace
	compileCmd.Env = append(compileCmd.Env, "PATH=$PATH:/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin")

	err = CCompileHelper(session, compileCmd, session.Workspace + "/main.c")
	if err != nil {
		return nil, err
	}

	c := exec.Command(session.Workspace+"/main")
	c.Dir = session.Workspace
	return c, nil
}
