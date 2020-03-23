package helpers

import (
	"testing"

	"github.com/qlik-oss/gopherciser/randomizer"
)

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

	rndCompare := randomizer.NewSeededRandomizer(randomizer.GetPredictableSeed(1, 1))

	sample, err := settings.GetSample(rndCompare)
	if err != nil {
		t.Fatal(err)
	}

	if sample > 0.6 {
		t.Fatalf("Sample: Expected< < 0.6> got<%f>", sample)
	}

	if sample < 0.4 {
		t.Fatalf("Sample: Expected< > 0.4> got<%f>", sample)
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

	sample, err := settings.GetSample(nil)
	if err != nil {
		t.Fatal(err)
	}

	if sample != 0.1 {
		t.Fatalf("Sample: Expected<0.1> got<%f>", sample)
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
