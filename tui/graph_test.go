package tui

import (
	"testing"
	"time"

	"github.com/qlik-oss/gopherciser/statistics"
)

func TestNiceCeil(t *testing.T) {
	cases := map[float64]float64{0: 0, 0.4: 1, 1: 1, 3: 5, 7: 10, 12: 20, 45: 50, 51: 100, 1500: 2000}
	for value, want := range cases {
		if got := niceCeil(value); got != want {
			t.Errorf("niceCeil(%v) = %v, want %v", value, got, want)
		}
	}
}

func TestFormatValue(t *testing.T) {
	cases := []struct {
		value float64
		rate  bool
		want  string
	}{{0, false, "0"}, {2.5, true, "2.5"}, {2.4, false, "2"}, {12345, false, "12.3k"}, {2500000, false, "2.5M"}}
	for _, c := range cases {
		if got := formatValue(c.value, c.rate); got != c.want {
			t.Errorf("formatValue(%v, %v) = %q, want %q", c.value, c.rate, got, c.want)
		}
	}
}

func TestGraphPoints(t *testing.T) {
	g := newGraphView()
	counters := &statistics.ExecutionCounters{}
	start := time.Now()
	for i := 0; i < 5; i++ {
		counters.Sessions.Add(2)
		counters.Requests.Add(uint64(10 * i))
		g.record(counters, start.Add(time.Duration(i)*time.Second))
	}

	sessions := g.points(1, 3)
	if len(sessions) != 3 || sessions[0] != 6 || sessions[2] != 10 {
		t.Errorf("unexpected session points %v", sessions)
	}

	// requests grow by 10, 20, 30, 40 between samples, rate is the per second delta
	rates := g.points(2, 10)
	if len(rates) != 4 || rates[0] != 10 || rates[3] != 40 {
		t.Errorf("unexpected request rates %v", rates)
	}
}
