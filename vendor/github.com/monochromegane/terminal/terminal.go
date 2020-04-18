// +build linux darwin,!cgo freebsd,!cgo

package terminal

import (
	"os"
	"syscall"
	"unsafe"
)

func IsTerminal(file *os.File) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, file.Fd(),
		ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}
