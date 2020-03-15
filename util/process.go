package util

import (
	"bytes"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

func GetPtrsFromCmd(cmd *exec.Cmd) (*os.File, error) {
	f, err := os.Open(cmd.Path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func IsPidRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	} else {
		err := process.Signal(unix.Signal(0))
		return err == nil || (err.Error() != "no such process" && err.Error() != "os: process already finished")
	}
}

// grabs the string at the given address without process_vm_readv.
func ReadPeekString(pid int, address uintptr) (string, error) {
	word := make([]byte, unix.PathMax)
	_, err := unix.PtracePeekData(pid, address, word)
	if err != nil {
		return "", err
	}
	length := bytes.IndexByte(word, 0)
	if length == -1 {
		length = syscall.PathMax
	}
	//v := uint64(0x2Bc0ffee)
	//err = binary.Read(bytes.NewReader(word), binary.LittleEndian, &v)
	return string(word[:length]), nil
}

// process_vm_readv
// modified from here: https://github.com/criyle/go-sandbox/

func ProcessVmReadVStr(pid int, address uintptr) (string, error) {
	buff := make([]byte, unix.PathMax)

	n := 0
	r := os.Getpagesize() - int(address % uintptr(os.Getpagesize()))
	if r == 0 {
		r = os.Getpagesize()
	}

	for len(buff) > 0 {
		if l := len(buff); r < l {
			r = l
		}

		nn, err := processVmReadV(pid, address+uintptr(n), buff[:r])
		if err != nil {
			return "", err
		}

		if hasNull(buff[:nn]) {
			break
		}

		n += nn
		buff = buff[nn:]
		r = os.Getpagesize()
	}
	return string(buff[:clen(buff)]), nil
}

func processVmReadV(pid int, addr uintptr, buff []byte) (int, error) {
	l := len(buff)
	localIov := getIovecs(&buff[0], l)
	remoteIov := getIovecs((*byte)(unsafe.Pointer(addr)), l)
	n, _, err := unix.Syscall6(unix.SYS_PROCESS_VM_READV, uintptr(pid), uintptr(unsafe.Pointer(&localIov[0])), uintptr(len(localIov)), uintptr(unsafe.Pointer(&remoteIov[0])), uintptr(len(remoteIov)), 0)
	if err == 0 {
		return int(n), nil
	}
	return int(n), err
}

func clen(b []byte) int {
	for i := 0; i < len(b); i++ {
		if b[i] == 0 {
			return i
		}
	}
	return len(b) + 1
}

func hasNull(buff []byte) bool {
	for _, b := range buff {
		if b == 0 {
			return true
		}
	}
	return false
}

func getIovecs(base *byte, l int) []unix.Iovec {
	return []unix.Iovec{{
		Base: base,
		Len:  uint64(l),
	}}
}