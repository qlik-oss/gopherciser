//go:build windows

package tui

import "os"

// ttyOutput returns the terminal output for the UI, the console handle held
// before redirectStdio swaps the Go level stdout.
func ttyOutput() (*os.File, func()) { return os.Stdout, func() {} }

// redirectStdio swaps the Go level stdout and stderr for w, the returned
// function restores them.
func redirectStdio(w *os.File) func() {
	stdout, stderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = w, w
	return func() { os.Stdout, os.Stderr = stdout, stderr }
}
