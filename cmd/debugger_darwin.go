package cmd

import (
	"os"

	"golang.org/x/sys/unix"
)

// pTraced is the P_TRACED process flag from sys/proc.h
const pTraced = 0x00000800

// IsLaunchedByDebugger discovers if process is traced by a debugger (e.g. delve)
func IsLaunchedByDebugger() bool {
	kp, err := unix.SysctlKinfoProc("kern.proc.pid", os.Getpid())
	if err != nil {
		return false
	}
	return kp.Proc.P_flag&pTraced != 0
}
