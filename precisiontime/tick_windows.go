package precisiontime

import (
	"github.com/pkg/errors"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.MustLoadDLL("kernel32.dll")
	qpc      = kernel32.MustFindProc("QueryPerformanceCounter")
)

//Tick Current clock tick via query performance counter
func Tick() (int64, error) {
	var tick int64
	ret, _, err := qpc.Call(uintptr(unsafe.Pointer(&tick)))
	if ret == 0 {
		return -1, errors.Wrap(error(err), "QueryPerformanceCounter request failed")
	}
	return tick, nil
}
