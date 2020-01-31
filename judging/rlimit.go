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
	// only second precision :/
	err := doRlimit(unix.RLIMIT_CPU, uint64(proc.Session.CurrentBatch.TimeLimit))
	if err != nil {
		return err
	}

	// maximum output size
	// 1e9 bytes -> 1 gigabyte
	err = doRlimit(unix.RLIMIT_FSIZE, uint64(1e9))
	if err != nil {
		return err
	}

	// maximum memory
	// MB -> bytes
	err = doRlimit(unix.RLIMIT_AS, uint64(proc.Session.CurrentBatch.MemLimit*1e6))
	if err != nil {
		return err
	}

	// other limits maybe?
	// RLIMIT_STACK - set max stack size
	return nil
}
