package scenario

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//BookmarkDeletionModeEnum defines what bookmarks to remove
	SheetDeletionModeEnum int

	//DeleteBookmarkSettings create bookmark settings
	DeleteSheetSettings struct {
		DeletionMode SheetDeletionModeEnum `json:"mode" displayname:"Deletion mode" doc-key:"deletesheet.mode"`
		Title        string                `json:"title" displayname:"Sheet title" doc-key:"deletesheet.title"`
		ID           string                `json:"id" displayname:"Sheet ID" doc-key:"deletesheet.id" appstructure:"sheet"`
	}
)

const (
	// SingleSheet delete specified sheet
	SingleSheet SheetDeletionModeEnum = iota
	// MatchingSheets delete sheets matching name
	MatchingSheets
	// AllUnpublished delete all unpublished sheets in app
	AllUnpublished
)

func (value SheetDeletionModeEnum) GetEnumMap() *enummap.EnumMap {
	enumMap, _ := enummap.NewEnumMap(map[string]int{
		"single":         int(SingleSheet),
		"matching":       int(MatchingSheets),
		"allunpublished": int(AllUnpublished),
	})
	return enumMap
}

// UnmarshalJSON unmarshal SheetDeletionModeEnum
func (value *SheetDeletionModeEnum) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal SheetDeletionModeEnum")
	}

	*value = SheetDeletionModeEnum(i)
	return nil
}

// MarshalJSON marshal SheetDeletionModeEnum type
func (value SheetDeletionModeEnum) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("unknown SheetDeletionModeEnum<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// Validate DeleteSheetSettings action (Implements ActionSettings interface)
func (settings DeleteSheetSettings) Validate() error {
	if settings.DeletionMode == SingleSheet || settings.DeletionMode == MatchingSheets {
		if (settings.Title == "") == (settings.ID == "") {
			return errors.New("either specify sheet name or sheet id")
		}
	} else if settings.Title != "" || settings.ID != "" {
		return errors.Errorf("sheet name or id cannot be specified in mode <%v>", settings.DeletionMode)
	}
	return nil
}

// Execute DeleteSheetSettings action (Implements ActionSettings interface)
func (settings DeleteSheetSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	zero := 0
	numDeleted := &zero
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("not connected to a Sense environment"))
		return
	}

	app := sessionState.Connection.Sense().CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("not connected to a Sense app"))
		return
	}

	sheetList, err := app.GetSheetList(sessionState, actionState)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	switch settings.DeletionMode {
	case SingleSheet:
		id := sessionState.IDMap.Get(settings.ID) // Lookup id in case of being a scenario key
		if id == "" {
			for _, item := range sheetList.Layout().AppObjectList.Items {
				if item.Data.Title == settings.Title {
					id = item.Info.Id
					break
				}
			}
			if id == "" {
				actionState.AddErrors(errors.Errorf("could not find sheet <%s>", settings.Title))
				return
			}
		}

		settings.destroySheetById(sessionState, app, id, actionState, numDeleted)
	case MatchingSheets:
		id := sessionState.IDMap.Get(settings.ID)
		items := sheetList.Layout().AppObjectList.Items
		for _, item := range items {
			if id == item.Info.Id || settings.Title == item.Data.Title {
				settings.destroySheetById(sessionState, app, item.Info.Id, actionState, numDeleted)
			}
		}
	case AllUnpublished:
		items := sheetList.Layout().AppObjectList.Items
		for _, i := range items {
			item := i
			sessionState.QueueRequest(func(ctx context.Context) error {
				sheet, err := senseobjects.GetSheet(ctx, app, item.Info.Id)
				if err != nil {
					return err
				}
				layout, err := sheet.GetLayout(ctx)
				if err != nil {
					return err
				}
				if !layout.Meta.Published {
					canDelete := Contains(layout.Meta.Privileges, func(s string) bool {
						return s == "delete"
					})
					if canDelete {
						settings.destroySheetById(sessionState, app, layout.Info.Id, actionState, numDeleted)
					}
				}
				return nil
			}, actionState, true, "error destroying sheet")
		}
	}

	sessionState.Wait(actionState)
	sessionState.LogEntry.LogInfo("NumDeletedSheets", fmt.Sprintf("%d", *numDeleted))
}

func (settings *DeleteSheetSettings) destroySheetById(sessionState *session.State, app *senseobjects.App, sheetId string, actionState *action.State, numDeleted *int) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		var success bool
		var err error
		if success, err = app.Doc.DestroyObject(ctx, sheetId); err != nil {
			return err
		}
		if !success {
			return errors.Errorf("failed to delete sheet with id <%s>", sheetId)
		}
		*numDeleted++
		return nil
	}, actionState, true, "error destroying sheet")
}
