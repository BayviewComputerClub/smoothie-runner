package judging

import (
	"github.com/elastic/go-seccomp-bpf"
)

func (proc *ForkProcess) LoadSeccompFilter() error {
	// create seccomp filter
	filter := seccomp.Filter{
		NoNewPrivs: true,
		Flag:       seccomp.FilterFlagTSync,
		Policy: seccomp.Policy{
			DefaultAction: seccomp.ActionErrno, // trap syscalls by default
			Syscalls: []seccomp.SyscallGroup{
				{ // allowed syscalls
					Action: seccomp.ActionAllow,
					Names: proc.Session.SandboxProfile.SyscallAllow,
				},
				{ // restricted syscalls
					Action: seccomp.ActionTrace,
					Names: proc.Session.SandboxProfile.SyscallTrace,
				},
			},
		},
	}

	return seccomp.LoadFilter(filter)
}
