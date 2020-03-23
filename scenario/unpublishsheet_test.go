package scenario

import "testing"

func TestValidateUnPublishSheetSettings(t *testing.T) {
	t.Parallel()

	tt := []struct {
		input   UnPublishSheetSettings
		isValid bool
	}{
		{UnPublishSheetSettings{AllSheets, nil}, true},
		{UnPublishSheetSettings{AllSheets, []string{"A", "B", "C"}}, false},
		{UnPublishSheetSettings{AllSheets, []string{}}, true},
		{UnPublishSheetSettings{SheetIDs, nil}, false},
		{UnPublishSheetSettings{SheetIDs, []string{}}, false},
		{UnPublishSheetSettings{SheetIDs, []string{"A", "B", "C"}}, true},
	}

	for _, tc := range tt {
		err := tc.input.Validate()
		if tc.isValid && err != nil {
			t.Errorf("Settings <%v> should be valid, but it's not <%v>", tc.input, err)
		} else if !tc.isValid && err == nil {
			t.Errorf("Settings <%v> should not be valid, but it do", tc.input)
		}
	}
}
