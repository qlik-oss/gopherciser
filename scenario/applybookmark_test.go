package scenario_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/session"
)

func TestUnmarshalApplyBookmark(t *testing.T) {
	title, err := session.NewSyncedTemplate("hasTitle")
	if err != nil {
		t.Fatal(err)
	}
	expectedBM := scenario.ApplyBookmarkSettings{
		BookMarkSettings: scenario.BookMarkSettings{
			Title: *title,
			ID:    "hasid",
		},
		SelectionsOnly: true,
	}

	raw := []byte(`{
		"title": "` + title.String() + `",
		"id": "` + expectedBM.ID + `",
		"selectionsonly" : ` + fmt.Sprintf("%v", expectedBM.SelectionsOnly) + `
}`)
	var bm scenario.ApplyBookmarkSettings
	if err := json.Unmarshal(raw, &bm); err != nil {
		t.Fatal(err)
	}

	if expectedBM.Title.String() != bm.Title.String() {
		t.Errorf("unexpected title<%s>, expected<%s>", bm.Title.String(), expectedBM.Title.String())
	}

	if expectedBM.ID != bm.ID {
		t.Errorf("unexpected ID<%s>, expected<%s>", bm.ID, expectedBM.ID)
	}

	if expectedBM.SelectionsOnly != bm.SelectionsOnly {
		t.Errorf("unexpected selectionsonly<%v>, expected<%v>", bm.SelectionsOnly, expectedBM.SelectionsOnly)
	}
}
