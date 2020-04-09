package adapters

import (
	"errors"
	"github.com/BayviewComputerClub/smoothie-runner/sandbox"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"
)

func CCompileHelper(session *shared.JudgeSession, compileCmd *exec.Cmd, fileName string) error {
	// setup compiler profile for c/c++
	se := util.SANDBOX_COMPILER_PROFILE
	se.AllowWrite = make(map[string]bool)
	for k, v := range util.SANDBOX_COMPILER_PROFILE.AllowWrite {
		se.AllowWrite[k] = v
	}
	se.AllowWrite[session.Workspace] = true
	se.AllowWrite[fileName] = true

	// compile
	rsr, err := sandboxCompileHelper(compileCmd, &sandbox.RunnerSession{SeccompProfile: se, SandboxWithSeccomp: false})
	if err != nil {
		return err
	}

	// read stdout and stderr from compile (truncate at 4096 bytes to not make it too long)
	dat := make([]byte, 4096)
	f, err := os.Open(session.Workspace + "/compileout")
	if err != nil {
		return err
	}
	io.ReadFull(f, dat)

	// fix utf8 (for grpc)
	errstr := strings.Map(func(r rune) rune {
		if r == utf8.RuneError {
			return -1
		}
		return r
	}, string(dat))

	// send error message
	if rsr.Status != sandbox.RunnerStatusOK || rsr.ExitCode != 0 {
		return errors.New(strings.ReplaceAll(errstr, session.Workspace + "/" + fileName, "") + " : " + rsr.Error)
	}
	return nil
}