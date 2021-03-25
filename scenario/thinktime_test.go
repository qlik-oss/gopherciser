package scenario

import (
	"testing"

	"github.com/qlik-oss/gopherciser/helpers"
)

func TestMarshal(t *testing.T) {
	t.Parallel()

	//static timer
	settings := ThinkTimeSettings{
		helpers.DistributionSettings{
			Type:  helpers.StaticDistribution,
			Delay: 1.0,
		},
	}
	if thinktime, err := jsonit.Marshal(settings); err != nil {
		t.Error(err)
	} else {
		validateString(t, "static think time", string(thinktime), `{"type":"static","delay":1}`)
	}

	//uniform timer
	settings = ThinkTimeSettings{
		helpers.DistributionSettings{
			Type:      helpers.UniformDistribution,
			Mean:      12.5,
			Deviation: 2.5,
		},
	}
	if thinktime, err := jsonit.Marshal(settings); err != nil {
		t.Error(err)
	} else {
		validateString(t, "uniform think time", string(thinktime), `{"type":"uniform","mean":12.5,"dev":2.5}`)
	}
}

func TestNegative(t *testing.T) {
	t.Parallel()

	badraw := `{
		"type" : "badtype"
	}`
	_, err := unmarshal(t, badraw)
	validateError(t, err, `scenario.ThinkTimeSettings.DistributionSettings: Type: unmarshalerDecoder: Failed to unmarshal DistributionType: Unknown enum<badtype>: Key<badtype> not found, error found in #10 byte of ...| "badtype"
	}|..., bigger context ...|{
		"type" : "badtype"
	}|...`)

	settings := ThinkTimeSettings{
		helpers.DistributionSettings{
			Type: 300,
		},
	}
	_, err = jsonit.Marshal(settings)
	validateError(t, err, "scenario.ThinkTimeSettings.DistributionSettings: Type: Unknown DistributionType<300>")

	_, err = settings.Validate()
	validateError(t, err, "distribution type<300> not supported")

	settings.Type = helpers.StaticDistribution
	_, err = settings.Validate()
	validateError(t, err, "Illegal static distribution value")

	settings.Type = helpers.UniformDistribution
	_, err = settings.Validate()
	validateError(t, err, "uniform distribution requires a (positive) mean value defined")

	settings.Mean = 1.0
	_, err = settings.Validate()
	validateError(t, err, "uniform distribution requires a (positive) deviation defined")

	settings.Deviation = 2.0
	_, err = settings.Validate()
	validateError(t, err, "uniform distribution requires a mean value<1.000000> greater than the deviation<2.000000>")
}

// *** Helpers ***

func unmarshal(t *testing.T, raw string) (*ThinkTimeSettings, error) {
	t.Helper()

	var settings ThinkTimeSettings
	if err := jsonit.Unmarshal([]byte(raw), &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}
