//go:build !windows

package tui

import (
	"os"

	"golang.org/x/sys/unix"
)

// ttyOutput returns the terminal output for the UI, a private copy of stdout
// that stays attached to the terminal while redirectStdio captures fd 1.
func ttyOutput() (*os.File, func()) {
	dup, err := unix.Dup(1)
	if err != nil {
		return os.Stdout, func() {}
	}
	out := os.NewFile(uintptr(dup), "tui-output")
	return out, func() { _ = out.Close() }
}

// redirectStdio points file descriptors 1 and 2 at w, catching writes from
// code holding its own reference to the original stdout. The returned function
// restores the descriptors.
func redirectStdio(w *os.File) func() {
	saved := map[int]int{}
	for _, fd := range []int{1, 2} {
		dup, err := unix.Dup(fd)
		if err != nil {
			continue
		}
		if err := unix.Dup2(int(w.Fd()), fd); err != nil {
			_ = unix.Close(dup)
			continue
		}
		saved[fd] = dup
	}
	return func() {
		for fd, dup := range saved {
			_ = unix.Dup2(dup, fd)
			_ = unix.Close(dup)
		}
	}
}
