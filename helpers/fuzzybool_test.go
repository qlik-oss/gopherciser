package helpers_test

import (
	"fmt"
	"testing"

	"github.com/qlik-oss/gopherciser/helpers"
)

func TestFuzzyBool_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		args    []byte
		wantErr bool
		wantSb  helpers.FuzzyBool
	}{
		{
			name:    "Test bool false",
			args:    []byte(`false`),
			wantErr: false,
			wantSb:  false,
		},
		{
			name:    "Test bool true",
			args:    []byte(`true`),
			wantErr: false,
			wantSb:  true,
		},
		{
			name:    "Test string false",
			args:    []byte(`"false"`),
			wantErr: false,
			wantSb:  false,
		},
		{
			name:    "Test string NaN",
			args:    []byte(`"NaN"`),
			wantErr: false,
			wantSb:  true,
		},
		{
			name:    "Test string 0",
			args:    []byte(`"0"`),
			wantErr: false,
			wantSb:  false,
		},
		{
			name:    "Test int 0",
			args:    []byte(`0`),
			wantErr: false,
			wantSb:  false,
		},
		{
			name:    "Test int 1",
			args:    []byte(`1`),
			wantErr: false,
			wantSb:  true,
		},
		{
			name:    "Test int -1",
			args:    []byte(`-1`),
			wantErr: false,
			wantSb:  true,
		},
		{
			name:    "Test float 0.0",
			args:    []byte(`0.0`),
			wantErr: false,
			wantSb:  false,
		},
		{
			name:    fmt.Sprintf("Test float %f", helpers.DefaultEpsilon),
			args:    []byte(fmt.Sprintf("%f", helpers.DefaultEpsilon)),
			wantErr: false,
			wantSb:  false,
		},
		{
			name:    "Test float 1.0",
			args:    []byte(`1.0`),
			wantErr: false,
			wantSb:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := helpers.FuzzyBool(false)
			if err := (&sb).UnmarshalJSON(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("FuzzyBool.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if sb != tt.wantSb {
				t.Errorf("FuzzyBool.UnmarshalJSON() sb value = %v, want %v", sb, tt.wantSb)
			}
		})
	}
}
