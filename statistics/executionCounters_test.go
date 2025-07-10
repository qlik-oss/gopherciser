package statistics

import (
	"sync"
	"testing"

	"github.com/qlik-oss/gopherciser/helpers"
)

type collectTestDefSamples struct {
	Path  string
	Count uint64
	Size  uint64
}

type collectTestDef struct {
	Name    string
	Samples []collectTestDefSamples
}

func TestCounters(t *testing.T) {
	testDef := []collectTestDef{
		{
			Name: "1 path 10148 samples",
			Samples: []collectTestDefSamples{
				{
					Count: 10148,
					Path:  "/test/path/1",
					Size:  123456,
				},
			},
		},
		{
			Name: "2 path 2048 samples",
			Samples: []collectTestDefSamples{
				{
					Count: 1024,
					Path:  "/test/path/1",
					Size:  12345,
				},
				{
					Count: 1024,
					Path:  "/test/path/2",
					Size:  12345678,
				},
			},
		},
	}

	for _, def := range testDef {
		t.Run(def.Name, func(t *testing.T) {
			collectorTest(t, def)
		})
	}
}

func collectorTest(t *testing.T, def collectTestDef) {
	collector := NewCollector()
	collector.SetLevel(StatsLevelFull)

	if collector.Level != StatsLevelFull {
		t.Errorf("incorrect stats level: %v", collector.Level)
		return
	}

	if !collector.IsOn() {
		t.Errorf("collector not on")
		return
	}

	var wg sync.WaitGroup

	for _, sample := range def.Samples {
		wg.Add(1)
		go func() {
			defer wg.Done()

			stats := collector.GetOrAddRequestStats("head", sample.Path)
			if stats == nil {
				t.Errorf("stats returned nil for sample<%v>", sample)
				return
			}

			for range sample.Count {
				stats.RespAvg.AddSample(sample.Size)
			}

			average, requests := stats.RespAvg.Average()
			if !helpers.NearlyEqual(requests, float64(sample.Count)) {
				t.Errorf("path<%s> expected sample count<%d> got<%f>", sample.Path, sample.Count, requests)
			}
			if !helpers.NearlyEqual(average, float64(sample.Size)) {
				t.Errorf("path<%s> expected average<%d> got<%f>", sample.Path, sample.Size, average)
			}
		}()
	}

	wg.Wait()
}
