package syncedcounter

import "sync"

type (
	Counter struct {
		c int
		m sync.Mutex
	}
)

// Inc counter
func (counter *Counter) Inc() int {
	counter.m.Lock()
	defer counter.m.Unlock()

	counter.c++
	return counter.c
}

// Dec counter
func (counter *Counter) Dec() int {
	counter.m.Lock()
	defer counter.m.Unlock()

	counter.c--
	return counter.c
}

// Reset counter
func (counter *Counter) Reset() {
	counter.m.Lock()
	defer counter.m.Unlock()

	counter.c = 0
}
