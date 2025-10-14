package helpers

import "sync"

type SyncCounter struct {
	c  int
	mu sync.Mutex
}

// Inc increase and return new value
func (counter *SyncCounter) Inc() int {
	counter.mu.Lock()
	defer counter.mu.Unlock()
	counter.c++
	return counter.c
}

// Dec decrease and return new value
func (counter *SyncCounter) Dec() int {
	counter.mu.Lock()
	defer counter.mu.Unlock()
	counter.c--
	return counter.c
}

// Current value
func (counter *SyncCounter) Current() int {
	counter.mu.Lock()
	defer counter.mu.Unlock()
	return counter.c
}
