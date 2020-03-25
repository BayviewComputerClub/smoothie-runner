package sandbox

import "golang.org/x/sys/unix"

// https://github.com/DMOJ/judge-server/blob/c4b5e52665dfb6100dfa2de7429a05fd1a72ae7f/dmoj/cptbox/helper.cpp

// init rlimit
// https://linux.die.net/man/2/setrlimit
func (session *RunnerSession) InitRLimits() {
	session.RLimits = []RLimit{
		{
			Type: unix.RLIMIT_STACK,
			Cur:  uint64(unix.RLIM_INFINITY),
			Max:  uint64(unix.RLIM_INFINITY),
		},
		{ // no core dump
			Type: unix.RLIMIT_CORE,
			Cur:  0,
			Max:  0,
		},
	}

	if session.TimeLimit.Seconds() > 0 {
		session.RLimits = append(session.RLimits, RLimit{
			Type: unix.RLIMIT_CPU,
			Cur:  uint64(session.TimeLimit.Seconds()),
			Max:  uint64(session.TimeLimit.Seconds() + 1),
		})
	}

	if session.MemoryLimit > 0 {
		session.RLimits = append(session.RLimits, RLimit{
			Type: unix.RLIMIT_DATA,
			Cur:  session.MemoryLimit,
			Max:  session.MemoryLimit,
		}, RLimit{
			Type: unix.RLIMIT_AS,
			Cur:  session.MemoryLimit + 4096*1024,
			Max:  session.MemoryLimit + 4096*1024,
		})
	}

	if session.FSizeLimit >= 0 {
		session.RLimits = append(session.RLimits, RLimit{
			Type: unix.RLIMIT_FSIZE,
			Cur:  uint64(session.FSizeLimit),
			Max:  uint64(session.FSizeLimit),
		})
	}

	if session.NProcLimit >= 0 {
		session.RLimits = append(session.RLimits, RLimit{
			Type: unix.RLIMIT_NPROC,
			Cur:  uint64(session.NProcLimit),
			Max:  uint64(session.NProcLimit),
		})
	}
}

func (session *RunnerSession) SetRlimits() error {
	for _, rlimit := range session.RLimits {
		err := unix.Setrlimit(rlimit.Type, &unix.Rlimit{
			Cur: rlimit.Cur,
			Max: rlimit.Max,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
