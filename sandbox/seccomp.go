package sandbox

import "github.com/elastic/go-seccomp-bpf"

func (session *RunnerSession) LoadSeccompFilter() error {
	// create seccomp filter
	filter := seccomp.Filter{
		NoNewPrivs: true,
		Flag:       seccomp.FilterFlagTSync,
		Policy: seccomp.Policy{
			DefaultAction: seccomp.ActionErrno, // trap syscalls by default
			Syscalls: []seccomp.SyscallGroup{
				{ // allowed syscalls
					Action: seccomp.ActionAllow,
					Names: session.SeccompProfile.SyscallAllow,
				},
				{ // restricted syscalls
					Action: seccomp.ActionTrace,
					Names: session.SeccompProfile.SyscallTrace,
				},
			},
		},
	}

	return seccomp.LoadFilter(filter)
}