package senseobjdef

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

var (
	testObjDefs = ObjectDefs{
		// listbox override
		"listbox": {
			DataDef: DataDef{
				Type: DataDefListObject,
				Path: DataPath("/qListObject"),
			},
			Data: []Data{
				{
					Constraints: nil,
					Requests: []GetDataRequests{
						{
							Type:   DataTypeListObject,
							Path:   "/qListObjectDef",
							Height: 10000,
						},
					},
				},
			},
			Select: &Select{
				Type: SelectTypeListObjectValues,
				Path: "/qListObjectDef",
			},
		},
		"customscatterplot": {
			DataDef: DataDef{
				Type: DataDefHyperCube,
				Path: DataPath("/qHyperCube"),
			},
			Data: []Data{
				{
					Constraints: []*Constraint{&Constraint{
						Path:  "/qHyperCube/qSize/qcy",
						Value: ">1000",
					}},
					Requests: []GetDataRequests{
						{
							Type:   DataTypeHyperCubeBinnedData,
							Path:   "/qHyperCubeDef",
							Height: 10000,
						},
					},
				},
				{
					Constraints: nil,
					Requests: []GetDataRequests{
						{
							Type:   DataTypeHyperCubeData,
							Path:   "/qHyperCubeDef",
							Height: 1000,
						},
					},
				},
			},
			Select: &Select{
				Type: SelectTypeHypercubeValues,
				Path: "/qHyperCubeDef",
			},
		},
	}
)

func TestConstraints(t *testing.T) {
	t.Parallel()

	def := &ObjectDef{
		Data: []Data{
			{
				Constraints: []*Constraint{&Constraint{
					Path:  DataPath("/qHyperCube/qSize/qcy"),
					Value: ConstraintValue(">1000"),
				}},
				Requests: []GetDataRequests{
					{
						Type: DataTypeHyperCubeBinnedData,
					},
				},
			}, {
				Constraints: nil,
				Requests: []GetDataRequests{
					{
						Type: DataTypeHyperCubeData,
					},
				},
			},
		},
	}

	data1 := `{
		"qHyperCube" : {
			"qSize": {
				"qcy": 2000
			}
		}
	}`

	requests, err := def.Evaluate(json.RawMessage(data1))
	t.Logf("%+v", requests)
	if err != nil {
		t.Fatal(err)
	}
	if requests == nil || len(requests) != 1 {
		t.Fatal("unexpected requests length for data1")
	}
	if requests[0].Type != DataTypeHyperCubeBinnedData {
		t.Errorf("unexpected request type<%d> evaulated, expected<%d>", requests[0].Type, DataTypeHyperCubeBinnedData)
	}

	data2 := `{
		"qHyperCube" : {
			"qSize": {
				"qcy": 500
			}
		}
	}`
	requests, err = def.Evaluate(json.RawMessage(data2))
	t.Logf("%+v", requests)
	if err != nil {
		t.Fatal(err)
	}
	if requests == nil || len(requests) != 1 {
		t.Fatal("unexpected requests length for data2")
	}
	if requests[0].Type != DataTypeHyperCubeData {
		t.Errorf("unexpected request type<%d> evaulated, expected<%d>", requests[0].Type, DataTypeHyperCubeData)
	}

}

func TestConfig(t *testing.T) {
	t.Parallel()

	raw := `{
		"listbox": {
			"datadef": {
				"type": "listobject",
				"path": "/qListObject"
			},
			"data": [
				{
					"requests": [
						{
							"type": "ListObjectData",
							"path": "/qListObjectDef",
							"height": 10000
						}
					]
				}
			],
			"select": {
				"type": "ListObjectValues",
				"path": "/qListObjectDef"
			}
		},
		"customscatterplot": {
			"datadef": {
				"type": "hypercube",
				"path": "/qHyperCube"
			},
			"data": [
				{
					"constraints": [{
						"path": "/qHyperCube/qSize/qcy",
						"value": ">1000"
					}],
					"requests": [
						{
							"type": "HyperCubeBinnedData",
							"path": "/qHyperCubeDef",
							"height": 10000
						}
					]
				},
				{
					"requests": [
						{
							"type": "HyperCubeData",
							"path": "/qHyperCubeDef",
							"height": 1000
						}
					]
				}
			],
			"select": {
				"type": "HyperCubeValues",
				"path": "/qHyperCubeDef"
			}
		}	
	}`

	var defs ObjectDefs
	if err := jsonit.Unmarshal([]byte(raw), &defs); err != nil {
		t.Fatal(err)
	}

	validateConfigStruct(t, defs, testObjDefs)

	jString, err := jsonit.Marshal(defs)
	if err != nil {
		t.Fatal(err)
	}
	objDefs := string(jString)
	t.Log("marshaled json:", objDefs)

	expectedJSON := `{"customscatterplot":{"datadef":{"type":"hypercube","path":"/qHyperCube"},"data":[{"constraints":[{"path":"/qHyperCube/qSize/qcy","value":"\u003e1000"}],"requests":[{"type":"hypercubebinneddata","path":"/qHyperCubeDef","height":10000}]},{"requests":[{"type":"hypercubedata","path":"/qHyperCubeDef","height":1000}]}],"select":{"type":"hypercubevalues","path":"/qHyperCubeDef"}},"listbox":{"datadef":{"type":"listobject","path":"/qListObject"},"data":[{"requests":[{"type":"listobjectdata","path":"/qListObjectDef","height":10000}]}],"select":{"type":"listobjectvalues","path":"/qListObjectDef"}}}`
	if objDefs != expectedJSON {
		t.Log("expected json:", expectedJSON)
		t.Error("unexpected marshaled json")
	}
}

func TestOverideFromFile(t *testing.T) {
	// test unmarshaling from file
	odf, err := ioutil.TempFile("", "objdef")
	if err != nil {
		t.Fatal("failed to create temporary file", err)
	}
	defer func() {
		if err := odf.Close(); err != nil {
			t.Error("failed closing file", err)
		}
		// todo delete file
	}()

	overrides, err := jsonit.Marshal(testObjDefs)
	if err != nil {
		t.Fatal("error marshaling testObjDefs", err)
	}

	if _, err := odf.Write(overrides); err != nil {
		t.Fatal("error writing object definitions to temporary file")
	}

	defs := make(ObjectDefs, len(DefaultObjectDefs))
	for k, v := range DefaultObjectDefs {
		defs[k] = v
	}

	err = defs.OverrideFromFile(odf.Name())
	if err != nil {
		t.Fatal("error unmarshaling object definitions from temp file", err)
	}
	t.Log(defs)

	defaultListboxHeight := DefaultObjectDefs["listbox"].Data[0].Requests[0].Height // todo nil checks etc
	if defaultListboxHeight != DefaultDataHeight {
		t.Errorf("incorrect default listbox height<%d> expected<%d>", defaultListboxHeight, DefaultDataHeight)
	}

	overridenListboxHeight := defs["listbox"].Data[0].Requests[0].Height // todo nil checks etc
	expectedHeight := 10000
	if overridenListboxHeight != expectedHeight {
		t.Errorf("incorrect overriden listbox height<%d> expected<%d>", overridenListboxHeight, expectedHeight)
	}

	// validateConfigStruct(t, defs, DefaultObjectDefs)
}

func TestDefault(t *testing.T) {
	if len(DefaultObjectDefs) < 1 {
		t.Fatal("Default objects list is empty")
	}

	// Validate "simple" object
	treemap := DefaultObjectDefs["treemap"]
	if treemap == nil {
		t.Error("no treemap definition found")
	} else {
		dataConstraintsCount := len(treemap.Data)
		expectedDataConstraintsCount := 1
		if dataConstraintsCount != expectedDataConstraintsCount {
			t.Errorf("incorrect amount of data constraints for treemap<%d> expected<%d>", dataConstraintsCount, expectedDataConstraintsCount)
		} else {
			dataConstraint := treemap.Data[0]
			if dataConstraint.Constraints != nil {
				t.Error("tree map contains unexpected data request constraint:", dataConstraint.Constraints)
			}

			dataRequestsCount := len(dataConstraint.Requests)
			expectedDataRequestCount := 1
			if dataRequestsCount != expectedDataRequestCount {
				t.Errorf("Treemap default data constraint contains unexpected request count<%d> expected<%d>", dataRequestsCount, expectedDataRequestCount)
			} else {
				dataRequest := dataConstraint.Requests[0]
				if dataRequest.Path != "" {
					t.Error("treemap data request unexpectedly contains a path:", dataRequest.Path)
				}
				if dataRequest.Type != DataTypeLayout {
					t.Error("treemap data request not of type DataTypeLayout, type:", dataRequest.Type)
				}
				if dataRequest.Height != 0 && dataRequest.Height != DefaultDataHeight {
					t.Errorf("treemap data request height<%d> not set to 0 or default height<%d>", dataRequest.Height, DefaultDataHeight)
				}
			}
		}
	}

	// todo add multi constraint object test

	// todo add multi data request object test
}

func validateConfigStruct(t *testing.T, defs ObjectDefs, tests ObjectDefs) {
	t.Helper()

	if defs == nil {
		t.Fatal("empty object definition list")
	}

	if len(defs) != len(tests) {
		t.Fatalf("tests def<%+v> length not equal to object defs<%+v> length", tests, defs)
	}

	for k, v := range tests {
		object := k
		test := v
		t.Run(object, func(t *testing.T) {
			t.Parallel()

			def := defs[object]

			if err := def.Validate(); err != nil {
				t.Fatal(err)
			}

			if err := validateDataDef(object, def.DataDef, test.DataDef); err != nil {
				t.Error(err)
			}

			if err := validateData(object, def.Data, test.Data); err != nil {
				t.Error(err)
			}

			if def.Select == nil || test.Select == nil {
				if def.Select != test.Select {
					t.Errorf("select unexpectedly nil def<%p> test<%p>", def.Select, test.Select)
				}
			} else if err := validateSelect(object, *def.Select, *test.Select); err != nil {
				t.Error(err)
			}
		})
	}
}

func validateDataDef(object string, def DataDef, test DataDef) error {
	if def.Path != test.Path {
		return fmt.Errorf("object<%s> data def path<%s> not the expected<%s>",
			object, def.Path, test.Path)
	}

	if def.Type != test.Type {
		return fmt.Errorf("object<%s> data def type<%v> not the expected<%v>",
			object, def.Type, test.Type)
	}
	return nil
}

func validateData(object string, data []Data, test []Data) error {
	if len(data) != len(test) {
		return fmt.Errorf("object<%s> data length<%d> not the expected<%d>", object, len(data), len(test))
	}

	for i, v := range data {
		for j, c := range v.Constraints {
			if err := validateConstraint(object, c, test[i].Constraints[j]); err != nil {
				return err
			}
		}

		if err := validateRequests(object, v.Requests, test[i].Requests); err != nil {
			return err
		}
	}

	return nil
}

func validateConstraint(object string, constraint *Constraint, test *Constraint) error {
	if constraint == nil || test == nil {
		if constraint == test {
			return nil
		}
		return fmt.Errorf("object<%s> constraint<%+v> not expected<%+v>", object, constraint, test)
	}

	if string(constraint.Path) != string(test.Path) {
		return fmt.Errorf("object<%s> constraint path<%s> not expected<%s>",
			object, string(constraint.Path), string(test.Path))
	}

	if string(constraint.Value) != string(test.Value) {
		return fmt.Errorf("object<%s> constraint value<%s> not expected<%s>",
			object, string(constraint.Value), string(test.Value))
	}

	return nil
}

func validateRequests(object string, requests []GetDataRequests, tests []GetDataRequests) error {
	if requests == nil && tests == nil {
		return nil
	}

	if requests == nil || tests == nil || len(requests) != len(tests) {
		return fmt.Errorf("object<%s> requests<%+v> not expected<%+v>", object, requests, tests)
	}

	for i, v := range requests {
		if v.Type != tests[i].Type {
			return fmt.Errorf("object<%s> request type<%d> not the expected<%d>",
				object, v.Type, tests[i].Type)
		}

		if v.Path != tests[i].Path {
			return fmt.Errorf("object<%s> request path<%s> not the expected<%s>",
				object, v.Path, tests[i].Path)
		}

		if v.Height != tests[i].Height {
			return fmt.Errorf("object<%s> request height<%d> not the expected<%d>",
				object, v.Height, tests[i].Height)
		}
	}

	return nil
}

func validateSelect(object string, sel Select, test Select) error {
	if sel.Type != test.Type {
		return fmt.Errorf("object<%s> select type<%d> not expected<%d>",
			object, sel.Type, test.Type)
	}

	if sel.Path != test.Path {
		return fmt.Errorf("object<%s> select path<%s> not expected<%s>",
			object, sel.Path, test.Path)
	}

	return nil
}
