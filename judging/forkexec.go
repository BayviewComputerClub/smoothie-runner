package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"syscall"
)

func (tracer *PTracer) ForkExec() {
	pid, _, err := unix.Syscall6(syscall.SYS_CLONE, uintptr(unix.SIGCHLD), 0, 0, 0, 0, 0)
	if err != 0 || pid != 0 {
		if err != 0 {
			util.Warn("forkexec clone: " + err.Error())
			tracer.StreamDone <- CaseReturn{Result: shared.OUTCOME_ISE, ResultInfo: err.Error()}
			return
		}

		tracer.Pid = pid
	}
}