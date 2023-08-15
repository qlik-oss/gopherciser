package enigmahandlers

import (
	"reflect"
	"testing"

	"github.com/qlik-oss/enigma-go/v4"
)

var varDummy = &enigma.GenericVariable{}

func Test_varCache_Lookup(t *testing.T) {
	names := [...]string{"variable1", "variable2", "variable3"}

	preFilledVarCache := &VarCache{
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
		vc           *VarCache
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
