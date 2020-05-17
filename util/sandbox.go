package util

import (
	"github.com/elastic/go-seccomp-bpf"
)

type SandboxProfile struct {
	AllowRead  map[string]bool // will include AllowWrite entries
	AllowWrite map[string]bool

	DisallowRead  map[string]bool // TODO will automatically include DisallowWrite entries
	DisallowWrite map[string]bool // TODO

	SeccompPolicy seccomp.Policy
}

// "inspired" by https://github.com/DMOJ/judge-server/blob/master/dmoj/cptbox/isolate.py
// and https://github.com/criyle/go-sandbox/tree/master/config
// ~~~ thx there's no way i'm going through all that ~~~
var (

	// default profile
	SANDBOX_DEFAULT_PROFILE = SandboxProfile{
		AllowRead: map[string]bool{
			"/etc/ld.so.nohwcap":             true,
			"/etc/ld.so.preload":             true,
			"/etc/ld.so.cache":               true,
			"/usr/lib/locale/locale-archive": true,
			"/proc/self/exe":                 true,
			"/etc/timezone":                  true,
			"/usr/share/zoneinfo":            true,
			"/dev/random":                    true,
			"/dev/urandom":                   true,
			"/proc/meminfo":                  true,
			"/etc/localtime":                 true,

			"/usr/lib":       true,
			"/usr/lib64":     true,
			"/lib":           true,
			"/usr/local/lib": true,

			"main.py": true,
		},
		AllowWrite: map[string]bool{
			"/dev/null": true,
		},
		DisallowRead: map[string]bool{

		},
		DisallowWrite: map[string]bool{

		},
		SeccompPolicy: seccomp.Policy{
			DefaultAction: seccomp.ActionErrno, // trap syscalls by default
			Syscalls: []seccomp.SyscallGroup{
				{
					Action: seccomp.ActionAllow,
					Names: []string{
						// file access through fd
						"read",
						"readv",
						"pread64",
						"write",
						"writev",
						"statfs",
						"getpgrp",
						"restart_syscall",
						"select",
						"modify_ldt",
						"ppoll",

						"sched_getaffinity",
						"sched_getparam",
						"sched_getscheduler",
						"sched_get_priority_min",
						"sched_get_priority_max",
						"timerfd_create",
						"timer_create",
						"timer_settime",
						"timer_delete",

						"rt_sigreturn",
						"nanosleep",
						"sysinfo",
						"getrandom",

						"close",
						"dup",
						"dup2",
						"dup3",
						"fstat",
						"mmap",
						"mremap",
						"mprotect",
						"madvise",
						"munmap",
						"brk",
						"fcntl",
						"arch_prctl",
						"set_tid_address",
						"set_robust_list",
						"futex",
						"rt_sigaction",
						"rt_sigprocmask",
						"getrlimit",
						"ioctl",
						"getcwd",
						"geteuid",
						"getuid",
						"getegid",
						"getgid",
						"getdents",
						"getdents64",
						"lseek",
						"getrusage",
						"sigaltstack",
						"pipe",
						"pipe2",
						"clock_gettime",
						"clock_getres",
						"gettimeofday",
						"getpid",
						"getppid",
						"sched_yield",

						"clone",
						"exit",
						"exit_group",
						"gettid",

						// extra
						"fadvise64",

						"msync",
						"mincore",

						"rt_sigpending",

						"times",
						"time",

						"set_thread_area",
						"uname",
						"setrlimit", // restricted once exec is called (this is to avoid memory issues before exec, especially when loading seccomp filter)
					},
				},
				{
					Action: seccomp.ActionTrace,
					Names: []string{
						// execute initial file (then blocked)
						"execve",
						"execveat",

						// file open
						"open",
						"openat",

						// file delete
						"unlink",
						"unlinkat",

						// soft link
						"readlink",
						"readlinkat",

						// permission check
						"lstat",
						"stat",

						"access",
						"faccessat",
					},
				},
			},
		},
	}

	// compiler profile
	SANDBOX_COMPILER_PROFILE = SandboxProfile{
		AllowRead: map[string]bool{
			".":          true,
			"../runtime": true,
			"/etc/oracle/java/usagetracker.properties": true,
			"/usr":                    true,
			"/lib":                    true,
			"/lib64":                  true,
			"/bin":                    true,
			"/sbin":                   true,
			"/sys/devices/system/cpu": true,
			"/proc":                   true,
			"/etc/timezone":           true,
			"/etc/fpc-2.6.2.cfg.d":    true,
			"/etc/fpc.cfg":            true,
		},
		AllowWrite: map[string]bool{
			"/tmp": true,
			".":    true,
		},
		DisallowRead: map[string]bool{
			"/dev/null": true,
			"/dev/tty":  true,
			"/dev/zero": true,
		},
		DisallowWrite: map[string]bool{
			"/etc/nsswitch.conf": true,
			"/etc/resolv.conf":   true,
			"/etc/passwd":        true,
			"/etc/malloc.conf":   true,
		},
		SeccompPolicy: seccomp.Policy{
			DefaultAction: seccomp.ActionAllow, // allow syscalls by default
			Syscalls: []seccomp.SyscallGroup{
				{
					Action: seccomp.ActionTrace,
					Names: []string{
						// file open
						"open",
						"openat",

						// file delete
						"unlink",
						"unlinkat",

						// soft link
						"readlink",
						"readlinkat",

						// permission check
						"lstat",
						"stat",

						"access",
						"faccessat",
					},
				},
			},
		},
	}
)

func init() {
	for k, v := range SANDBOX_DEFAULT_PROFILE.AllowRead {
		SANDBOX_COMPILER_PROFILE.AllowRead[k] = v
	}
	for k, v := range SANDBOX_DEFAULT_PROFILE.AllowWrite {
		SANDBOX_COMPILER_PROFILE.AllowRead[k] = v
	}
}
