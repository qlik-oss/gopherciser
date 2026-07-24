//go:build !linux && !windows && !darwin

package cmd

// IsLaunchedByDebugger discovers if process is traced by a debugger (e.g. delve)
func IsLaunchedByDebugger() bool {
	return false
}
