package util

type SandboxProfile struct {
	AllowRead  map[string]bool
	AllowWrite map[string]bool

	SyscallAllow map[string]bool
	SyscallTrace map[string]bool // restricted calls
}

// "inspired" by https://github.com/criyle/go-sandbox/tree/master/config
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
			"/usr/share/zoneinfo/":           true,
			"/dev/random":                    true,
			"/dev/urandom":                   true,
			"/proc/meminfo":                  true,
			"/etc/localtime":                 true,
		},
		AllowWrite: map[string]bool{
			"/dev/null": true,
		},
		SyscallAllow: map[string]bool{
			// file access through fd
			"read": true,
			"write": true,
			"readv": true,
			"writev": true,
			"close": true,
			"fstat": true,
			"lseek": true,
			"dup": true,
			"dup2": true,
			"dup3": true,
			"ioctl": true,
			"fcntl": true,
			"fadvise64": true,

			// memory action
			"mmap": true,
			"mprotect": true,
			"munmap": true,
			"brk": true,
			"mremap": true,
			"msync": true,
			"mincore": true,
			"madvise": true,

			// signal action
			"rt_sigaction": true,
			"rt_sigprocmask": true,
			"rt_sigreturn": true,
			"rt_sigpending": true,
			"sigaltstack": true,

			// get current work dir
			"getcwd": true,

			// process exit
			"exit": true,
			"exit_group": true,

			// others
			"arch_prctl": true,
			"gettimeofday": true,
			"getrlimit": true,
			"getrusage": true,
			"times": true,
			"time": true,
			"clock_gettime": true,
			"restart_syscall": true,
		},
		SyscallTrace: map[string]bool{
			// execute file
			"execve": true,
			"execveat": true,

			// file open
			"open": true,
			"openat": true,

			// file delete
			"unlink": true,
			"unlinkat": true,

			// soft link
			"readlink": true,
			"readlinkat": true,

			// permission check
			"lstat": true,
			"stat": true,
			"access": true,
			"faccessat": true,
		},
	}

	// compiler profile
)
