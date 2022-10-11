package session_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/goccy/go-json"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/users"
)

func Test_AppSelection_RoundRobin(t *testing.T) {
	js := []byte(`{
		"appmode" : "roundguidfromlist",
		"app" : "my app {{.Session}}",
		"list" : ["3869923b-6ccc-4a2d-9c1e-75a509269efe", "b2649ab9-877c-4dbf-adbf-508cb33e7f22"]
	}`)

	var appSelection session.AppSelection
	if err := json.Unmarshal(js, &appSelection); err != nil {
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

func Test_AppSelection_RoundRobin2(t *testing.T) {
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

	appSelection, err := session.NewAppSelection(session.AppModeRoundGUIDFromList, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i < 101; i++ {
		appSelection.AppList = append(appSelection.AppList, strconv.Itoa(i))
	}

	for j := 0; j < 2; j++ {
		for i := 1; i < 101; i++ {
			entry, err := appSelection.Select(sessionState)
			if err != nil {
				t.Fatal(err)
			}
			if entry.ID != strconv.Itoa(i) {
				t.Errorf("expected<%d> got<%s>", i, entry.ID)
			}
		}
	}
}

func Test_AppSelection_RoundRobin3(t *testing.T) {
	sessionState := &session.State{
		User: &users.User{
			UserName:  "mytestuser_1",
			Directory: "mydirectory",
		},
		LogEntry: &logger.LogEntry{
			Session: &logger.SessionEntry{},
			Action:  &logger.ActionEntry{},
		},
		ArtifactMap: session.NewArtifactMap(),
		Counters:    &statistics.ExecutionCounters{},
	}

	appSelection, err := session.NewAppSelection(session.AppModeRound, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i < 101; i++ {
		id := fmt.Sprintf("%03d", i)
		sessionState.ArtifactMap.Append(session.ResourceTypeApp, &session.ArtifactEntry{
			Name:         id,
			ID:           id,
			ItemID:       id,
			ResourceType: session.ResourceTypeApp,
			Data:         nil,
		})
	}

	for j := 0; j < 2; j++ {
		for i := 1; i < 101; i++ {
			entry, err := appSelection.Select(sessionState)
			if err != nil {
				t.Fatal(err)
			}
			if entry.ID != fmt.Sprintf("%03d", i) {
				t.Errorf("expected<%d> got<%s>", i, entry.ID)
			}
		}
	}
}
