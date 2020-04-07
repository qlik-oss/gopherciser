package helpers

import (
	"testing"
	"time"

	"github.com/qlik-oss/gopherciser/randomizer"
)

type Rnd struct {
	*randomizer.Randomizer
}

func (rnd *Rnd) Reset(instance, session uint64, onlyinstanceSeed bool) {
	rnd.Randomizer = randomizer.NewSeededRandomizer(randomizer.GetPredictableSeed(int(instance), int(session)))
}

func TestDistributionUniform(t *testing.T) {
	t.Parallel()

	raw := `{
		"type" : "uniform",
		"mean" : 0.5,
		"dev" : 0.1
	}`

	settings, errUnmarshal := unmarshal(t, raw)
	if errUnmarshal != nil {
		t.Fatal(errUnmarshal)
	}

	if settings.Mean != 0.5 {
		t.Fatalf("Mean: Expected<0.5> got<%f>", settings.Mean)
	}

	if settings.Deviation != 0.1 {
		t.Fatalf("Deviation: Expected<0.1> got<%f>", settings.Deviation)
	}

	rndCompare := &Rnd{}
	rndCompare.Reset(1, 1, false)

	sample, err := settings.RandDuration(rndCompare)
	if err != nil {
		t.Fatal(err)
	}

	if sample > time.Duration(0.6*float64(time.Second)) {
		t.Fatalf("Sample: Expected< < 0.6s> got<%v>", sample)
	}

	if sample < time.Duration(0.4*float64(time.Second)) {
		t.Fatalf("Sample: Expected< > 0.4> got<%v>", sample)
	}
}

func TestThinkTimeStatic(t *testing.T) {
	t.Parallel()

	raw := `{
			"type" : "static",
			"delay" : 0.1
		}`

	settings, errUnmarshal := unmarshal(t, raw)
	if errUnmarshal != nil {
		t.Fatal(errUnmarshal)
	}

	if settings.Delay != 0.1 {
		t.Fatalf("Delay: Expected<0.1> got<%f>", settings.Delay)
	}

	sample, err := settings.RandDuration(nil)
	if err != nil {
		t.Fatal(err)
	}

	if sample != time.Duration(0.1*float64(time.Second)) {
		t.Fatalf("Sample: Expected<0.1> got<%v>", sample)
	}
}

// *** Helpers ***

func unmarshal(t *testing.T, raw string) (*DistributionSettings, error) {
	t.Helper()

	var settings DistributionSettings
	if err := jsonit.Unmarshal([]byte(raw), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}
