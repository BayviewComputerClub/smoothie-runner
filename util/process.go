package util

import (
	"golang.org/x/sys/unix"
	"os"
)

func IsPidRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	} else {
		err := process.Signal(unix.Signal(0))
		return err == nil || (err.Error() != "no such process" && err.Error() != "os: process already finished")
	}
}

func BlockRestrictedCalls() {

}