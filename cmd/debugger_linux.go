package cmd

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// IsLaunchedByDebugger discovers if process is traced by a debugger (e.g. delve)
func IsLaunchedByDebugger() bool {
	f, err := os.Open("/proc/self/status")
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "TracerPid:") {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "TracerPid:")))
		if err != nil {
			return false
		}
		return pid != 0
	}
	return false
}
