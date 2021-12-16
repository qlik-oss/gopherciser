package helpers

import (
	"github.com/goccy/go-json"
	"fmt"
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

type LookupAndSetResult struct {
	Node struct {
		Intval    int
		Floatval  float64
		Stringval string
	}
}

func TestLookupAndSet(t *testing.T) {
	data := []byte(`{
	"node" : {
		"intval" : 2,
		"floatval": 4.2,
		"stringval": "string1"
	}
}`)
	t.Logf("data: %s", data)

	tests := []struct {
		Path          DataPath
		NewData       []byte
		ExpectedValue interface{}
	}{
		{
			Path:          DataPath("node/intval"),
			NewData:       []byte("4"),
			ExpectedValue: 4,
		},
		{
			Path:          DataPath("node/floatval"),
			NewData:       []byte("5.2"),
			ExpectedValue: 5.2,
		},
		{
			Path:          DataPath("node/stringval"),
			NewData:       []byte(`"yadda"`),
			ExpectedValue: "yadda",
		},
		{
			Path:          DataPath("new/path/novaluehere"),
			NewData:       []byte("404"),
			ExpectedValue: 404,
		},
		{
			Path:          DataPath("nodenode"),
			NewData:       []byte(`{"newkey":"newvalue"}`),
			ExpectedValue: `{"newkey":"newvalue"}`,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("TestLookupAndSet%d", i), func(t *testing.T) {
			// Lookup and modify
			t.Logf("%s;%s;%v", test.Path, string(test.NewData), test.ExpectedValue)
			modifiedData, err := test.Path.Set(data, test.NewData)
			if len(modifiedData) > 0 {
				t.Logf("modified data: %s", modifiedData)
			}
			if err != nil {
				t.Fatal(err)
			}

			// unmarshal results
			var result LookupAndSetResult
			err = json.Unmarshal(modifiedData, &result)
			if err != nil {
				t.Fatal(err)
			}

			// validate data
			newVal, err := test.Path.LookupNoQuotes(modifiedData)
			if err != nil {
				t.Fatal(err)
			}

			// compare values as strings
			newValueAsString := string(newVal)
			expectedValueAsString := fmt.Sprintf("%v", test.ExpectedValue)
			if newValueAsString != expectedValueAsString {
				t.Errorf("unexpected result, got<%s> expected<%s>", newValueAsString, expectedValueAsString)
			}
		})
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
