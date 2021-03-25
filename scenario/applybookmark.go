package scenario

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//ApplyBookmarkSettings apply bookmark settings
	ApplyBookmarkSettings struct {
		BookMarkSettings
		SelectionsOnly bool `json:"selectionsonly" displayname:"Apply selections only" doc-key:"applybookmark.selectionsonly"`
	}

	bmSearchTerm int
)

const (
	bmSearchTitle bmSearchTerm = iota
	bmSearchId
)

// Validate ApplyBookmarkSettings action (Implements ActionSettings interface)
func (settings ApplyBookmarkSettings) Validate() ([]string, error) {
	if (settings.Title.String() == "") == (settings.ID == "") {
		return nil, errors.New("specify exactly one of the following - bookmark title or bookmark id")
	}
	return nil, nil
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

	// Send GetApplayout request
	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.CurrentApp.Doc.GetAppLayout(ctx)
		return errors.WithStack(err)
	}, actionState, false, "GetAppLayout request failed")

	// Get and add Bookmark to session objects
	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := settings.getBookmarkObject(sessionState, actionState, uplink)
		return errors.WithStack(err)
	}, actionState, true, "failed to get bookmark")

	if sessionState.Wait(actionState) {
		return // An error occurred
	}

	if sheetID != "" && !settings.SelectionsOnly {
		sessionState.LogEntry.LogDebug(fmt.Sprint("ApplyBookmark: Change sheet to ", sheetID))
		(&ChangeSheetSettings{
			ID: sheetID,
		}).Execute(sessionState, actionState, connectionSettings, label, reset)
	}
	actionState.Details = fmt.Sprintf("%v;%s;%v", sheetID != "", sheetID, settings.SelectionsOnly) // log details in results as {Has sheet};{Sheet ID};{Apply Selections Only}

	sessionState.Wait(actionState)
}

// AffectsAppObjectsAction implements AffectsAppObjectsAction interface
func (settings ApplyBookmarkSettings) AffectsAppObjectsAction(structure appstructure.AppStructure) ([]*appstructure.AppStructurePopulatedObjects, []string, bool) {
	id := settings.BookMarkSettings.ID
	if id == "" { // No ID, specified, search by title
		title := settings.BookMarkSettings.Title.String()
		for _, v := range structure.Bookmarks {
			if v.Title == title {
				id = v.ID
				break
			}
		}
	}
	if id == "" { // Bookmark not found
		return nil, nil, false
	}
	bookmark := structure.Bookmarks[id]
	if bookmark.SheetId == nil { // Bookmark not associated with a sheet
		return nil, nil, false
	}

	// Found sheet id, now find sheet objects
	selectables, err := structure.GetSelectables(*bookmark.SheetId)
	if err != nil {
		return nil, nil, false
	}
	newObjs := appstructure.AppStructurePopulatedObjects{
		Parent:    settings.ID,
		Objects:   make([]appstructure.AppStructureObject, 0),
		Bookmarks: nil,
	}
	newObjs.Objects = append(newObjs.Objects, selectables...)
	return []*appstructure.AppStructurePopulatedObjects{&newObjs}, nil, true
}
