package sandbox

import (
	"github.com/elastic/go-seccomp-bpf"
	"github.com/pkg/errors"
	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"
)

// from go-seccomp-bpf/seccomp_linux
func (session *RunnerSession) CreateSeccompFilter() error {
	// create seccomp filter
	filter := seccomp.Filter{
		NoNewPrivs: true,
		Flag:       seccomp.FilterFlagTSync,
		Policy:     session.SeccompProfile.SeccompPolicy,
	}

	insts, err := filter.Policy.Assemble()
	if err != nil {
		return errors.Wrap(err, "failed to assemble policy")
	}

	raw, err := bpf.Assemble(insts)
	if err != nil {
		return errors.Wrap(err, "failed to assemble BPF instructions")
	}

	sockFilter := sockFilter(raw)
	session.Seccomp = &unix.SockFprog{
		Len:    uint16(len(sockFilter)),
		Filter: &sockFilter[0],
	}
	return nil
}

func sockFilter(raw []bpf.RawInstruction) []unix.SockFilter {
	filter := make([]unix.SockFilter, 0, len(raw))
	for _, instruction := range raw {
		filter = append(filter, unix.SockFilter{
			Code: instruction.Op,
			Jt:   instruction.Jt,
			Jf:   instruction.Jf,
			K:    instruction.K,
		})
	}
	return filter
}