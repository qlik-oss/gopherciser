package atomichandlers

import (
	"fmt"
	"sync/atomic"
)

type (
	// AtomicCounter atomic integer counter
	AtomicCounter struct {
		counter uint64
	}
)

// Inc increase counter by 1
func (value *AtomicCounter) Inc() uint64 {
	return atomic.AddUint64(&value.counter, 1)
}

// Current value of counter
func (value *AtomicCounter) Current() uint64 {
	return atomic.LoadUint64(&value.counter)
}

// Reset counter to 0
func (value *AtomicCounter) Reset() {
	atomic.StoreUint64(&value.counter, 0)
}

// Dec decrease the counter by 1
func (value *AtomicCounter) Dec() uint64 {
	return atomic.AddUint64(&value.counter, ^uint64(0))
}

// Add value to counter
func (value *AtomicCounter) Add(u uint64) uint64 {
	return atomic.AddUint64(&value.counter, u)
}

// SwapAndCompare new value, returns false without updating to new value if old value no longer the same
func (value *AtomicCounter) CompareAndSwap(old, new uint64) bool {
	return atomic.CompareAndSwapUint64(&value.counter, old, new)
}

// String representation of counter
func (value *AtomicCounter) String() string {
	return fmt.Sprintf("%d", value.Current())
}
