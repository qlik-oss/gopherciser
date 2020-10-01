package main

import (
	"go/format"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestCompile(t *testing.T) {
	templateFile := "../data/documentation.template"
	dataRoot := "testdata/base/data"
	expectedOutput := "testdata/base/expected-output/documentation.go"

	generatedDocs := compile(dataRoot, templateFile)
	expectedDocs, err := ioutil.ReadFile(expectedOutput)
	if err != nil {
		t.Fatal(err)
	}

	// maybe check format
	formattedDocs, err := format.Source(generatedDocs)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(generatedDocs, formattedDocs) {
		t.Error("generated docs do not follow gofmt")
	}

	if !reflect.DeepEqual(generatedDocs, expectedDocs) {
		t.Error("generated docs were not correct")
	}
}
