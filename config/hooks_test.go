package config

import (
	"testing"

	"github.com/qlik-oss/gopherciser/synced"
)

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
			value: a,
		},
	}

	return validator.ValidateValue(b)
}

func TestHookData(t *testing.T) {
	cfg, err := NewEmptyConfig()
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := synced.New("{ \"text\": \"Test finished with {{ .Counters.Errors }} errors and {{ .Counters.Warnings }} warnings. Total Sessions: {{ .Counters.Sessions }}\"}")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Hooks.Pre = &Hook{
		HookCore: HookCore{
			Content: *tmpl,
		},
	}

	errors := cfg.Counters.Errors.Inc()
	if errors != uint64(1) {
		t.Errorf("expected error counter to be 1, was %d", errors)
	}

	warnings := cfg.Counters.Warnings.Add(3)
	if warnings != uint64(3) {
		t.Errorf("expected warning counter to be 3, was %d", warnings)
	}

	cfg.PopulateHookData()
	warn, err := cfg.Hooks.Validate()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("warnings:", warn)
	payload, err := cfg.Hooks.Pre.Content.ExecuteString(cfg.Hooks.data)
	if err != nil {
		t.Fatal(err)
	}
	expectedPayload := `{ "text": "Test finished with 1 errors and 3 warnings. Total Sessions: 0"}`

	if payload != expectedPayload {
		t.Errorf("payload was<%s> expected<%s>", payload, expectedPayload)
	}
}
