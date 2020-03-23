package atomichandlers

import (
	"sync"
	"testing"
)

func TestCounters(t *testing.T) {
	t.Parallel()

	var counter AtomicCounter

	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			counter.Inc()
			wg.Done()
		}()
	}
	wg.Wait()
	c := counter.Current()
	if c != 3 {
		t.Errorf("Unexpected counter value<%d>, expected<3>", c)
	}

	c = counter.Dec()
	if c != 2 {
		t.Errorf("Unexpected counter value<%d>, expected<2>", c)
	}

	c = counter.Add(3)
	if c != 5 {
		t.Errorf("Unexpected counter value<%d>, expected<5>", c)
	}

	counter.Reset()
	c = counter.Current()
	if c != 0 {
		t.Errorf("Unexpected counter value<%d>, expected<0>", c)
	}
}

func TestParallelCounters(t *testing.T) {
	var counter AtomicCounter

	var wg sync.WaitGroup

	add := func(c int) {
		for i := 0; i < c; i++ {
			go func() {
				defer wg.Done()
				counter.Inc()
			}()
		}
	}

	sub := func(c int) {
		for i := 0; i < c; i++ {
			go func() {
				defer wg.Done()
				counter.Dec()
			}()
		}
	}

	addCount := 50
	subCount := 40

	//test concurrent increasing and decreasing
	wg.Add(addCount)
	add(addCount)

	wg.Add(subCount)
	sub(subCount)

	wg.Wait()

	current := counter.Current()
	expected := uint64(addCount - subCount)
	if current != expected {
		t.Errorf("Expected counter<%d> got<%d>", expected, current)
	}

	counter.Reset()

	//test reset during counting
	wg.Add(addCount)
	add(addCount)
	counter.Reset()
	wg.Wait()
	t.Log(counter.Current())
}

func BenchmarkCounter(b *testing.B) {
	var c AtomicCounter
	for i := 0; i < b.N; i++ {
		c.Inc()
	}
}

type muC struct {
	c  int64
	mu sync.Mutex
}

func (c *muC) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.c++
}

func BenchmarkMuCounter(b *testing.B) {
	var c muC
	for i := 0; i < b.N; i++ {
		c.Inc()
	}
}

func BenchmarkNonSafeCounter(b *testing.B) {
	var c uint64
	for i := 0; i < b.N; i++ {
		c++
	}
}
