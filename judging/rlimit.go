package judging

import "golang.org/x/sys/unix"

func doRlimit(t int, v uint64) error {
	return unix.Setrlimit(t, &unix.Rlimit{
		Cur: v,
		Max: v,
	})
}

func (proc *ForkProcess) SetRlimits() error {
	// https://linux.die.net/man/2/setrlimit

	// max time
	if proc.Session.Limit.CpuTime > 0 {
		err := doRlimit(unix.RLIMIT_CPU, proc.Session.Limit.CpuTime)
		if err != nil {
			return err
		}
	}

	// maximum output size
	if proc.Session.Limit.Fsize > 0 {
		err := doRlimit(unix.RLIMIT_FSIZE, proc.Session.Limit.Fsize)
		if err != nil {
			return err
		}
	}

	// maximum memory
	if proc.Session.Limit.Memory > 0 {
		err := doRlimit(unix.RLIMIT_AS, proc.Session.Limit.Memory)
		if err != nil {
			return err
		}
	}

	// other limits maybe?
	// RLIMIT_STACK - set max stack size
	return nil
}
