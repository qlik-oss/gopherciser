package enigmahandlers

import (
	"reflect"
	"testing"

	"github.com/qlik-oss/enigma-go/v4"
)

var fieldDummy = &enigma.Field{}

func Test_fieldCache_Lookup(t *testing.T) {
	names := [...]string{"field1", "field2", "field3"}

	preFilledFieldCache := &FieldCache{
		fieldMap: map[string]*enigma.Field{
			names[0]: fieldDummy,
			names[1]: fieldDummy,
			names[2]: fieldDummy,
		},
	}

	tests := []struct {
		name      string
		fieldName string
		wantField *enigma.Field
		wantHit   bool
		fc        *FieldCache
	}{
		{"hit1", names[0], fieldDummy, true, preFilledFieldCache},
		{"hit2", names[1], fieldDummy, true, preFilledFieldCache},
		{"noHit1", names[2], nil, false, nil},
		{"noHit2", "", nil, false, preFilledFieldCache},
		{"noHit3", "abc", nil, false, preFilledFieldCache},
		{"noHit4", "", nil, false, preFilledFieldCache},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fc == nil {
				newFC := NewFieldCache()
				tt.fc = &newFC
			}
			gotField, gotHit := tt.fc.Lookup(tt.fieldName)
			if !reflect.DeepEqual(gotField, tt.wantField) {
				t.Errorf("fieldCache.Lookup() gotField = %v, want %v", gotField, tt.wantField)
			}
			if gotHit != tt.wantHit {
				t.Errorf("fieldCache.Lookup() gotHit = %v, want %v", gotHit, tt.wantHit)
			}
		})
	}
}

func Test_fieldCache_Store(t *testing.T) {
	tests := []struct {
		name           string
		storeNames     []string
		notStoredNames []string
	}{
		{"storeTest1", []string{"field1", "field2", "field3"}, []string{"", "xyz"}},
		{"storeTest1", []string{"field1", "field2"}, []string{"", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFieldCache()
			for _, fieldName := range tt.storeNames {
				fc.Store(fieldName, fieldDummy)
			}
			wantLen := len(tt.storeNames)
			gotLen := len(fc.fieldMap)
			if wantLen != gotLen {
				t.Errorf("fieldCache.Store() unexpected cache size")
			}
			for _, fieldName := range tt.storeNames {
				f, hit := fc.fieldMap[fieldName]
				if !hit {
					t.Errorf("fieldCache.Store() want hit==true")
				}
				if f == nil {
					t.Errorf("fieldCache.Store() want field!=nil")
				}
			}
			for _, fieldName := range tt.notStoredNames {
				f, hit := fc.fieldMap[fieldName]
				if hit {
					t.Errorf("fieldCache.Store() want hit==false")
				}
				if f != nil {
					t.Errorf("fieldCache.Store() want field==nil")
				}
			}
		})
	}
}
