package helpers_test

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/qlik-oss/gopherciser/helpers"
)

func TestTimerPool(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		refTime := 50 * time.Millisecond

		t1 := helpers.GlobalTimerPool.Get(refTime)
		helpers.GlobalTimerPool.Put(t1)

		t1 = helpers.GlobalTimerPool.Get(2 * refTime)
		ts := time.Now()
		<-t1.C
		dur := time.Since(ts)
		if dur != 2*refTime {
			t.Fatalf("timer fired after<%v> expected<%v>", dur, 2*refTime)
		}
	})
}

func BenchmarkTimerPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tmr := helpers.GlobalTimerPool.Get(time.Millisecond)
		<-tmr.C
		helpers.GlobalTimerPool.Put(tmr)
	}
}

func BenchmarkTimer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tmr := time.NewTimer(time.Millisecond)
		<-tmr.C
		tmr.Stop()
	}
}
