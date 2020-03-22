package sandbox

import (
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"math"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	usePeekRead = false
)

// if a file is in the map or sub directories in the map
func isFileInSet(file string, s map[string]bool) bool {
	if s[file] {
		return true
	}

	spl := strings.Split(file, "/")
	comp := ""
	for i, sp := range spl {
		if i == 1 {
			comp += sp
		} else {
			comp += "/" + sp
		}

		if s[comp] {
			return true
		}
	}

	return false
}

// check for symlinks in file paths
func getFileRealPath(p string) string {
	f, err := filepath.EvalSymlinks(p)
	if err != nil {
		return ""
	}
	return f
}

// check if the file is in the set
func fileCheck(name string, s map[string]bool) bool {
	name = path.Clean(name)
	return isFileInSet(name, s) || isFileInSet(getFileRealPath(name), s)
}

// get registers
func getPregs(pid int) (*unix.PtraceRegs, error) {
	pregs := unix.PtraceRegs{}
	err := unix.PtraceGetRegs(pid, &pregs)
	if err != nil {
		return nil, err
	}
	return &pregs, nil
}

// change the syscall
func sandboxChangeCall(pregs *unix.PtraceRegs, pid int, call uint64) {
	pregs.Orig_rax = call
	err := unix.PtraceSetRegs(pid, pregs)
	if err != nil {
		util.Warn("ptracesetregs: " + err.Error())
	}
}

// block the syscall
func blockCall(pregs *unix.PtraceRegs, pid int) {
	shared.Debug(fmt.Sprintf("Blocked: %v", pregs.Orig_rax))
	sandboxChangeCall(pregs, pid, uint64(math.Inf(0)))
	pregs.Rax = uint64(math.Inf(0))
}

// read file names
func readStringAtAddr(pid int, address uintptr) (string, error) {
	var (
		s string
		err error
	)
	/*if usePeekRead { switch to just using process_vm_readv if possible at all times
		s, err = util.ReadPeekString(pid, address)
	} else {*/
		s, err = util.ProcessVmReadVStr(pid, address)
		if err != nil {
			if no, ok := err.(unix.Errno); ok {
				if no == unix.ENOSYS {
					s, err = util.ReadPeekString(pid, address)
					usePeekRead = true
					util.Warn("Unable to use process_vm_readv, switching to ptrace peek read.")
				}
			}
		//}
	}
	return s, err
}

/* -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=- */

func (session *RunnerSession) TraceCheckWrite(pid int, name string, pregs *unix.PtraceRegs) {
	shared.Debug("Checking file write access: " + name)
	if !fileCheck(name, session.SeccompProfile.AllowWrite) {
		blockCall(pregs, pid)
	}
}

func (session *RunnerSession) TraceCheckRead(pid int, name string, pregs *unix.PtraceRegs) {
	shared.Debug("Checking file read access: " + name)
	if !fileCheck(name, session.SeccompProfile.AllowWrite) && !fileCheck(name, session.SeccompProfile.AllowRead) {
		blockCall(pregs, pid)
	}
}

func (session *RunnerSession) TraceCheckStat(pid int, name string, pregs *unix.PtraceRegs) {
	shared.Debug("Checking file stat access: " + name)
	session.TraceCheckRead(pid, name, pregs)
}

func (session *RunnerSession) TraceCheckOpen(pid int, name string, flags uint64, pregs *unix.PtraceRegs) {
	isReadOnly := (flags&unix.O_ACCMODE == unix.O_RDONLY) && (flags&unix.O_CREAT == 0) && (flags&unix.O_EXCL == 0) && (flags&unix.O_TRUNC == 0)
	if isReadOnly {
		session.TraceCheckRead(pid, name, pregs)
	} else {
		session.TraceCheckWrite(pid, name, pregs)
	}
}

// restrict call if necessary
func (session *RunnerSession) CheckRestrictedCall(pid int, pregs *unix.PtraceRegs) {
	shared.Debug("Checking syscall: " + strconv.Itoa(int(pregs.Orig_rax)))

	var (
		err error
		wd  string
	)

	// https://github.com/criyle/go-sandbox/blob/d1ed5f0f21ddb6a472d200fa32bfdb10c8d6a466/runner/ptrace/handle.go#L61
	switch int(pregs.Orig_rax) {
	case unix.SYS_OPEN:
		wd, err = readStringAtAddr(pid, uintptr(pregs.Rdi))
		if err == nil {
			session.TraceCheckOpen(pid, wd, pregs.Rsi, pregs)
		}

	case unix.SYS_OPENAT:
		wd, err = readStringAtAddr(pid, uintptr(pregs.Rsi))
		if err == nil {
			session.TraceCheckOpen(pid, wd, pregs.Rdx, pregs)
		}

	case unix.SYS_READLINK:
		wd, err = readStringAtAddr(pid, uintptr(pregs.Rdi))
		if err == nil {
			session.TraceCheckRead(pid, wd, pregs)
		}

	case unix.SYS_READLINKAT:
		wd, err = readStringAtAddr(pid, uintptr(pregs.Rsi))
		if err == nil {
			session.TraceCheckRead(pid, wd, pregs)
		}

	case unix.SYS_UNLINK, unix.SYS_CHMOD, unix.SYS_RENAME:
		wd, err = readStringAtAddr(pid, uintptr(pregs.Rdi))
		if err == nil {
			session.TraceCheckWrite(pid, wd, pregs)
		}

	case unix.SYS_UNLINKAT:
		wd, err = readStringAtAddr(pid, uintptr(pregs.Rsi))
		if err == nil {
			session.TraceCheckWrite(pid, wd, pregs)
		}

	case unix.SYS_ACCESS, unix.SYS_STAT, unix.SYS_LSTAT:
		wd, err = readStringAtAddr(pid, uintptr(pregs.Rdi))
		if err == nil {
			session.TraceCheckStat(pid, wd, pregs)
		}

	case unix.SYS_FACCESSAT, unix.SYS_NEWFSTATAT:
		wd, err = readStringAtAddr(pid, uintptr(pregs.Rsi))
		if err == nil {
			session.TraceCheckStat(pid, wd, pregs)
		}

	case unix.SYS_EXECVE:
		if !session.ExecUsed { // on first execveat to run program, ignore call
			session.ExecUsed = true
		} else {
			blockCall(pregs, pid)
		}
		//wd, err = readStringAtAddr(pid, uintptr(pregs.Rdi))
		//if err == nil {
		//	session.TraceCheckRead(pid, wd, pregs)
		//}

	case unix.SYS_EXECVEAT:
		if !session.ExecUsed { // on first execveat to run program, ignore call
			session.ExecUsed = true
		} else {
			blockCall(pregs, pid)
		}
		//wd, err = readStringAtAddr(pid, uintptr(pregs.Rsi))
		//session.TraceCheckRead(pid, wd, pregs)

	default:
		// ban by default (allowed calls should have been allowed by seccomp)
		blockCall(pregs, pid)
	}

	if err != nil {
		util.Warn("readstring: " + err.Error())
		return
	}
}

func (session *RunnerSession) handleTrap(pid int) error {
	pregs, err := getPregs(pid)
	if err != nil {
		util.Warn("ptracegetregs: " + err.Error())
		return err
	}
	session.CheckRestrictedCall(pid, pregs)
	return nil
}