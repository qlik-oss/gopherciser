package main

import (
	"encoding/json"
	"go/format"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/qlik-oss/gopherciser/generatedocs/common"
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

func TestOverload(t *testing.T) {
	for _, tc := range []struct {
		name     string
		base     *Data
		new      *Data
		expected *Data
	}{
		{
			name: "simple overload",
			base: &Data{
				ParamMap: map[string][]string{
					"param1": {""},
				},
				Groups: []common.GroupsEntry{
					{
						Name:    "group1",
						Title:   "Group 1",
						Actions: []string{"action1"},
						DocEntry: common.DocEntry{
							Description: "",
							Examples:    "",
						},
					},
				},
				Actions: []string{"action1"},
				ActionMap: map[string]common.DocEntry{
					"action1": common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
				ConfigFields: []string{"configField1"},
				ConfigMap: map[string]common.DocEntry{
					"configField1": common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
				Extra: []string{"extra1"},
				ExtraMap: map[string]common.DocEntry{
					"extra1": common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
			},
			new: &Data{
				ParamMap: map[string][]string{
					"param2": {""},
				},
				Groups: []common.GroupsEntry{
					{
						Name:     "group2",
						Title:    "Group 2",
						Actions:  []string{"action2"},
						DocEntry: common.DocEntry{},
					},
				},
				Actions: []string{"action2"},
				ActionMap: map[string]common.DocEntry{
					"action2": common.DocEntry{},
				},
				ConfigFields: []string{"configField2"},
				ConfigMap: map[string]common.DocEntry{
					"configField2": common.DocEntry{},
				},
				Extra: []string{"extra2"},
				ExtraMap: map[string]common.DocEntry{
					"extra2": common.DocEntry{},
				},
			},
			expected: &Data{
				ParamMap: map[string][]string{
					"param1": {""},
					"param2": {""},
				},
				Groups: []common.GroupsEntry{
					{
						Name:     "group1",
						Title:    "Group 1",
						Actions:  []string{"action1"},
						DocEntry: common.DocEntry{},
					},
					{
						Name:     "group2",
						Title:    "Group 2",
						Actions:  []string{"action2"},
						DocEntry: common.DocEntry{},
					},
				},
				Actions: []string{"action1", "action2"},
				ActionMap: map[string]common.DocEntry{
					"action1": common.DocEntry{},
					"action2": common.DocEntry{},
				},
				ConfigFields: []string{"configField1", "configField2"},
				ConfigMap: map[string]common.DocEntry{
					"configField1": common.DocEntry{},
					"configField2": common.DocEntry{},
				},
				Extra: []string{"extra1", "extra2"},
				ExtraMap: map[string]common.DocEntry{
					"extra1": common.DocEntry{},
					"extra2": common.DocEntry{},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.base.overload(tc.new)
			if !reflect.DeepEqual(tc.base, tc.expected) {
				t.Log(objDiff(tc.base, tc.expected))
				t.Fatalf("overload gave unexpected result")
			}
		})

	}
}

func objDiff(obj1, obj2 interface{}) string {
	return diff.LineDiff(pretty(obj1), pretty(obj2))
}

func pretty(i interface{}) string {
	bytes, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
