//go:build windows

package figlet

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

type coord struct {
	X int16
	Y int16
}

type smallRect struct {
	Left   int16
	Top    int16
	Right  int16
	Bottom int16
}

type consoleScreenBufferInfo struct {
	Size              coord
	CursorPosition    coord
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize coord
}

// GetColumns returns the terminal width
func GetColumns() int {
	handle, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return -1
	}

	var info consoleScreenBufferInfo
	r1, _, _ := procGetConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&info)))
	if r1 == 0 {
		return -1
	}

	return int(info.Size.X)
}

// Suppress unused import warnings
var _ = os.Stdout
