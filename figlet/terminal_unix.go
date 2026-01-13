//go:build !windows && !js

package figlet

import (
	"os"
	"syscall"
	"unsafe"
)

// GetColumns returns the terminal width
func GetColumns() int {
	fd, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return -1
	}
	defer fd.Close()

	var ws struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return -1
	}
	return int(ws.Col)
}
