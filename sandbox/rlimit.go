package sandbox

import "golang.org/x/sys/unix"

func (session *RunnerSession) SetRlimits() error {
	// https://linux.die.net/man/2/setrlimit

	for _, rlimit := range session.RLimits {
		if rlimit.Cur > 0 {
			err := unix.Setrlimit(rlimit.Type, &unix.Rlimit{
				Cur: rlimit.Cur,
				Max: rlimit.Max,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}