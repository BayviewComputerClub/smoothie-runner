package main

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"os"
	"os/exec"
	"runtime"
)

/*
 * Tasks:
 * setrlimit to prevent fork() in processes (use golang.org/x/sys)
 * set nice value to be low
 * run process as noperm user (or just run the whole program with no perm)
 * use ptrace unix calls
 * can use syscall credentials
 * make sure thread is locked
 * set timeout
 */

func judgeCase(compiledProgram os.File, batchCase pb.ProblemBatchCase) pb.TestCaseResult {
	c := exec.Command("./" + compiledProgram.Name())

	stderrPipe, err := c.StderrPipe()
	if err != nil {
		warn(err.Error())
		return
	}
	stdoutPipe, err := c.StdoutPipe()
	stdinPipe, err := c.StdinPipe()

	err = c.Start()

}