package senseobjdef

import (
	"github.com/goccy/go-json"
	"github.com/qlik-oss/gopherciser/helpers"
	"strconv"
	"testing"

	"github.com/pkg/errors"
)

func TestConstraintValue(t *testing.T) {
	jsonData := `{
		"some" : {
			"data" : [
				{
					"toEval" : 100
				},{
					"toEval" : 4.6
				},{
					"toEval" : -23.154
				},{
					"toEval" : true
				},{
					"toEval" : "sometext"
				}
			]
		}
	}`

	type testData struct {
		constraint     *Constraint
		expectedResult bool
		expectedErr    error
	}

	constraints := []testData{
		//float64 tests
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[0]/toEval"),
				Value: ConstraintValue("=100"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[0]/toEval"),
				Value: ConstraintValue("=50"),
			}, false, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[0]/toEval"),
				Value: ConstraintValue("<500"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[0]/toEval"),
				Value: ConstraintValue("<100"),
			}, false, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[0]/toEval"),
				Value: ConstraintValue("<50"),
			}, false, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[0]/toEval"),
				Value: ConstraintValue(">50"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[0]/toEval"),
				Value: ConstraintValue(">500"),
			}, false, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[0]/toEval"),
				Value: ConstraintValue(">100"),
			}, false, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[0]/toEval"),
				Value: ConstraintValue("!50"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[0]/toEval"),
				Value: ConstraintValue("!100"),
			}, false, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[0]/toEval"),
				Value: ConstraintValue("*100"),
			}, false, MalformedConstraintValueError("*100"),
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[1]/toEval"),
				Value: ConstraintValue("=4.6"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[1]/toEval"),
				Value: ConstraintValue("!4.6"),
			}, false, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[1]/toEval"),
				Value: ConstraintValue("<5"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[1]/toEval"),
				Value: ConstraintValue(">4"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[2]/toEval"),
				Value: ConstraintValue("=-23.154"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[2]/toEval"),
				Value: ConstraintValue("<-23.153"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[2]/toEval"),
				Value: ConstraintValue(">3.154"),
			}, false, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[2]/toEval"),
				Value: ConstraintValue("!-23.154"),
			}, false, nil,
		},
		//bool tests
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[3]/toEval"),
				Value: ConstraintValue("=true"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[3]/toEval"),
				Value: ConstraintValue("=1"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[3]/toEval"),
				Value: ConstraintValue("=false"),
			}, false, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[3]/toEval"),
				Value: ConstraintValue("!false"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[3]/toEval"),
				Value: ConstraintValue("!0"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[3]/toEval"),
				Value: ConstraintValue("!true"),
			}, false, nil,
		},
		//string tests
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[4]/toEval"),
				Value: ConstraintValue("=sometext"),
			}, true, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[4]/toEval"),
				Value: ConstraintValue("=notmytext"),
			}, false, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[4]/toEval"),
				Value: ConstraintValue("!sometext"),
			}, false, nil,
		},
		{
			&Constraint{
				Path:  helpers.DataPath("/some/data/[4]/toEval"),
				Value: ConstraintValue("!mytext"),
			}, true, nil,
		},
	}

	for i, v := range constraints {
		test := v
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			result, err := test.constraint.Evaluate(json.RawMessage(jsonData))
			if errors.Cause(err) != test.expectedErr {
				t.Error(err)
			}

			if result != test.expectedResult {
				t.Errorf("result<%v> not expected<%v> constraint<%s>",
					result, test.expectedResult, string(test.constraint.Value))
			}
		})
	}
}
