package scenario

import (
	"encoding/json"
	"testing"
)

func TestHasDeprecatedFields(t *testing.T) {
	type args struct {
		rawJson         json.RawMessage
		deprecatedPaths []string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "One non existing",
			args: args{
				rawJson:         []byte(`{"param1" : "val1","param2" : false,"param3" : false,"param4" : 0.1}`),
				deprecatedPaths: []string{"/param5"},
			},
			wantErr: false,
		},
		{
			name: "One existing",
			args: args{
				rawJson:         []byte(`{"param1" : "val1","param2" : false,"param3" : false,"param4" : 0.1}`),
				deprecatedPaths: []string{"/param3"},
			},
			wantErr: true,
		},
		{
			name: "Multiple non existing",
			args: args{
				rawJson:         []byte(`{"param1" : "val1","param2" : false,"param3" : false,"param4" : 0.1}`),
				deprecatedPaths: []string{"/param7", "/param6", "/param5"},
			},
			wantErr: false,
		},
		{
			name: "Multiple existing",
			args: args{
				rawJson:         []byte(`{"param1" : "val1","param2" : false,"param3" : false,"param4" : 0.1}`),
				deprecatedPaths: []string{"/param2", "/param3", "/param4"},
			},
			wantErr: true,
		},
		{
			name: "Multiple mixed",
			args: args{
				rawJson:         []byte(`{"param1" : "val1","param2" : false,"param3" : false,"param4" : 0.1}`),
				deprecatedPaths: []string{"/param1", "/param6", "/param5"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := HasDeprecatedFields(tt.args.rawJson, tt.args.deprecatedPaths); (err != nil) != tt.wantErr {
				t.Errorf("HasDeprecatedFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
