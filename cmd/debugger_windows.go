package cmd

import (
	"syscall"
)

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	procIsDebuggerPresent = kernel32.NewProc("IsDebuggerPresent")
)

// IsLaunchedByDebugger discovers if process is traced by a debugger (e.g. delve)
func IsLaunchedByDebugger() bool {
	ret, _, _ := procIsDebuggerPresent.Call()
	return ret != 0
}
