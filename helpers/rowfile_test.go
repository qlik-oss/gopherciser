package helpers_test

import (
	"github.com/goccy/go-json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/qlik-oss/gopherciser/helpers"
)

func Test_RowFile_Marshal(t *testing.T) {
	dir, err := ioutil.TempDir("", "rowfile")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if dir != "" && dir != "/" {
			if err := os.Remove(dir); err != nil {
				t.Fatal(err)
			}
		}
	}()

	filepath := fmt.Sprintf("%s/rowfile.txt", dir)
	defer func() {
		if err := os.Remove(filepath); err != nil {
			t.Fatal(err)
		}
	}()

	testData := []string{
		"row1",
		"row2",
		"row3",
	}

	// prepare test file
	if err := ioutil.WriteFile(filepath, []byte(strings.Join(testData, "\n")), 0644); err != nil {
		t.Fatal(err)
	}

	js := []byte(`{"filepath":"` + filepath + `"}`)

	t.Log(string(js))

	fakeStruct := struct {
		RowFile helpers.RowFile `json:"filepath"`
	}{}

	if err := json.Unmarshal(js, &fakeStruct); err != nil {
		t.Fatal(err)
	}

	rowFileTest(t, fakeStruct.RowFile, testData)

	mar, err := json.Marshal(fakeStruct)
	if err != nil {
		t.Fatal(err)
	}
	if string(mar) != string(js) {
		t.Errorf("marshal result not as expected\nExpected:\n%s\nGot:\n%s\n", js, mar)
	}

	rf, err := helpers.NewRowFile(filepath)
	if err != nil {
		t.Fatal(err)
	}

	rowFileTest(t, rf, testData)
}

func rowFileTest(t *testing.T, rowFile helpers.RowFile, testData []string) {
	t.Helper()
	if len(rowFile.Rows()) != len(testData) {
		t.Fatalf("rows<%d> and testdata<%d>", len(rowFile.Rows()), len(testData))
	}

	for i, row := range rowFile.Rows() {
		if row != testData[i] {
			t.Errorf("row<%d> not expected<%s> got<%s>", i, testData[i], row)
		}
	}
}
