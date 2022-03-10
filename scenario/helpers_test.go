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
	defaultThinkTime := &ThinkTimeSettings{}
	var nilThinkTime *ThinkTimeSettings
	nonDefaultThinkTime := &ThinkTimeSettings{
		helpers.DistributionSettings{
			Type:  helpers.StaticDistribution,
			Delay: 0.0000000000001,
		},
	}
	cases := []struct {
		name        string
		input       *ThinkTimeSettings
		fallback    ThinkTimeSettings
		useFallback bool
	}{
		{
			name:        "use fallback, cause default",
			input:       defaultThinkTime,
			fallback:    askHubAdvisorDefaultThinktimeSettings,
			useFallback: true,
		},
		{
			name:        "use fallback, cause nil",
			input:       nilThinkTime,
			fallback:    askHubAdvisorDefaultThinktimeSettings,
			useFallback: true,
		},
		{
			name:     "use input",
			input:    nonDefaultThinkTime,
			fallback: askHubAdvisorDefaultThinktimeSettings,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := thinkTimeWithFallback(c.input, c.fallback)
			if result == c.input {
				t.Errorf("return and argument has the same address: %p == %p>", result, c.input)
			}
			if c.useFallback {
				if err := checkThinkTimeEquality(*result, c.fallback); err != nil {
					t.Error(errors.Wrap(err, "not using fallback when should have"))
				}
			} else {
				if err := checkThinkTimeEquality(*result, *result); err != nil {
					t.Error(errors.Wrap(err, "not using input when should have"))
				}
			}

		})
	}
}
