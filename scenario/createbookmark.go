package scenario

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/creation"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//CreateBookmarkSettings create bookmark settings
	CreateBookmarkSettings struct {
		BookMarkSettings
		Description     string `json:"description" displayname:"Bookmark description" doc-key:"createbookmark.description"`
		NoSheetLocation bool   `json:"nosheet" displayname:"Exclude sheet location" doc-key:"createbookmark.nosheet"`
		SaveLayout      bool   `json:"savelayout" displayname:"Save layout" doc-key:"createbookmark.savelayout"`
	}
)

// Validate CreateBookmarkSettings action (Implements ActionSettings interface)
func (settings CreateBookmarkSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute CreateBookmarkSettings action (Implements ActionSettings interface)
func (settings CreateBookmarkSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
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

	// Get or create current selection object
	sessionState.QueueRequest(func(ctx context.Context) error {
		if _, err := uplink.CurrentApp.GetCurrentSelections(sessionState, actionState); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}, actionState, true, "failed to create CurrentSelection object")

	sessionState.Wait(actionState)

	fields := uplink.CurrentApp.GetAggregatedSelectionFields()

	sheetID := ""
	if !settings.NoSheetLocation {
		// Find out the current sheet from the object map
		sheets := uplink.Objects.GetAllObjectHandles(false, enigmahandlers.ObjTypeSheet)
		if len(sheets) == 0 {
			actionState.AddErrors(errors.New("no sheet in current context: a sheet must be selected to be create a bookmark"))
			return
		}
		if len(sheets) > 1 {
			actionState.AddErrors(errors.New("more than one sheet in current context"))
			return
		}
		sheetHandle := sheets[0]
		sheet := uplink.Objects.Load(sheetHandle)

		sheetID = sheet.ID
	}

	title, err := sessionState.ReplaceSessionVariables(&settings.Title)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	// Mirrors the fields in the SDK
	props := map[string]interface{}{
		"sheetId":         sheetID,
		"selectionFields": fields,
		"creationDate":    time.Now().Format("01/02/06 "), // US short date format
		"qMetaDef":        creation.StubMetaDef(title, settings.Description),
		"qInfo":           creation.StubNxInfo("bookmark"),
	}

	var requestToSend func(context.Context) (*enigma.GenericBookmark, error)
	if settings.SaveLayout {
		requestToSend = func(ctx context.Context) (*enigma.GenericBookmark, error) {
			return uplink.CurrentApp.Doc.CreateBookmarkExRaw(ctx, props, []string{})
		}
	} else {
		requestToSend = func(ctx context.Context) (*enigma.GenericBookmark, error) {
			return uplink.CurrentApp.Doc.CreateBookmarkRaw(ctx, props)
		}
	}

	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		bookmark, err := requestToSend(ctx)
		if err != nil {
			return err
		}
		if bookmark == nil {
			return errors.Errorf("creating of bookmark<%s> resulted in empty object", settings.ID)
		}
		if settings.ID != "" {
			return sessionState.IDMap.Add(settings.ID, bookmark.GenericId, sessionState.LogEntry)
		}

		return nil
	}); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to CreateBookmark"))
	}

	sessionState.Wait(actionState)
}

// AffectsAppObjectsAction implements AffectsAppObjectsAction interface
func (settings CreateBookmarkSettings) AffectsAppObjectsAction(structure appstructure.AppStructure) ([]*appstructure.AppStructurePopulatedObjects, []string, bool) {
	if settings.ID == "" {
		return nil, nil, false
	}
	newObjs := appstructure.AppStructurePopulatedObjects{
		Parent:    settings.ID,
		Bookmarks: make([]appstructure.AppStructureBookmark, 0),
	}
	newObjs.Bookmarks = append(newObjs.Bookmarks, appstructure.AppStructureBookmark{
		ID:          settings.ID,
		Title:       settings.Title.String(),
		Description: settings.Description, //TODO: Should specify current sheet
	})
	return []*appstructure.AppStructurePopulatedObjects{&newObjs}, nil, false
}
