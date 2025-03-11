package helpers

import (
	"sync"
	"time"
)

type (
	TimerPool struct {
		pool sync.Pool
	}
)

var GlobalTimerPool = NewTimerPool()

func NewTimerPool() *TimerPool {
	return &TimerPool{sync.Pool{New: newTimer}}
}

func newTimer() any {
	return time.NewTimer(0)
}

func (tp *TimerPool) Get(delay time.Duration) *time.Timer {
	tmr := tp.pool.Get().(*time.Timer)
	tmr.Reset(delay)
	return tmr
}

func (tp *TimerPool) Put(timer *time.Timer) {
	timer.Stop()
	tp.pool.Put(timer)
}
