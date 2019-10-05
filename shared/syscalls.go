package shared

import "golang.org/x/sys/unix"

var ALLOWED_CALLS = []uint64{
	unix.SYS_READ,
	unix.SYS_WRITE,
	unix.SYS_EXIT,
	unix.SYS_STAT,
	unix.SYS_FSTAT,
	unix.SYS_CLOSE,
	unix.SYS_RT_SIGRETURN,
	unix.SYS_OPENAT,
	unix.SYS_WRITEV,
	unix.SYS_EXIT_GROUP,
	unix.SYS_BRK,
	unix.SYS_NANOSLEEP,
	unix.SYS_CLOCK_NANOSLEEP,
	unix.SYS_MMAP,
	unix.SYS_MPROTECT,
	unix.SYS_ACCESS,
	unix.SYS_ARCH_PRCTL,
	unix.SYS_MUNMAP,
}
