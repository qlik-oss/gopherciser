package scenario

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//ApplyBookmarkSettings apply bookmark settings
	ApplyBookmarkSettings struct {
		Title string `json:"title" displayname:"Bookmark title" doc-key:"applybookmark.title"`
		Id    string `json:"id" displayname:"Bookmark ID" doc-key:"applybookmark.id"`
	}

	bmSearchTerm int
)

const (
	bmSearchTitle bmSearchTerm = iota
	bmSearchId
)

// Validate ApplyBookmarkSettings action (Implements ActionSettings interface)
func (settings ApplyBookmarkSettings) Validate() error {
	if (settings.Title == "") == (settings.Id == "") {
		return errors.New("specify exactly one of the following - bookmark title or bookmark id")
	}
	return nil
}

// Execute ApplyBookmarkSettings action (Implements ActionSettings interface)
func (settings ApplyBookmarkSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("not connected to a Sense environment"))
		return
	}

	app := sessionState.Connection.Sense().CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("not connected to a Sense app"))
		return
	}

	uplink := sessionState.Connection.Sense()

	bl, err := uplink.CurrentApp.GetBookmarkList(sessionState, actionState)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to GetBookmarkList"))
		return
	}

	var input string
	var term bmSearchTerm

	if settings.Id == "" {
		input = settings.Title
		term = bmSearchTitle
	} else {
		input = sessionState.IDMap.Get(settings.Id)
		term = bmSearchId
	}

	id, sheetID, err := getBookmarkData(bl, input, term)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	if id == "" { // No ID and named bookmark not found
		actionState.AddErrors(errors.New("could not find specified bookmark"))
		return
	}

	err = sessionState.SendRequest(actionState, func(ctx context.Context) error {
		success, err := uplink.CurrentApp.Doc.ApplyBookmark(ctx, id)
		if err != nil {
			return err
		}
		if success {
			return nil
		}
		return errors.Errorf("applying bookmark<%s> unsuccessful.", id)
	})
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to apply bookmark"))
		return
	}

	sessionState.Wait(actionState)

	sessionState.LogEntry.LogDebug(fmt.Sprint("ApplyBookmark: Change sheet to ", sheetID))

	if sheetID == "" {
		sessionState.LogEntry.Log(logger.WarningLevel, "no sheet id found in bookmark, sheet not changed")
	} else {
		(&ChangeSheetSettings{
			ID: sheetID,
		}).Execute(sessionState, actionState, connectionSettings, label, reset)

	}

	sessionState.Wait(actionState)
}

func getBookmarkData(bl *senseobjects.BookmarkList, input string, term bmSearchTerm) (string /*id*/, string /*sheetId*/, error) {
	for _, bookmark := range bl.GetBookmarks() {
		switch term {
		case bmSearchTitle:
			if input == bookmark.Data.Title {
				return bookmark.Info.Id, bookmark.Data.SheetId, nil
			}
		case bmSearchId:
			if input == bookmark.Info.Id {
				return bookmark.Info.Id, bookmark.Data.SheetId, nil
			}
		}
	}

	return "", "", errors.New("bookmark not found")
}
