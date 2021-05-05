package config

import "testing"

func TestValidator(t *testing.T) {
	tests := []struct {
		Value       interface{}
		stringVal   string
		Type        ValidationType
		expectError bool
	}{
		{7.0, "7", ValidationTypeNumber, false},
		{7.0, "7.00", ValidationTypeNumber, false},
		{6.00000000014, "6.0000000015", ValidationTypeNumber, false},
		{6.000000014, "6.00000015", ValidationTypeNumber, true},
		{true, "true", ValidationTypeBool, false},
		{true, "false", ValidationTypeBool, true},
		{true, "11", ValidationTypeBool, true},
		{true, "1", ValidationTypeBool, false},
		{"stuff", "stuff", ValidationTypeString, false},
		{"stuff", "stuff ", ValidationTypeString, true},
	}

	for _, test := range tests {
		err := cmp(test.Value, test.stringVal, test.Type)
		switch test.expectError {
		case true:
			if err == nil {
				t.Errorf("value<%f> and value<%s> compared equal but expected error", test.Value, test.stringVal)
			}
		case false:
			if err != nil {
				t.Errorf("value<%f> and value<%s> compared with error %v", test.Value, test.stringVal, err)
			}
		}
	}
}

func cmp(a interface{}, b string, validationType ValidationType) error {
	validator := &Validator{
		ValidatorCore{
			Type:  validationType,
			Level: FailLevelError,
			Value: a,
		},
	}

	return validator.Validate(b)
}
