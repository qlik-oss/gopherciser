package doccompiler

import (
	"github.com/goccy/go-json"
	"fmt"
	"go/format"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
	generated "github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompiler/testdata/base/expected-output"
)

func TestCompile(t *testing.T) {
	docDataRoot := "testdata/base/data"
	expectedOutput := "testdata/base/expected-output/documentation.go"

	for _, tc := range []struct {
		name     string
		compiler func() DocCompiler
	}{
		{
			name: "from directory",
			compiler: func() DocCompiler {
				compiler := New()
				compiler.AddDataFromDir(docDataRoot)
				return compiler
			},
		},
		{
			name: "from generated",
			compiler: func() DocCompiler {
				docData := New()
				docData.AddDataFromGenerated(generated.Actions, generated.Config, generated.Extra, generated.Params, generated.Groups)
				return docData
			},
		},
	} {

		t.Run(tc.name, func(t *testing.T) {
			generatedDocs := tc.compiler().Compile()
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
				fmt.Println(diff.LineDiff(string(expectedDocs), string(generatedDocs)))
			}

		})
	}

}

func TestOverload(t *testing.T) {

	exampleData := func() *docData {
		return &docData{
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
				"action1": {},
				"action2": {},
			},
			ConfigFields: []string{"configField1", "configField2"},
			ConfigMap: map[string]common.DocEntry{
				"configField1": {},
				"configField2": {},
			},
			Extra: []string{"extra1", "extra2"},
			ExtraMap: map[string]common.DocEntry{
				"extra1": {},
				"extra2": {},
			},
		}
	}

	for _, tc := range []struct {
		name     string
		base     *docData
		new      *docData
		expected *docData
	}{
		{
			name: "simple overload",
			base: &docData{
				ParamMap: map[string][]string{
					"param1": {""},
				},
				Groups: []common.GroupsEntry{
					{
						Name:     "group1",
						Title:    "Group 1",
						Actions:  []string{"action1"},
						DocEntry: common.DocEntry{},
					},
				},
				Actions: []string{"action1"},
				ActionMap: map[string]common.DocEntry{
					"action1": {
						Description: "",
						Examples:    "",
					},
				},
				ConfigFields: []string{"configField1"},
				ConfigMap: map[string]common.DocEntry{
					"configField1": {
						Description: "",
						Examples:    "",
					},
				},
				Extra: []string{"extra1"},
				ExtraMap: map[string]common.DocEntry{
					"extra1": {
						Description: "",
						Examples:    "",
					},
				},
			},
			new: &docData{
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
					"action2": {},
				},
				ConfigFields: []string{"configField2"},
				ConfigMap: map[string]common.DocEntry{
					"configField2": {},
				},
				Extra: []string{"extra2"},
				ExtraMap: map[string]common.DocEntry{
					"extra2": {},
				},
			},
			expected: &docData{
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
					"action1": {},
					"action2": {},
				},
				ConfigFields: []string{"configField1", "configField2"},
				ConfigMap: map[string]common.DocEntry{
					"configField1": {},
					"configField2": {},
				},
				Extra: []string{"extra1", "extra2"},
				ExtraMap: map[string]common.DocEntry{
					"extra1": {},
					"extra2": {},
				},
			},
		},
		{
			name: "complex overload",
			base: &docData{
				ParamMap: map[string][]string{
					"param1": {""},
				},
				Groups: []common.GroupsEntry{
					{
						Name:    "group1",
						Title:   "Group 1",
						Actions: []string{"action1"},
						DocEntry: common.DocEntry{
							Description: "init",
							Examples:    "init",
						},
					},
				},
				Actions: []string{"action1"},
				ActionMap: map[string]common.DocEntry{
					"action1": {
						Description: "init",
						Examples:    "init",
					},
				},
				ConfigFields: []string{"configField1"},
				ConfigMap: map[string]common.DocEntry{
					"configField1": {
						Description: "",
						Examples:    "",
					},
				},
				Extra: []string{"extra1"},
				ExtraMap: map[string]common.DocEntry{
					"extra1": {
						Description: "",
						Examples:    "",
					},
				},
			},
			new: &docData{
				ParamMap: map[string][]string{
					"param2": {""},
				},
				Groups: []common.GroupsEntry{
					{
						Name:    "group1",
						Title:   "Group 1 updated",
						Actions: []string{"action2"},
						DocEntry: common.DocEntry{
							Description: "updated",
							Examples:    "updated",
						},
					},
				},
				Actions: []string{"action2"},
				ActionMap: map[string]common.DocEntry{
					"action1": {
						Description: "updated",
						Examples:    "updated",
					},
					"action2": {},
				},
				ConfigFields: []string{"configField2"},
				ConfigMap: map[string]common.DocEntry{
					"configField2": {},
				},
				Extra: []string{"extra2"},
				ExtraMap: map[string]common.DocEntry{
					"extra2": {},
				},
			},
			expected: &docData{
				ParamMap: map[string][]string{
					"param1": {""},
					"param2": {""},
				},
				Groups: []common.GroupsEntry{
					{
						Name:    "group1",
						Title:   "Group 1 updated",
						Actions: []string{"action1", "action2"},
						DocEntry: common.DocEntry{
							Description: "updated",
							Examples:    "updated",
						},
					},
				},
				Actions: []string{"action1", "action2"},
				ActionMap: map[string]common.DocEntry{
					"action1": {
						Description: "updated",
						Examples:    "updated",
					},
					"action2": {},
				},
				ConfigFields: []string{"configField1", "configField2"},
				ConfigMap: map[string]common.DocEntry{
					"configField1": {},
					"configField2": {},
				},
				Extra: []string{"extra1", "extra2"},
				ExtraMap: map[string]common.DocEntry{
					"extra1": {},
					"extra2": {},
				},
			},
		},
		{
			name:     "empty base",
			base:     newData(),
			new:      exampleData(),
			expected: exampleData(),
		},
		{
			name:     "empty new",
			base:     exampleData(),
			new:      newData(),
			expected: exampleData(),
		},
		{
			name:     "both empty",
			base:     newData(),
			new:      newData(),
			expected: newData(),
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

func TestOverloadGroups(t *testing.T) {
	for _, tc := range []struct {
		name     string
		base     []common.GroupsEntry
		new      []common.GroupsEntry
		expected []common.GroupsEntry
	}{
		{
			name: "actions with common group",
			base: []common.GroupsEntry{
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
			new: []common.GroupsEntry{
				{
					Name:    "group1",
					Title:   "Group 1",
					Actions: []string{"action2"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
			},
			expected: []common.GroupsEntry{
				{
					Name:    "group1",
					Title:   "Group 1",
					Actions: []string{"action1", "action2"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
			},
		},
		{
			name: "actions with separate groups",
			base: []common.GroupsEntry{
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
			new: []common.GroupsEntry{
				{
					Name:    "group2",
					Title:   "Group 2",
					Actions: []string{"action2"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
			},
			expected: []common.GroupsEntry{
				{
					Name:    "group1",
					Title:   "Group 1",
					Actions: []string{"action1"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
				{
					Name:    "group2",
					Title:   "Group 2",
					Actions: []string{"action2"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
			},
		},
		{
			name: "mixed case",
			base: []common.GroupsEntry{
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
			new: []common.GroupsEntry{
				{
					Name:    "group1",
					Title:   "Group 1",
					Actions: []string{"action4"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
				{
					Name:    "group2",
					Title:   "Group 2",
					Actions: []string{"action2", "action3"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
			},
			expected: []common.GroupsEntry{
				{
					Name:    "group1",
					Title:   "Group 1",
					Actions: []string{"action1", "action4"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
				{
					Name:    "group2",
					Title:   "Group 2",
					Actions: []string{"action2", "action3"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
			},
		},
		{
			name: "empty base",
			base: []common.GroupsEntry{},
			new: []common.GroupsEntry{
				{
					Name:    "group1",
					Title:   "Group 1",
					Actions: []string{"action2"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
			},
			expected: []common.GroupsEntry{
				{
					Name:    "group1",
					Title:   "Group 1",
					Actions: []string{"action2"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
			},
		},
		{
			name: "empty new",
			base: []common.GroupsEntry{
				{
					Name:    "group1",
					Title:   "Group 1",
					Actions: []string{"action2"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
			},
			new: []common.GroupsEntry{},
			expected: []common.GroupsEntry{
				{
					Name:    "group1",
					Title:   "Group 1",
					Actions: []string{"action2"},
					DocEntry: common.DocEntry{
						Description: "",
						Examples:    "",
					},
				},
			},
		},
		{
			name:     "both empty",
			base:     []common.GroupsEntry{},
			new:      []common.GroupsEntry{},
			expected: []common.GroupsEntry{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mergedGroups := mergeGroups(tc.base, tc.new)
			if !reflect.DeepEqual(mergedGroups, tc.expected) {
				t.Log(objDiff(mergedGroups, tc.expected))
				t.Fatalf("group overload gave unexpected result")
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
