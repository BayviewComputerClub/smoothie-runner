package judging

import (
	"bytes"
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"math"
	"strings"
	"syscall"
)

// "inspired" by DMOJ allowed syscalls hehe

var RESTRICTED_CALLS = []uint64 {
	unix.SYS_OPENAT,
	unix.SYS_FACCESSAT,
	unix.SYS_OPEN,
	unix.SYS_ACCESS,
	unix.SYS_MKDIR,
	unix.SYS_UNLINK,
	unix.SYS_READLINK,
	unix.SYS_READLINKAT,
	unix.SYS_STAT,
	unix.SYS_LSTAT,
	unix.SYS_FSTATFS,
	unix.SYS_TGKILL,
	unix.SYS_KILL,
	unix.SYS_PRCTL,
}

var ALLOWED_CALLS = []uint64{
	unix.SYS_READ,
	unix.SYS_WRITE,
	unix.SYS_WRITEV,
	unix.SYS_STATFS,
	unix.SYS_GETPGRP,
	unix.SYS_RESTART_SYSCALL,
	unix.SYS_SELECT,
	unix.SYS_MODIFY_LDT,
	unix.SYS_PPOLL,

	unix.SYS_GETGROUPS,
	unix.SYS_SCHED_GETAFFINITY,
	unix.SYS_SCHED_GETPARAM,
	unix.SYS_SCHED_GETSCHEDULER,
	unix.SYS_SCHED_GET_PRIORITY_MIN,
	unix.SYS_SCHED_GET_PRIORITY_MAX,
	unix.SYS_TIMER_CREATE,
	unix.SYS_TIMER_SETTIME,
	unix.SYS_TIMER_DELETE,

	unix.SYS_RT_SIGPROCMASK,
	unix.SYS_RT_SIGRETURN,
	unix.SYS_NANOSLEEP,
	unix.SYS_SYSINFO,
	unix.SYS_GETRANDOM,

	unix.SYS_CLOSE,
	unix.SYS_DUP,
	unix.SYS_DUP2,
	unix.SYS_DUP3,
	unix.SYS_FSTAT,
	unix.SYS_MMAP,
	unix.SYS_MREMAP,
	unix.SYS_MPROTECT,
	unix.SYS_MADVISE,
	unix.SYS_MUNMAP,
	unix.SYS_BRK,
	unix.SYS_FCNTL,
	unix.SYS_ARCH_PRCTL,
	unix.SYS_SET_TID_ADDRESS,
	unix.SYS_SET_ROBUST_LIST,
	unix.SYS_FUTEX,
	unix.SYS_RT_SIGACTION,
	unix.SYS_GETRLIMIT,
	unix.SYS_IOCTL,
	unix.SYS_GETCWD,
	unix.SYS_GETEUID,
	unix.SYS_GETUID,
	unix.SYS_GETEGID,
	unix.SYS_GETGID,
	unix.SYS_GETDENTS,
	unix.SYS_LSEEK,
	unix.SYS_GETRUSAGE,
	unix.SYS_SIGALTSTACK,
	unix.SYS_PIPE,
	unix.SYS_CLOCK_GETTIME,
	unix.SYS_CLOCK_GETRES,
	unix.SYS_GETTIMEOFDAY,
	unix.SYS_GETPID,
	unix.SYS_GETPPID,
	unix.SYS_SCHED_YIELD,

	// no clone
	unix.SYS_EXIT,
	unix.SYS_EXIT_GROUP,
	unix.SYS_GETTID,

	unix.SYS_CLOCK_NANOSLEEP,
}

var STOP_SIGNALS = []unix.Signal {
	unix.SIGQUIT,
	unix.SIGILL,
	unix.SIGABRT,
	unix.SIGFPE,
	unix.SIGBUS,
	unix.SIGSEGV,
	unix.SIGSYS,
	unix.SIGXCPU,
	unix.SIGXFSZ,
}

const (
	LONG_SIZE = 8
	BUFFER_SIZE = 4096

	// 64 bit registers
	RBX = 5
	RCX = 11
	RDX = 12
)

func isStopSignal(signal unix.Signal) bool {
	for _, a := range STOP_SIGNALS {
		if a == signal {
			return true
		}
	}
	return false
}

func isAllowedSyscall(call uint64) bool {
	for _, a := range ALLOWED_CALLS {
		if a == call {
			return true
		}
	}
	return false
}

func isRestrictedSyscall(call uint64) bool {
	for _, a := range RESTRICTED_CALLS {
		if a == call {
			return true
		}
	}
	return false
}

// grabs the string at the given address.
func readPeekString(pid int, address uintptr) (string, error) {
	word := make([]byte, unix.PathMax)
	_, err := unix.PtracePeekData(pid, address, word)
	if err != nil {
		return "", err
	}
	length := bytes.IndexByte(word, 0)
	if length == -1 {
		length = syscall.PathMax
	}
	//v := uint64(0x2Bc0ffee)
	//err = binary.Read(bytes.NewReader(word), binary.LittleEndian, &v)
	return string(word[:length]), nil
}

func sandboxChangeCall(pregs *unix.PtraceRegs, pid int, call uint64) {
	pregs.Orig_rax = call
	err := unix.PtraceSetRegs(pid, pregs)
	if err != nil {
		util.Warn("ptracesetregs: " + err.Error())
	}
}

func blockCall(pregs *unix.PtraceRegs, pid int) {
	shared.Debug(fmt.Sprintf("Blocked: %v", pregs.Orig_rax))
	sandboxChangeCall(pregs, pid, uint64(math.Inf(0)))
}

func isAllowedFile(path string) bool {
	return strings.HasPrefix(path, "/usr") || strings.HasPrefix(path, "/lib") || strings.HasPrefix(path, "/lib64") || strings.HasPrefix(path, "/bin")
}

// restrict call if necessary
// returns whether or not the call should be blocked
func correctRestrictedCall(pregs *unix.PtraceRegs, pid int) bool {
	switch int(pregs.Orig_rax) {
	case unix.SYS_OPENAT, unix.SYS_FACCESSAT:

		wd, err := readPeekString(pid, uintptr(pregs.Rsi)/*0x00400000*/)
		if err != nil {
			util.Warn("readpeekstring: " + err.Error())
			return true
		}
		shared.Debug(fmt.Sprintf("PEEKREAD: %v", wd))
		if !isAllowedFile(wd) {
			return true
		}

	case unix.SYS_OPEN, unix.SYS_ACCESS, unix.SYS_MKDIR, unix.SYS_UNLINK, unix.SYS_READLINK, unix.SYS_READLINKAT, unix.SYS_STAT, unix.SYS_LSTAT, unix.SYS_FSTATFS:

		wd, err := readPeekString(pid, uintptr(pregs.Rdi)/*0x00400000*/)
		if err != nil {
			util.Warn("readpeekstring: " + err.Error())
			return true
		}
		shared.Debug(fmt.Sprintf("PEEKREAD: %v", wd))
		if !isAllowedFile(wd) {
			return true
		}

	case unix.SYS_TGKILL:

	case unix.SYS_KILL:

	case unix.SYS_PRCTL:


		}
	return false
}

func blockRestrictedCalls(pregs *unix.PtraceRegs, pid int) bool {
	var blockedCall bool

	if isRestrictedSyscall(pregs.Orig_rax) {
		shared.Debug(fmt.Sprintf("Restricted: %v", pregs.Orig_rax))
		// linux support only in this section (peek and poke not on BSDs)
		// i wish i knew how to call process_vm_readv
		if correctRestrictedCall(pregs, pid) {
			blockCall(pregs, pid)
			blockedCall = true
		}

	} else if blockedCall = !isAllowedSyscall(pregs.Orig_rax); blockedCall {
		blockCall(pregs, pid)
	}

	return blockedCall
}


