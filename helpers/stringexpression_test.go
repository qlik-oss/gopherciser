package helpers_test

import (
	"testing"

	"github.com/qlik-oss/gopherciser/helpers"
)

func TestStringExpression_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		arg        []byte
		wantErr    bool
		wantResult string
	}{
		{
			name:       "Quoted string",
			arg:        []byte(`"myexpr1"`),
			wantErr:    false,
			wantResult: "myexpr1",
		},
		{
			name:       "Unquoted string",
			arg:        []byte(`myexpr2"`),
			wantErr:    false,
			wantResult: "myexpr2\"",
		},
		{
			name:       "Unquoted string with quotes",
			arg:        []byte(`\"myexpr2\"`),
			wantErr:    false,
			wantResult: `\"myexpr2\"`,
		},
		{
			name:       "Expression object",
			arg:        []byte(`{"qStringExpression":{"qExpr":"myexpr3"}}`),
			wantErr:    false,
			wantResult: "myexpr3",
		},
		{
			name:       "empty string",
			arg:        []byte{},
			wantErr:    false,
			wantResult: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var expr helpers.StringExpression
			gotErr := expr.UnmarshalJSON(tt.arg)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UnmarshalJSON() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Error("UnmarshalJSON() succeeded unexpectedly")
			} else if expr != helpers.StringExpression(tt.wantResult) {
				t.Errorf("Unexpected result<%s> wanted<%s>", expr, tt.wantResult)
			}
		})
	}
}

func TestStringExpression_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		input   helpers.StringExpression
		want    []byte
		wantErr bool
	}{
		{
			name:    "string",
			input:   helpers.StringExpression("mystring"),
			want:    []byte("mystring"),
			wantErr: false,
		},
		{
			name:    "expr with =",
			input:   helpers.StringExpression("='myexpr1'"),
			want:    []byte(`{"qStringExpression":{"qExpr":"='myexpr1'"}}`),
			wantErr: false,
		},
		{
			name:    "expr without =",
			input:   helpers.StringExpression("'myexpr2'"),
			want:    []byte(`{"qStringExpression":{"qExpr":"'myexpr2'"}}`),
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   helpers.StringExpression(""),
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.input.MarshalJSON()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("MarshalJSON() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("MarshalJSON() succeeded unexpectedly")
			}
			if string(tt.want) != string(got) {
				t.Errorf("MarshalJSON() result<%s>, want<%s>", got, tt.want)
			}
		})
	}
}
