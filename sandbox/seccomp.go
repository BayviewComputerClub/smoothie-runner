package sandbox

import "github.com/elastic/go-seccomp-bpf"

func (session *RunnerSession) LoadSeccompFilter() error {
	// create seccomp filter
	filter := seccomp.Filter{
		NoNewPrivs: true,
		Flag:       seccomp.FilterFlagTSync,
		Policy:     session.SeccompProfile.SeccompPolicy,
	}

	return seccomp.LoadFilter(filter)
}
