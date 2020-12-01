package session_test

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/users"
)

var jsonit = jsoniter.ConfigCompatibleWithStandardLibrary

func Test_AppSelection_RoundRobin(t *testing.T) {
	js := []byte(`{
		"appmode" : "roundguidfromlist",
		"app" : "my app {{.Session}}",
		"list" : ["3869923b-6ccc-4a2d-9c1e-75a509269efe", "b2649ab9-877c-4dbf-adbf-508cb33e7f22"]
	}`)

	var appSelection session.AppSelection
	if err := jsonit.Unmarshal(js, &appSelection); err != nil {
		t.Fatal(err)
	}

	appSelection2 := appSelection

	sessionState := &session.State{
		User: &users.User{
			UserName:  "mytestuser_1",
			Directory: "mydirectory",
		},
		LogEntry: &logger.LogEntry{
			Session: &logger.SessionEntry{},
			Action:  &logger.ActionEntry{},
		},
	}
	entry, err := appSelection.Select(sessionState)
	if err != nil {
		t.Fatal(err)
	}

	verifyRoundPos(t, appSelection, entry.ID, 0)
	entry, err = appSelection2.Select(sessionState)
	if err != nil {
		t.Fatal(err)
	}
	verifyRoundPos(t, appSelection2, entry.ID, 1)

	entry, err = appSelection.Select(sessionState)
	if err != nil {
		t.Fatal(err)
	}
	verifyRoundPos(t, appSelection, entry.ID, 0)
}

func verifyRoundPos(t *testing.T, appSelection session.AppSelection, result string, pos int) {
	t.Helper()
	if result != appSelection.AppList[pos] {
		t.Errorf("app selection mode<%s> failed expected<%s> got<%s>", appSelection.AppMode, appSelection.AppList[pos], result)
	}
}
