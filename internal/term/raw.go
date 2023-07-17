//go:build !windows
// +build !windows

package term

import (
	"syscall"

	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
)

// SetRaw put terminal into a raw mode.
func SetRaw(fd int) error {
	originalTermios, err := getOriginalTermios(fd)
	if err != nil {
		return err
	}

	// Copy the state.
	term := *originalTermios

	term.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK |
		syscall.ISTRIP | syscall.INLCR | syscall.IGNCR |
		syscall.ICRNL | syscall.IXON
	term.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG | syscall.ECHONL
	term.Cflag &^= syscall.CSIZE | syscall.PARENB
	term.Cflag |= syscall.CS8 // Set to 8-bit wide.  Typical value for displaying characters.
	term.Cc[syscall.VMIN] = 1
	term.Cc[syscall.VTIME] = 0

	return termios.Tcsetattr(uintptr(fd), termios.TCSANOW, (*unix.Termios)(&term))
}
