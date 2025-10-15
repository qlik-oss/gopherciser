package scenario

import (
	"testing"
	"time"
)

func TestValidateUnPublishSheetSettings(t *testing.T) {
	t.Parallel()

	tt := []struct {
		input   UnPublishSheetSettings
		isValid bool
	}{
		{UnPublishSheetSettings{AllSheets, nil, time.Nanosecond}, true},
		{UnPublishSheetSettings{AllSheets, []string{"A", "B", "C"}, time.Nanosecond}, false},
		{UnPublishSheetSettings{AllSheets, []string{}, time.Nanosecond}, true},
		{UnPublishSheetSettings{SheetIDs, nil, time.Nanosecond}, false},
		{UnPublishSheetSettings{SheetIDs, []string{}, time.Nanosecond}, false},
		{UnPublishSheetSettings{SheetIDs, []string{"A", "B", "C"}, time.Nanosecond}, true},
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
