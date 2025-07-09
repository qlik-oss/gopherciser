package statistics

import (
	"math"
	"testing"
)

func TestCounters(t *testing.T) {
	collector := NewCollector()
	collector.SetLevel(StatsLevelFull)
	t.Log("collector level is full:", collector.IsFull())
	t.Log("collector is on:", collector.IsOn())

	if collector.Level != StatsLevelFull {
		t.Fatalf("incorrect stats level: %v", collector.Level)
	}

	if !collector.IsOn() {
		t.Fatalf("collector not on")
	}

	n := 10148
	var stats *RequestStats
	for range n {
		stats = collector.GetOrAddRequestStats("head", "/test/path/1")
		if stats == nil {
			t.Fatal("stats returned nil")
		}
		stats.RespAvg.AddSample(uint64(123456))
	}

	_, requests := stats.RespAvg.Average()
	current := uint64(math.Round(requests))
	expected := uint64(n)
	if current != expected {
		t.Errorf("expected<%d> got<%d>", expected, current)
	}
}
