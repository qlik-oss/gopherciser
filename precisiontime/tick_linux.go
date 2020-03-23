package precisiontime

import (
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

// Current linux clocks:
// #define CLOCK_REALTIME			0
// #define CLOCK_MONOTONIC			1
// #define CLOCK_PROCESS_CPUTIME_ID	2
// #define CLOCK_THREAD_CPUTIME_ID	3
// #define CLOCK_MONOTONIC_RAW		4
// #define CLOCK_REALTIME_COARSE	5
// #define CLOCK_MONOTONIC_COARSE	6
const (
	clockMonotonicRaw = 4
)

//Tick Current clock tick via
func Tick() (int64, error) {
	var ts syscall.Timespec
	if _, _, err := syscall.Syscall(syscall.SYS_CLOCK_GETTIME, clockMonotonicRaw, uintptr(unsafe.Pointer(&ts)), 0); err != 0 {
		return -1, errors.Wrap(error(err), "get_time request failed")
	}
	return ts.Nano(), nil
}
