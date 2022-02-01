package scenario

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestParseSearchTerms(t *testing.T) {
	cases := []struct {
		input  string
		output []string
	}{
		{
			`"a b c`,
			[]string{
				`a b c`,
			},
		},
		{
			`"a b c" "e f g`,
			[]string{
				`a b c`,
				`e f g`,
			},
		},
		{
			`"a b c""e f g`,
			[]string{
				`a b ce f g`,
			},
		},
		{
			`"a b c ""e f g`,
			[]string{
				`a b c e f g`,
			},
		},
		{
			`"a b c"asd`,
			[]string{
				`a b casd`,
			},
		},
		{
			`"a b c"asd"`,
			[]string{
				`a b casd`,
			},
		},
		{
			`"a b c" asd "aa`,
			[]string{
				`a b c`,
				`asd`,
				`aa`,
			},
		},
		{
			`"a b    c    "as     d`,
			[]string{
				`a b c as`,
				`d`,
			},
		},
		{
			`""""""""a """"""""`,
			[]string{
				`a`,
			},
		},
		{
			`\\ \"\"`,
			[]string{
				`\`,
				`""`,
			},
		},
		{
			`\"abc\"`,
			[]string{
				`"abc"`,
			},
		},
		{
			`1 2 3 \4 \[] ten "a sentense this is"`,
			[]string{
				`1`,
				`2`,
				`3`,
				`4`,
				`[]`,
				`ten`,
				`a sentense this is`,
			},
		},
	}
	for idx, c := range cases {
		t.Run(fmt.Sprint(idx), func(t *testing.T) {
			res := parseSearchTerms(c.input)
			if !reflect.DeepEqual(res, c.output) {
				expected, err := json.Marshal(c.output)
				if err != nil {
					t.Fatal(err)
				}
				got, err := json.Marshal(res)
				if err != nil {
					t.Fatal(err)
				}
				t.Errorf("expected %s, got %s", expected, got)
			}

		})

	}
}
