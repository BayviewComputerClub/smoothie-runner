package adapters

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Cpp11Adapter struct {}

func (adapter Cpp11Adapter) GetName() string {
	return "c++11"
}

func (adapter Cpp11Adapter) Compile(code string) (*exec.Cmd, error) {

	curTime := strconv.FormatInt(time.Now().Unix(), 10)
	err := ioutil.WriteFile(shared.TESTING_DIR + "/" + curTime + ".cpp", []byte(code), 0644)
	if err != nil {
		return nil, err
	}

	c := exec.Command("g++", "-std=c++11", shared.TESTING_DIR + "/" + curTime + ".cpp", "-o " + curTime)
	err = c.Run()
	if err != nil {
		return nil, err
	}

	err = os.Remove(shared.TESTING_DIR + "/" + curTime + ".cpp")
	if err != nil {
		return nil, err
	}

	return exec.Command(shared.TESTING_DIR + "/" + curTime), nil
}