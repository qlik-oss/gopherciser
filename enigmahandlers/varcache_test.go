package enigmahandlers

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/qlik-oss/enigma-go"
)

var varDummy = &enigma.GenericVariable{}

func Test_varCache_Lookup(t *testing.T) {
	names := [...]string{"variable1", "variable2", "variable3"}

	preFilledVarCache := &varCache{
		varMap: map[string]*enigma.GenericVariable{
			names[0]: varDummy,
			names[1]: varDummy,
			names[2]: varDummy,
		},
	}

	tests := []struct {
		name         string
		variableName string
		wantVariable *enigma.GenericVariable
		wantHit      bool
		vc           *varCache
	}{
		{"hit1", names[0], varDummy, true, preFilledVarCache},
		{"hit2", names[1], varDummy, true, preFilledVarCache},
		{"noHit1", names[2], nil, false, nil},
		{"noHit2", "", nil, false, preFilledVarCache},
		{"noHit3", "abc", nil, false, preFilledVarCache},
		{"noHit4", "", nil, false, preFilledVarCache},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.vc == nil {
				newFC := NewVarCache()
				tt.vc = &newFC
			}
			gotVariable, gotHit := tt.vc.Lookup(tt.variableName)
			if !reflect.DeepEqual(gotVariable, tt.wantVariable) {
				t.Errorf("varCache.Lookup() gotVariable = %v, want %v", gotVariable, tt.wantVariable)
			}
			if gotHit != tt.wantHit {
				t.Errorf("varCache.Lookup() gotHit = %v, want %v", gotHit, tt.wantHit)
			}
		})
	}
}

func Test_varCache_LookupWithFallback(t *testing.T) {
	fallBackMock := func(name string) (*enigma.GenericVariable, error) {
		switch name {
		case "variable1", "variable2", "variable3": // existing variables
			return varDummy, nil
		default:
			return nil, fmt.Errorf(`"%s" is not the name of a variable`, name)
		}
	}
	tests := []struct {
		name         string
		variableName string
		want         *enigma.GenericVariable
		wantErr      bool
	}{
		{"shallExist1", "variable1", varDummy, false},
		{"shallExist2", "variable1", varDummy, false},
		{"shallExist3", "variable3", varDummy, false},
		{"shallFail1", "", nil, true},
		{"shallFail2", "xyz", nil, true},
	}
	vc := NewVarCache()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vc.LookupWithFallback(tt.variableName, fallBackMock)
			if (err != nil) != tt.wantErr {
				t.Errorf("varCache.LookupWithFallback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("varCache.LookupWithFallback() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_varCache_Store(t *testing.T) {
	tests := []struct {
		name           string
		storeNames     []string
		notStoredNames []string
	}{
		{"storeTest1", []string{"variable1", "variable2", "variable3"}, []string{"", "xyz"}},
		{"storeTest1", []string{"variable1", "variable2"}, []string{"", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vc := NewVarCache()
			for _, variableName := range tt.storeNames {
				vc.Store(variableName, varDummy)
			}
			wantLen := len(tt.storeNames)
			gotLen := len(vc.varMap)
			if wantLen != gotLen {
				t.Errorf("varCache.Store() unexpected cache size")
			}
			for _, variableName := range tt.storeNames {
				v, hit := vc.varMap[variableName]
				if !hit {
					t.Errorf("varCache.Store() want hit==true")
				}
				if v == nil {
					t.Errorf("varCache.Store() want variable!=nil")
				}
			}
			for _, variableName := range tt.notStoredNames {
				v, hit := vc.varMap[variableName]
				if hit {
					t.Errorf("varCache.Store() want hit==false")
				}
				if v != nil {
					t.Errorf("varCache.Store() want variable==nil")
				}
			}
		})
	}
}
