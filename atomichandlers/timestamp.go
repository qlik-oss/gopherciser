package atomichandlers

import (
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

type (
	// AtomicTimeStamp goroutine safe timestamp
	AtomicTimeStamp struct {
		p unsafe.Pointer
	}
)

const (
	// DefaultSwapTries max tries for swaping timestamp
	DefaultSwapTries = 100
)

//Current timestamp
func (ats *AtomicTimeStamp) Current() time.Time {
	ts := (*time.Time)(atomic.LoadPointer(&ats.p))
	if ts == nil {
		return time.Time{}
	}
	return *ts
}

//Reset timestamp to zero
func (ats *AtomicTimeStamp) Reset() {
	ats.Set(time.Time{})
}

//Set new timestamp
func (ats *AtomicTimeStamp) Set(ts time.Time) {
	// nolint: gas
	atomic.StorePointer(&ats.p, unsafe.Pointer(&ts))
}

//SetIfOlder set new timestamp if older than current
func (ats *AtomicTimeStamp) SetIfOlder(ts time.Time) error {
	condition := func(currentP unsafe.Pointer, currentV *time.Time, newV time.Time) bool {
		return currentP == nil || currentV == nil || currentV.IsZero() || newV.Before(*currentV)
	}
	if !ats.compareAndSwap(ts, condition) {
		return errors.Errorf("failed to set older atomic timestamp within %d tries", DefaultSwapTries)
	}

	return nil
}

//SetIfNewer set new timestamp if new than current
func (ats *AtomicTimeStamp) SetIfNewer(ts time.Time) error {
	condition := func(currentP unsafe.Pointer, currentV *time.Time, newV time.Time) bool {
		return currentP == nil || currentV == nil || currentV.IsZero() || newV.After(*currentV)
	}
	if !ats.compareAndSwap(ts, condition) {
		return errors.Errorf("failed to set newer atomic timestamp within %d tries", DefaultSwapTries)
	}

	return nil
}

func (ats *AtomicTimeStamp) compareAndSwap(ts time.Time, condition func(currentP unsafe.Pointer, currentV *time.Time, newV time.Time) bool) bool {
	// nolint: gas
	newp := unsafe.Pointer(&ts)
	for i := 0; i < DefaultSwapTries; i++ {
		currentP := atomic.LoadPointer(&ats.p)
		currentV := (*time.Time)(currentP)
		if condition(currentP, currentV, ts) {
			if atomic.CompareAndSwapPointer(&ats.p, currentP, newp) {
				return true
			}
		} else {
			return true
		}
	}
	return false
}
