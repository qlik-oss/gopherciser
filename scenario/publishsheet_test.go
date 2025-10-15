package scenario

import (
	"testing"
	"time"
)

func TestUnmarshalPublishSheetMode(t *testing.T) {
	t.Parallel()

	tt := []struct {
		input  string
		isErr  bool
		output PublishSheetMode
	}{
		{`"allsheets"`, false, AllSheets},
		{`"sheetids"`, false, SheetIDs},
		{`"wabalabadubdub"`, true, 0},
		{`""`, true, 0},
	}

	for _, tc := range tt {
		var val PublishSheetMode
		err := (&val).UnmarshalJSON([]byte(tc.input))
		if err == nil && tc.isErr {
			t.Errorf("Expected to get an error but got <%v>", err)
		} else if err != nil && !tc.isErr {
			t.Errorf("No error expected but got <%v>", err)
		}
		if val != tc.output {
			t.Errorf("Expected value <%v>, got <%v>", tc.output, val)
		}
	}
}

func TestMarshalPublishSheetMode(t *testing.T) {
	t.Parallel()

	tt := []struct {
		input  PublishSheetMode
		isErr  bool
		output string
	}{
		{AllSheets, false, `"allsheets"`},
		{SheetIDs, false, `"sheetids"`},
		{999, true, ""},
	}

	for _, tc := range tt {
		b, err := tc.input.MarshalJSON()
		if err == nil && tc.isErr {
			t.Errorf("Expected to get an error but got <%v>", err)
		} else if err != nil && !tc.isErr {
			t.Errorf("No error expected but got <%v>", err)
		}
		s := string(b)
		if s != tc.output {
			t.Errorf("Expected marshalled value <%v>, got <%v>", tc.output, s)
		}
	}
}

func TestValidatePublishSheetSettings(t *testing.T) {
	t.Parallel()

	tt := []struct {
		input   PublishSheetSettings
		isValid bool
	}{
		{PublishSheetSettings{AllSheets, nil, false, time.Nanosecond}, true},
		{PublishSheetSettings{AllSheets, []string{"A", "B", "C"}, false, time.Nanosecond}, false},
		{PublishSheetSettings{AllSheets, []string{}, true, time.Nanosecond}, true},
		{PublishSheetSettings{SheetIDs, nil, false, time.Nanosecond}, false},
		{PublishSheetSettings{SheetIDs, []string{}, true, time.Nanosecond}, false},
		{PublishSheetSettings{SheetIDs, []string{"A", "B", "C"}, false, time.Nanosecond}, true},
	}

	for _, tc := range tt {
		_, err := tc.input.Validate()
		if tc.isValid && err != nil {
			t.Errorf("Settings <%v> should be valid, but it's not <%v>", tc.input, err)
		} else if !tc.isValid && err == nil {
			t.Errorf("Settings <%v> should not be valid, but it do", tc.input)
		}
	}
}
