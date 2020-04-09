package adapters

import (
	"errors"
	"github.com/BayviewComputerClub/smoothie-runner/sandbox"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)



var (
	adapters = make(map[string]shared.SmoothieAdapter)
)

func init() {
	adapters["c++98"] = Cpp98Adapter{}
	adapters["c++11"] = Cpp11Adapter{}
	adapters["c++14"] = Cpp14Adapter{}
	adapters["c++17"] = Cpp17Adapter{}
	adapters["c11"] = C11Adapter{}
	adapters["java11"] = Java11Adapter{}
	adapters["python3"] = Python3Adapter{}
}

func CompileAndGetRunCommand(session *shared.JudgeSession) (*exec.Cmd, error) {
	if adapters[session.Language] == nil {
		return nil, errors.New("language not supported (" + session.Language + ")")
	}

	// set language adapter
	session.LanguageAdapter = adapters[session.Language]
	return session.LanguageAdapter.Compile(session)
}

// run compile command with sandbox
func sandboxCompileHelper(compileCommand *exec.Cmd, session *sandbox.RunnerSession) (*sandbox.RunnerSessionResult, error) {
	session.ResultChan = make(chan sandbox.RunnerSessionResult)
	session.InternalResultChan = make(chan sandbox.RunnerResult)
	session.ExecArgs = compileCommand.Args
	session.ExecEnv = compileCommand.Env
	session.Files = make(map[int]uintptr)
	session.Workspace = compileCommand.Dir
	session.RLimits = nil
	session.HardTimeout = 30 * time.Second
	session.TimeLimit = 20 * time.Second // set compile time limit to 30 seconds
	session.MemoryLimit = 1e9            // set compile memory limit to 1GB
	session.FSizeLimit = 1e9             // set maximum write to 1GB
	session.NProcLimit = -1              // no process limit

	// create output file
	err := ioutil.WriteFile(compileCommand.Dir+"/compileout", []byte(""), 0644)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(compileCommand.Dir+"/compileout", os.O_RDWR, os.ModeAppend) // open with read/write fd
	if err != nil {
		return nil, err
	}
	session.Files[1] = file.Fd()
	session.Files[2] = file.Fd()

	// add exec file ptr
	f, err := util.GetPtrsFromCmd(compileCommand)
	if err != nil {
		return nil, err
	}
	session.ExecFile = f.Fd()
	defer f.Close()

	go session.Start()
	res := <-session.ResultChan
	return &res, nil
}
