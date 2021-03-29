package helpers

import "testing"

func TestToValidWindowsFileName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			`06ee5d4d-979a-4808-8eb9-7a76402066c5:0:Chart_image.png`,
			`06ee5d4d-979a-4808-8eb9-7a76402066c5_0_Chart_image.png`,
		},
		{
			`\\\`,
			`___`,
		},
		{
			`*<>"/\:|?`,
			`_________`,
		},
		{
			`A---B`,
			`A---B`,
		},
	}

	for _, tc := range testCases {
		if output := ToValidWindowsFileName(tc.input); output != tc.expected {
			t.Errorf("input<%s>\ngot<%s>\nexpected<%s>\n", tc.input, output, tc.expected)
		}

	}

}
