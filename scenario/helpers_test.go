package scenario

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
)

func checkThinkTimeEquality(expected, actual *ThinkTimeSettings) error {
	if expected != actual || !reflect.DeepEqual(expected, actual) {
		return errors.Errorf("expected address<%p> value<%#v>, got address<%p> value<%#v>", expected, expected, actual, actual)
	}
	return nil
}
func TestSetThinkTimeIfNotSet(t *testing.T) {
	shallBeInternalSetThinkTimeError := func(err error) error {
		if err != internalSetThinkTimeError {
			return errors.Errorf("expected error %s, got error %s", internalSetThinkTimeError, err)
		}
		return nil
	}
	defaultThinkTime := &ThinkTimeSettings{}
	var nilThinkTime *ThinkTimeSettings
	nonDefaultThinkTime := &ThinkTimeSettings{
		helpers.DistributionSettings{
			Type:  helpers.StaticDistribution,
			Delay: 0.0000000000001,
		},
	}
	cases := []struct {
		name     string
		input    **ThinkTimeSettings
		fallback *ThinkTimeSettings
		checkErr func(actualErr error) error
		shallSet bool
	}{
		{
			name:     "nil input",
			input:    nil,
			fallback: &askHubAdvisorDefaultThinktimeSettings,
			checkErr: shallBeInternalSetThinkTimeError,
		},
		{
			name:     "use fallback, cause default",
			input:    &defaultThinkTime,
			fallback: &askHubAdvisorDefaultThinktimeSettings,
			shallSet: true,
		},
		{
			name:     "use fallback, cause nil",
			input:    &nilThinkTime,
			fallback: &askHubAdvisorDefaultThinktimeSettings,
			shallSet: true,
		},
		{
			name:     "use input",
			input:    &nonDefaultThinkTime,
			fallback: &askHubAdvisorDefaultThinktimeSettings,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var inputBefore *ThinkTimeSettings
			if c.input != nil {
				inputBefore = *c.input
			}
			setThinkTimeErr := setThinkTimeIfNotSet(c.input, c.fallback)
			if c.checkErr != nil {
				if err := c.checkErr(setThinkTimeErr); err != nil {
					t.Error(err)
				}
			}
			if setThinkTimeErr != nil {
				return
			}
			if c.shallSet {
				if err := checkThinkTimeEquality(*c.input, c.fallback); err != nil {
					t.Error(errors.Wrap(err, "not using fallback when should have"))
				}
			} else {
				if err := checkThinkTimeEquality(*c.input, inputBefore); err != nil {
					t.Error(errors.Wrap(err, "not using input when should have"))
				}
			}

		})
	}
}
