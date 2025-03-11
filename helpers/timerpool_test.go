package helpers_test

import (
	"testing"
	"time"

	"github.com/qlik-oss/gopherciser/helpers"
)

func TestTimerPool(t *testing.T) {
	refTime := 50 * time.Millisecond

	t1 := helpers.GlobalTimerPool.Get(refTime)
	helpers.GlobalTimerPool.Put(t1)

	t1 = helpers.GlobalTimerPool.Get(2 * refTime)
	ts := time.Now()
	<-t1.C
	dur := time.Since(ts)
	if dur < 2*refTime {
		t.Fatal("timer ticked to fast")
	}
	if dur > time.Duration(1.1*float64(2*refTime)) {
		t.Fatal("timer took more than 10% longer than expected")
	}
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
