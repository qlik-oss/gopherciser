package scenario

import (
	"testing"
	"time"

	"github.com/qlik-oss/gopherciser/helpers"
)

func TestValidateUnPublishSheetSettings(t *testing.T) {
	t.Parallel()

	tt := []struct {
		input   UnPublishSheetSettings
		isValid bool
	}{
		{UnPublishSheetSettings{AllSheets, nil, helpers.TimeDuration(time.Nanosecond)}, true},
		{UnPublishSheetSettings{AllSheets, []string{"A", "B", "C"}, helpers.TimeDuration(time.Nanosecond)}, false},
		{UnPublishSheetSettings{AllSheets, []string{}, helpers.TimeDuration(time.Nanosecond)}, true},
		{UnPublishSheetSettings{SheetIDs, nil, helpers.TimeDuration(time.Nanosecond)}, false},
		{UnPublishSheetSettings{SheetIDs, []string{}, helpers.TimeDuration(time.Nanosecond)}, false},
		{UnPublishSheetSettings{SheetIDs, []string{"A", "B", "C"}, helpers.TimeDuration(time.Nanosecond)}, true},
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
