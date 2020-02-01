package util

type SandboxProfile struct {
	AllowRead  map[string]bool
	AllowWrite map[string]bool

	SyscallAllow []string
	SyscallTrace []string // restricted calls
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
		SyscallAllow: []string{
			// file access through fd
			"read",
			"write",
			"readv",
			"writev",
			"close",
			"fstat",
			"lseek",
			"dup",
			"dup2",
			"dup3",
			"ioctl",
			"fcntl",
			"fadvise64",

			// memory action
			"mmap",
			"mprotect",
			"munmap",
			"brk",
			"mremap",
			"msync",
			"mincore",
			"madvise",

			// signal action
			"rt_sigaction",
			"rt_sigprocmask",
			"rt_sigreturn",
			"rt_sigpending",
			"sigaltstack",

			// get current work dir
			"getcwd",

			// process exit
			"exit",
			"exit_group",

			// others
			"arch_prctl",
			"gettimeofday",
			"getrlimit",
			"getrusage",
			"times",
			"time",
			"clock_gettime",
			"restart_syscall",
		},
		SyscallTrace: []string{
			// execute file
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
	}

	// compiler profile
)
