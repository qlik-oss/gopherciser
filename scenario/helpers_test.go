package scenario

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
)

func checkThinkTimeEquality(expected, actual ThinkTimeSettings) error {
	if !reflect.DeepEqual(expected, actual) {
		return errors.Errorf("expected value<%#v>, got value<%#v>", expected, actual)
	}
	return nil
}
func TestSetThinkTimeIfNotSet(t *testing.T) {
	fallback := ThinkTimeSettings{
		helpers.DistributionSettings{
			Type:      helpers.StaticDistribution,
			Delay:     0.111111,
			Deviation: 0.222222,
			Mean:      0.333333,
		},
	}
	cases := []struct {
		name        string
		input       ThinkTimeSettings
		useFallback bool
	}{
		{
			name:        "use fallback, cause default",
			input:       ThinkTimeSettings{},
			useFallback: true,
		},
		{
			name: "use input static",
			input: ThinkTimeSettings{
				helpers.DistributionSettings{
					Type:  helpers.StaticDistribution,
					Delay: 0.0000000000001,
				},
			},
		},
		{
			name: "use input erroneous",
			input: ThinkTimeSettings{
				helpers.DistributionSettings{
					Type:      helpers.StaticDistribution,
					Deviation: 10,
				},
			},
		},
		{
			name: "use input uniform",
			input: ThinkTimeSettings{
				helpers.DistributionSettings{
					Type: helpers.UniformDistribution,
				},
			},
		},
		{
			name: "use input uniform 2",
			input: ThinkTimeSettings{
				helpers.DistributionSettings{
					Type:      helpers.UniformDistribution,
					Mean:      20,
					Deviation: 10,
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := thinkTimeWithFallback(c.input, fallback)
			if c.useFallback {
				if err := checkThinkTimeEquality(result, fallback); err != nil {
					t.Error(errors.Wrap(err, "not using fallback when should have"))
				}
			} else {
				if err := checkThinkTimeEquality(c.input, result); err != nil {
					t.Error(errors.Wrap(err, "not using input when should have"))
				}
			}
		})
	}
}
