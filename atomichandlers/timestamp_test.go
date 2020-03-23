package atomichandlers

import (
	"sync"
	"testing"
	"time"
)

func TestTimestamp(t *testing.T) {
	t.Parallel()

	ats := AtomicTimeStamp{}

	ts := ats.Current()
	if !ts.IsZero() {
		t.Errorf("Unexpected inial timestamp<%v>", ts)
	}

	now := time.Now()

	ats.Set(now)
	ts = ats.Current()
	if !ts.Equal(now) {
		t.Fatalf("Unexpected now timestamp<%v>", ts)
	}

	syncChan := make(chan interface{})
	defer close(syncChan)

	now = time.Now()
	go func() {
		ats.Set(now)
		syncChan <- nil
	}()

	select {
	case <-syncChan:
	case <-time.After(time.Second):
	}

	ts = ats.Current()
	if !ts.Equal(now) {
		t.Errorf("Unexpected async now timestamp<%v>", ts)
	}

	ats.Reset()
	ts = ats.Current()
	if !ts.IsZero() {
		t.Errorf("Unexpected reset timestamp<%v>", ts)
	}

	time1 := time.Now()
	if err := ats.SetIfNewer(time1); err != nil {
		t.Fatal(err)
	}
	ts = ats.Current()
	if ts != time1 {
		t.Fatalf("Unexpected newer timestamp<%v> expected<%v>", ts, time1)
	}

	older := now.Add(-time.Hour)
	if err := ats.SetIfOlder(older); err != nil {
		t.Fatal(err)
	}
	ts = ats.Current()
	if ts != older {
		t.Errorf("Unexpected older timestamp<%v> expected<%v>", ts, now)
	}

	//Test nil pointer
	ats = AtomicTimeStamp{}
	if err := ats.SetIfNewer(now); err != nil {
		t.Error(err)
	}
	ts = ats.Current()
	if !ts.Equal(now) {
		t.Errorf("Unexpected current timestamp<%v>", ts)
	}

	ats = AtomicTimeStamp{}
	if err := ats.SetIfOlder(now); err != nil {
		t.Error(err)
	}
	ts = ats.Current()
	if !ts.Equal(now) {
		t.Errorf("Unexpected current timestamp<%v>", ts)
	}
}

func TestParallelTimestampWrites(t *testing.T) {
	t.Parallel()

	ats := AtomicTimeStamp{}
	var wg sync.WaitGroup

	// Test set
	for i := 0; i < 1000; i++ {
		wg.Add(1)

		switch i % 5 {
		case 0: // Test Set
			go func() {
				defer wg.Done()
				ats.Set(time.Now())
			}()
		case 1: // Test swap older
			go func() {
				defer wg.Done()
				ts := time.Now()
				if err := ats.SetIfOlder(ts); err != nil {
					t.Errorf("Failed to set if older: %v", ts)
				}
			}()
		case 2: // Test swap newer
			go func() {
				defer wg.Done()
				ts := time.Now()
				if err := ats.SetIfNewer(ts); err != nil {
					t.Errorf("Failed to set if newer: %v", ts)
				}
			}()
		case 3: // Test reset
			go func() {
				defer wg.Done()
				ats.Reset()
			}()
		case 4: // Test current
			go func() {
				defer wg.Done()
				ats.Current()
			}()
		}
	}

	wg.Wait()
}
