package scenario

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
)

func subscribeSheetObjectsAsync(sessionState *session.State, actionState *action.State, app *senseobjects.App, sheetID string) error {
	sheetID = sessionState.IDMap.Get(sheetID)
	sheetEntry, err := GetSheetEntry(sessionState, actionState, app, sheetID)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to objects")
	}

	for _, v := range sheetEntry.Data.Cells {
		sessionState.LogEntry.LogDebugf("subscribe to object<%s> type<%s>", v.Name, v.Type)
		session.GetAndAddObjectAsync(sessionState, actionState, v.Name)
	}

	return nil
}

func GetSheetEntry(sessionState *session.State, actionState *action.State, app *senseobjects.App, sheetid string) (*senseobjects.SheetNxContainerEntry, error) {
	sheetList, err := app.GetSheetList(sessionState, actionState)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return sheetList.GetSheetEntry(sheetid)
}

// GetCurrentSheet from objects
func GetCurrentSheet(uplink *enigmahandlers.SenseUplink) (*senseobjects.Sheet, error) {
	sheets := uplink.Objects.GetObjectsOfType(enigmahandlers.ObjTypeSheet)
	if len(sheets) < 1 {
		return nil, errors.New("no current sheet found")
	}
	if len(sheets) > 1 {
		return nil, errors.Errorf("%d current sheets found", len(sheets))
	}
	sheetObj, ok := sheets[0].EnigmaObject.(*senseobjects.Sheet)
	if !ok {
		return nil, errors.Errorf("failed to cast object id<%s> to sheet object", sheetObj.GenericId)
	}
	return sheetObj, nil
}

func DebugPrintObjectSubscriptions(sessionState *session.State) {
	if !sessionState.LogEntry.ShouldLogDebug() {
		return
	}

	upLink := sessionState.Connection.Sense()
	objectsPointers := upLink.Objects.GetObjectsOfType(enigmahandlers.ObjTypeGenericObject)
	objects := make([]string, 0, len(objectsPointers))
	for _, object := range objectsPointers {
		if object == nil {
			continue
		}
		objects = append(objects, object.ID)
	}
	sessionState.LogEntry.LogDebug(fmt.Sprintf("current object subscriptions: %v", objects))
}

// Contains check whether any element in the supplied list matches (match func(s string) bool)
func Contains(list []string, match func(s string) bool) bool {
	for _, item := range list {
		if match(item) {
			return true
		}
	}
	return false
}

// IndexOf returns index of first match in stringSlice or else -1
func IndexOf(match string, stringSlice []string) (int, bool) {
	for i, str := range stringSlice {
		if str == match {
			return i, true
		}
	}
	return -1, false
}

type (
	varReq   func(ctx context.Context, varName string) (*enigma.GenericVariable, error)
	fieldReq func(ctx context.Context, fieldName string, stateName string) (*enigma.Field, error)
)

func (getField fieldReq) WithCache(fc *enigmahandlers.FieldCache) fieldReq {
	return func(ctx context.Context, fieldName string, stateName string) (*enigma.Field, error) {
		if field, hit := fc.Lookup(fieldName); hit {
			return field, nil
		}
		field, err := getField(ctx, fieldName, stateName)
		if err != nil {
			return field, err
		}
		fc.Store(fieldName, field)
		return field, nil
	}
}

func (getVar varReq) WithCache(vc *enigmahandlers.VarCache) varReq {
	return func(ctx context.Context, varName string) (*enigma.GenericVariable, error) {
		if variable, hit := vc.Lookup(varName); hit {
			return variable, nil
		}
		variable, err := getVar(ctx, varName)
		if err != nil {
			return variable, err
		}
		vc.Store(varName, variable)
		return variable, nil
	}
}

func executeThinkTimeSubAction(sessionState *session.State, connection *connection.ConnectionSettings, label string, thinkTimeSettings *ThinkTimeSettings) error {
	if thinkTimeSettings == nil {
		return nil
	}
	return executeSubAction(sessionState, connection, ActionThinkTime, label, thinkTimeSettings)
}

func executeSubAction(sessionState *session.State, connectionSettings *connection.ConnectionSettings,
	actionType string, label string, settings ActionSettings) error {

	action := Action{
		ActionCore: ActionCore{
			Type:  actionType,
			Label: label,
		},
		Settings: settings,
	}
	return action.Execute(sessionState, connectionSettings)
}

func executeSubActionFuncFactory(baseSessionState *session.State, baseConnection *connection.ConnectionSettings, baseActionType, baseLabel string) func(actionType string, settings ActionSettings) error {
	return func(subActionType string, settings ActionSettings) error {
		if baseLabel == "" {
			baseLabel = baseActionType
		}
		return executeSubAction(
			baseSessionState,
			baseConnection,
			fmt.Sprintf(`%s-%s`, baseActionType, subActionType),
			fmt.Sprintf(`%s: %s`, baseLabel, subActionType),
			settings,
		)
	}
}

type executeFunc func(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func())

func (execute executeFunc) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	execute(sessionState, actionState, connectionSettings, label, reset)
}

func (execute executeFunc) Validate() ([]string, error) {
	if execute == nil {
		return nil, errors.New("executeFunc is nil")
	}
	return nil, nil
}

func thinkTimeWithFallback(thinkTime *ThinkTimeSettings, fallback ThinkTimeSettings) *ThinkTimeSettings {
	if thinkTime == nil || thinkTime.Type == helpers.StaticDistribution && thinkTime.Delay == 0 {
		return &fallback
	}
	aCopy := *thinkTime
	return &aCopy
}
