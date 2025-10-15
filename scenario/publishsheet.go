package scenario

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// PublishSheetMode specifies the mode for publishing the sheets
	PublishSheetMode int
	// PublishSheetSettings contains details for publishing sheet(s)
	PublishSheetSettings struct {
		Mode             PublishSheetMode `json:"mode" displayname:"Publish mode" doc-key:"publishsheet.mode"`
		SheetIDs         []string         `json:"sheetIds" displayname:"Sheet IDs" doc-key:"publishsheet.sheetIds"`
		IncludePublished bool             `json:"includePublished" displayname:"Try to publish already published sheets" doc-key:"publishsheet.includePublished"`
		Thinktime        time.Duration    `json:"thinktime" displayname:"Think time" doc-key:"publishsheet.thinktime"`
	}
)

const (
	// AllSheets publishes all of the sheets in the opened app
	AllSheets PublishSheetMode = iota
	// SheetIDs publishes sheets specified in the sheetIds array
	SheetIDs
)

const DefaultPublishSheetThinkTime = 100 * time.Millisecond

var publishSheetModeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
	"allsheets": int(AllSheets),
	"sheetids":  int(SheetIDs),
})

func (value PublishSheetMode) GetEnumMap() *enummap.EnumMap {
	return publishSheetModeEnumMap
}

// UnmarshalJSON unmarshal PublishSheetMode
func (value *PublishSheetMode) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal PublishSheetMode")
	}

	*value = PublishSheetMode(i)
	return nil
}

// MarshalJSON marshal PublishSheetMode type
func (value PublishSheetMode) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("unknown PublishSheetMode<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// Execute performs the publish sheet action
func (publishSheetSettings PublishSheetSettings) Execute(sessionState *session.State,
	actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {

	publishAction := func(sheet *senseobjects.Sheet, ctx context.Context) error {
		sheetLayout, err := sheet.GetLayout(ctx)
		if err != nil {
			return errors.Wrapf(err, `failed get layout of sheet<%s>`, sheet.ID)
		}
		published := sheetLayout.Meta.Published
		title := sheetLayout.Meta.Title
		accessLevel := "private"
		if published {
			accessLevel = "public"
		}

		if published {
			if !publishSheetSettings.IncludePublished {
				sessionState.LogDebugf(
					`not publishing sheet<%s> "%s" since it is already %s`,
					sheet.ID, title, accessLevel,
				)
				return nil
			}
			sessionState.LogEntry.Logf(
				logger.WarningLevel,
				`trying to publish already published sheet<%s> "%s"`,
				sheet.ID, title,
			)
		}

		sessionState.LogDebugf(`publishing %s sheet<%s> "%s"`, accessLevel, sheet.ID, title)

		err = sheet.Publish(ctx)
		if err != nil {
			return errors.Wrapf(
				err,
				`failed to publish %s sheet<%s> "%s"`,
				accessLevel, sheet.ID, title,
			)
		}
		return nil
	}

	thinkTime := time.Duration(publishSheetSettings.Thinktime)
	if thinkTime < 1 {
		sessionState.LogDebugf("no thinktime set, using default<%s>", DefaultPublishSheetThinkTime)
		thinkTime = DefaultPublishSheetThinkTime
	}

	executePubUnPubAction(publishSheetSettings.Mode, publishSheetSettings.SheetIDs, thinkTime,
		publishAction, "failed to publish sheet(s)",
		sessionState, actionState)
}

// executePubUnPubAction executes the publish/un-publish sheet action
func executePubUnPubAction(mode PublishSheetMode, sheetIDs []string, thinkTime time.Duration,
	action func(*senseobjects.Sheet, context.Context) error, errMsg string,
	sessionState *session.State,
	actionState *action.State) {

	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("not connected to a Sense environment"))
		return
	}

	app := sessionState.Connection.Sense().CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("not connected to a Sense app"))
		return
	}

	sheetList, err := sessionState.Connection.Sense().CurrentApp.GetSheetList(sessionState, actionState)
	if err != nil {
		actionState.AddErrors(err)
		return
	}
	items := sheetList.Layout().AppObjectList.Items

	allSheetIDs := make([]string, 0, len(items))
	for _, item := range items {
		allSheetIDs = append(allSheetIDs, item.Info.Id)
	}

	var selectedIDs []string
	if mode == AllSheets {
		selectedIDs = allSheetIDs
	} else {
		sort.StringSlice(allSheetIDs).Sort()
		for _, sheetId := range sheetIDs {
			sheetId = sessionState.IDMap.Get(sheetId)
			idx := sort.StringSlice(allSheetIDs).Search(sheetId)
			if idx == len(allSheetIDs) || sheetId != allSheetIDs[idx] {
				actionState.AddErrors(errors.Errorf("sheet <%v> not found in the app", sheetId))
				return
			}
			selectedIDs = append(selectedIDs, sheetId)
		}
	}

	for i, sheetId := range selectedIDs {
		var sheetObject *senseobjects.Sheet
		sheetId = sessionState.IDMap.Get(sheetId)
		getSheet := func(ctx context.Context) error {
			var err error
			sheetObject, err = senseobjects.GetSheet(ctx, app, sheetId)
			return err
		}

		if i != 0 {
			<-time.After(time.Duration(thinkTime))
		}

		if err := sessionState.SendRequest(actionState, getSheet); err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return
		}

		sessionState.QueueRequest(func(ctx context.Context) error {
			if err := action(sheetObject, ctx); err != nil {
				return errors.WithStack(err)
			}
			return nil
		}, actionState, true, errMsg)
	}

	sessionState.Wait(actionState)
}

// Validate checks the settings of the publish sheet action
func (publishSheetSettings PublishSheetSettings) Validate() ([]string, error) {
	return validatePubUnPubSettings(publishSheetSettings.Mode, publishSheetSettings.SheetIDs)
}

// validatePubUnPubSettings validates the publish/un-publish sheet action settings
func validatePubUnPubSettings(mode PublishSheetMode, sheetIds []string) ([]string, error) {
	if mode == SheetIDs && len(sheetIds) == 0 {
		return nil, errors.New("no sheet ids specified")
	}
	if mode == AllSheets && len(sheetIds) > 0 {
		return nil, errors.New("sheet ids should not be specified")
	}
	return nil, nil
}
