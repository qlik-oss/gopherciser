package scenario

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	BookMarkSettings struct {
		Title session.SyncedTemplate `json:"title" displayname:"Bookmark title" doc-key:"applybookmark.title"`
		ID    string                 `json:"id" displayname:"Bookmark ID" doc-key:"applybookmark.id"`
	}

	//ApplyBookmarkSettings apply bookmark settings
	ApplyBookmarkSettings struct {
		BookMarkSettings
	}

	bmSearchTerm int
)

const (
	bmSearchTitle bmSearchTerm = iota
	bmSearchId
)

// Validate ApplyBookmarkSettings action (Implements ActionSettings interface)
func (settings ApplyBookmarkSettings) Validate() error {
	if (settings.Title.String() == "") == (settings.ID == "") {
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

	id, sheetID, err := settings.getBookmark(sessionState, actionState, uplink)
	if err != nil {
		actionState.AddErrors(err)
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

	if sheetID != "" {
		sessionState.LogEntry.LogDebug(fmt.Sprint("ApplyBookmark: Change sheet to ", sheetID))
		(&ChangeSheetSettings{
			ID: sheetID,
		}).Execute(sessionState, actionState, connectionSettings, label, reset)
	}
	actionState.Details = fmt.Sprintf("%v;%s", sheetID != "", sheetID) // log details in results as {Has sheet};{Sheet ID}

	sessionState.Wait(actionState)
}

// getBookmark defined by BookMarkSettings, returns bookmark ID, sheet ID
func (settings BookMarkSettings) getBookmark(sessionState *session.State, actionState *action.State, uplink *enigmahandlers.SenseUplink) (string, string, error) {
	bl, err := uplink.CurrentApp.GetBookmarkList(sessionState, actionState)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to GetBookmarkList")
	}

	term, input, err := settings.getSearchTerm(sessionState)
	if err != nil {
		return "", "", err
	}

	id, sheetID, err := getBookmarkData(bl, input, term)
	if err != nil {
		return "", "", err
	}

	if id == "" { // No ID and named bookmark not found
		return id, sheetID, errors.New("could not find specified bookmark")
	}

	return id, sheetID, nil
}

func (settings BookMarkSettings) getSearchTerm(sessionState *session.State) (bmSearchTerm, string, error) {
	var input string
	var term bmSearchTerm
	var err error

	if settings.ID == "" {
		input, err = sessionState.ReplaceSessionVariables(&settings.Title)
		term = bmSearchTitle
	} else {
		input = sessionState.IDMap.Get(settings.ID)
		term = bmSearchId
	}
	return term, input, err
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
