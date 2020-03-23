package senseobjdef

import (
	"encoding/json"
	"testing"
)

func TestMultiLookup(t *testing.T) {
	data := `{
		"object" : {
			"array" : [
				{
					"subpath" : "value1"
				},
				{
					"subpath" : { "subobject" : "value2" }
				},
				{
					"subpath" : ["value3"]
				}
			]
		}
	}`

	multiPath := DataPath("/object/array/{pos}/subpath")

	rawArray, err := multiPath.LookupMulti(json.RawMessage(data), "{pos}")
	if err != nil {
		t.Fatal(err)
	}

	if len(rawArray) != 3 {
		t.Fatalf("Expected 3 values got %d : %+v", len(rawArray), rawArray)
	}

	valueArray := []string{
		`"value1"`, `{"subobject":"value2"}`, `["value3"]`,
	}

	for i, v := range rawArray {
		if valueArray[i] != string(v) {
			t.Errorf("Array pos<%d> Expected<%s> got<%s>", i, valueArray[i], v)
		}
	}
}

//TODO Both Lookup and MultiLookup needs performance optimizations

func BenchmarkMultiLookup(b *testing.B) {
	data := `{
		"object" : {
			"array" : [
				{
					"subpath" : "value1"
				},
				{
					"subpath" : { "subobject" : "value2" }
				},
				{
					"subpath" : ["value3"]
				}
			]
		}
	}`

	multiPath := DataPath("/object/array/{pos}/subpath")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = multiPath.LookupMulti(json.RawMessage(data), "{pos}")
	}
}

func BenchmarkLookup(b *testing.B) {
	data := `{
		"object" : {
			"array" : [
				{
					"subpath" : "value1"
				},
				{
					"subpath" : { "subobject" : "value2" }
				},
				{
					"subpath" : ["value3"]
				}
			]
		}
	}`

	path := DataPath("/object/array/0/subpath")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = path.Lookup(json.RawMessage(data))
	}
}
