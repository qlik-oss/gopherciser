package scenario

import (
	"github.com/goccy/go-json"
	"reflect"
	"strings"
	"testing"
)

func TestSubstituteVariable(t *testing.T) {
	cases := []struct {
		src      string
		repl     string
		expected string
	}{
		{"$value", "123", "123"},
		{"hello $value", "bob", "hello bob"},
		{"hello $value?", "alice", "hello alice?"},
		{"hello $value?", "$al$ice", "hello $al$ice?"},
		{"show $value!", "123", "show 123!"},
		{"show $value!", "measures", "show measures!"},
		{"show $value!", "x.y", "show x.y!"},
		{"x$value!", "y", "xy!"},
		{"x$value!", "y", "xy!"},
	}
	for _, c := range cases {
		res := substituteVariable(c.src, c.repl)
		if res != c.expected {
			t.Errorf("expected<%s>, got<%s>", c.expected, res)
		}

	}
}

func TestUnmarshalAndMarhsalWeightedQuery(t *testing.T) {
	cases := []struct {
		input              string
		stdForm            string
		wqs                []WeightedQuery
		shallFailUnmarshal bool
	}{
		{
			input:   `["query"]`,
			stdForm: `[{"weight":1,"query":"query"}]`,
			wqs:     []WeightedQuery{{WeightedQueryCore{1, "query"}}},
		},
		{
			input:   `[{"weight":2,"query":"weightedquery"}]`,
			stdForm: `[{"weight":2,"query":"weightedquery"}]`,
			wqs:     []WeightedQuery{{WeightedQueryCore{2, "weightedquery"}}},
		},
		{
			input:              `[["weightedquery",2]]`,
			shallFailUnmarshal: true,
		},
	}
	for _, c := range cases {
		// Unmarshal
		var wqs []WeightedQuery
		err := json.Unmarshal([]byte(c.input), &wqs)
		if err != nil {
			if !c.shallFailUnmarshal {
				t.Errorf("%s: %s", err, c.input)
			}
			return
		}
		if c.shallFailUnmarshal {
			t.Errorf("`%s`should have failed unmarshal", c.input)
			return
		}

		if !reflect.DeepEqual(c.wqs, wqs) {
			t.Errorf("expected %+v, got %+v", c.wqs, wqs)

		}

		// Marshal back to stdForm
		bytes, err := json.Marshal(wqs)
		if err != nil {
			t.Error(err)
		}
		if string(bytes) != c.stdForm {
			t.Errorf(`expected %s, got %s`, c.stdForm, string(bytes))
		}
	}
}

func TestParseWeightedQueries(t *testing.T) {
	input := strings.NewReader(`
		1; query foo

		query bar
		
		3; query baz`,
	)
	wqs, err := ParseWeightedQueries(input)
	if err != nil {
		t.Error(err)
	}
	expected := []WeightedQuery{
		{WeightedQueryCore{1, "query foo"}},
		{WeightedQueryCore{1, "query bar"}},
		{WeightedQueryCore{3, "query baz"}},
	}
	if !reflect.DeepEqual(wqs, expected) {
		t.Errorf("expected %+v, got %+v", expected, wqs)
	}
}
