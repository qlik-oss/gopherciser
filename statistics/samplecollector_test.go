package statistics

import (
	"fmt"
	"testing"
)

func TestSampleCollector_Average(t *testing.T) {
	t.Parallel()

	t.Run("small sample set", func(t *testing.T) {
		t.Parallel()
		collector := generateSamples(50)
		if err := validateAverage(collector); err != nil {
			t.Error(err.Error())
		}
	})
	t.Run("medium sample set", func(t *testing.T) {
		t.Parallel()
		collector := generateSamples(500)
		if err := validateAverage(collector); err != nil {
			t.Error(err.Error())
		}
	})
	t.Run("large sample set", func(t *testing.T) {
		t.Parallel()
		collector := generateSamples(10000000)
		if err := validateAverage(collector); err != nil {
			t.Error(err.Error())
		}
	})
}

func generateSamples(size int) *SampleCollector {
	collector := NewSampleCollector()
	for i := 0; i < size; i++ {
		collector.AddSample(12)
	}
	return collector
}

func validateAverage(collector *SampleCollector) error {
	a, _ := collector.Average()
	aString := fmt.Sprintf("%.2f", a) // two decimals due to possible float
	if aString != "12.00" {
		return fmt.Errorf("average<%s> not expected<12.00>", aString)
	}
	return nil
}
