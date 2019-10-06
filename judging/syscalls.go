package judging

import (
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"golang.org/x/sys/unix"
	"log"
	"math"
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

const (
	LONG_SIZE = 8
)

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

func readPeekString(pid int, addr uintptr, length int) string {
	str := ""

	i := 0
	j := length/LONG_SIZE

	for i < j {
		c, err := unix.PtracePeekData(pid, addr, nil)

	}
	return str
}

func blockRestrictedCalls(pregs *unix.PtraceRegs, pid int) bool {
	var blockedCall bool

	if blockedCall = isRestrictedSyscall(pregs.Orig_rax); blockedCall {


		// linux support only in this section (peek and poke not on BSDs)
		// i wish i knew how to call process_vm_readv

	} else if blockedCall = !isAllowedSyscall(pregs.Orig_rax); blockedCall {
		log.Printf("Blocked: %v\n", pregs.Orig_rax)// TODO

		pregs.Orig_rax = uint64(math.Inf(0))
		err := unix.PtraceSetRegs(pid, pregs)
		if err != nil {
			util.Warn("ptracesetregs: " + err.Error())
		}
	}

	return blockedCall
}


