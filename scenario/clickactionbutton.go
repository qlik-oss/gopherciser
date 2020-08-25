package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/senseobjdef"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ClickActionButtonSettings implements the ActionSettings and
	// ContainerAction interfaces. Executing this action replicates the
	// behaviour of clicking an "action-button" in Sense.
	ClickActionButtonSettings struct {
		// ID object id
		ID string `json:"id" appstructure:"active:action-button" displayname:"Button ID" doc-key:"clickactionbutton.id"`
	}
)

type (
	// *** Sub actions of an actionButton ***
	// An actionButton can contain multiple buttonActions and one buttonNavigationAction.
	// These sub-actions are executed in order when the an action-button is clicked.

	buttonActionType int

	buttonAction struct {
		ActionLabel         string           `json:"actionLabel"`
		ActionType          buttonActionType `json:"actionType"`
		Bookmark            string           `json:"bookmark"`
		Field               string           `json:"field"`
		Variable            string           `json:"variable"`
		ShowSystemVariables bool             `json:"showSystemVariables"`
		SoftLock            bool             `json:"softLock"`
		Value               string           `json:"value"`
		CID                 string           `json:"cId"`
	}

	buttonNavigationActionType int

	buttonNavigationAction struct {
		Action     buttonNavigationActionType `json:"action"`
		Sheet      string                     `json:"sheet"`
		Story      string                     `json:"story"`
		WebsiteURL string                     `json:"websiteUrl"`
		SameWindow bool                       `json:"sameWindow"`
	}
)

const (
	emptyAction buttonActionType = iota
	unknownAction
	applyBookmark
	moveBackwardsInSelections
	moveForwardsInSelections
	clearAllSelections
	clearSelectionsInOtherFields
	clearSelectionsInField
	selectAllValuesInField
	selectValuesInField
	selectValuesMatchingSearchCriteria
	selectAlternatives
	selectExcluded
	selectPossibleValuesInField
	toggleFieldSelection
	lockAllSelections
	lockSpecificField
	unlockAllSelections
	unlockSpecificField
	setVariableValue
)

const (
	emptyNavAction buttonNavigationActionType = iota
	noneNavAction
	unknownNavAction
	firstSheet
	lastSheet
	previousSheet
	nextSheet
	goToSheet
	goToSheetByID
	goToStory
	openWebsite
)

// IsContainerAction implements ContainerAction interface
// and sets container action logging to original action entry
func (settings ClickActionButtonSettings) IsContainerAction() {}

// Validate filter pane select action
func (settings ClickActionButtonSettings) Validate() error {
	if settings.ID == "" {
		return errors.Errorf("empty object ID")
	}
	return nil
}

// Execute button-actions contained by sense action-button
func (settings ClickActionButtonSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("not connected to a Sense environment"))
		return
	}

	if sessionState.Connection.Sense().CurrentApp == nil {
		actionState.AddErrors(errors.New("not connected to a Sense app"))
		return
	}

	// retrieve action-button-object
	objectID := sessionState.IDMap.Get(settings.ID)
	obj, err := sessionState.Connection.Sense().Objects.GetObjectByID(objectID)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed getting object<%s> from object list", objectID))
		return
	}

	// retrieve button-actions
	buttonActions, navigationAction, err := buttonActions(sessionState, actionState, obj)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to get button actions"))
		return
	}

	// run button-actions
	for _, buttonAction := range buttonActions {
		err := buttonAction.execute(sessionState, actionState, connectionSettings, label)
		if err != nil {
			actionState.AddErrors(errors.Wrapf(err, "buttonaction type<%s> label<%s> cid<%s> failed",
				buttonAction.ActionType, buttonAction.ActionLabel, buttonAction.CID))
		}
	}

	err = navigationAction.execute(sessionState, actionState, connectionSettings, label)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "button-navigationaction<%s> failed", navigationAction.Action))
	}

	sessionState.Wait(actionState)
}

// buttonAction returns button actions and navigation action for obj
func buttonActions(sessionState *session.State, actionState *action.State, obj *enigmahandlers.Object) ([]buttonAction, *buttonNavigationAction, error) {
	// TODO(atluq): Add buttonActions to enigma.GenericObjectProperties such
	// that obj.Properties() could be used, instead of
	// obj.EnigmaObject.GetpropertiesRaw()

	switch t := obj.EnigmaObject.(type) {
	case *enigma.GenericObject:
		genericObj := obj.EnigmaObject.(*enigma.GenericObject)

		propsRaw, err := sessionState.SendRequestRaw(actionState, genericObj.GetPropertiesRaw)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "properties request failed for object<%s>", obj.ID)
		}

		// parse button actions
		actionsRaw, err := senseobjdef.NewDataPath("actions").Lookup(propsRaw)
		if err != nil {
			return nil, nil, errors.Wrapf(err, `no "actions"-property exist for object<%s>`, obj.ID)
		}

		actions := []buttonAction{}
		err = json.Unmarshal(actionsRaw, &actions)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to unmarshal button actions")
		}

		// parse navigation action associated with button
		navigationRaw, err := senseobjdef.NewDataPath("navigation").Lookup(propsRaw)
		if err != nil {
			return nil, nil, errors.Wrapf(err, `no "navigation"-property exist for object<%s>`, obj.ID)
		}

		navigationAction := &buttonNavigationAction{}
		err = json.Unmarshal(navigationRaw, navigationAction)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to unmarshal button navigation action")
		}

		return actions, navigationAction, nil

	default:
		return nil, nil, errors.Errorf("expected generic object , got object type<%T>", t)
	}
}

func executeContainerAction(sessionState *session.State, connectionSettings *connection.ConnectionSettings,
	actionType string, settings ActionSettings) error {

	if err := settings.Validate(); err != nil {
		return errors.Wrapf(err, "%s settings not valid", actionType)
	}
	action := Action{
		ActionCore: ActionCore{
			Type:  actionType,
			Label: fmt.Sprintf("button-action-%s", actionType),
		},
		Settings: settings,
	}
	if isAborted, err := CheckActionError(action.Execute(sessionState, connectionSettings)); isAborted {
		return errors.Wrapf(err, "%s was aborted", actionType)
	} else if err != nil {
		return errors.Wrapf(err, "%s failed", actionType)
	}
	return nil
}

// execute one action contained by a Sense action-button
func (buttonAction *buttonAction) execute(sessionState *session.State, actionState *action.State,
	connectionSettings *connection.ConnectionSettings, label string) error {

	uplink := sessionState.Connection.Sense()
	doc := uplink.CurrentApp.Doc
	sendReq := func(f func(context.Context) error) error {
		return sessionState.SendRequest(actionState, f)
	}
	executeContainerAction := func(actionType string, settings ActionSettings) error {
		return executeContainerAction(sessionState, connectionSettings, actionType, settings)
	}

	switch buttonAction.ActionType {
	case emptyAction:
		return nil

	case unknownAction:
		return errors.New("unknown button action")

	// container button-actions:

	case applyBookmark:
		return executeContainerAction(
			ActionApplyBookmark,
			&ApplyBookmarkSettings{
				BookMarkSettings{
					ID: buttonAction.Bookmark,
				},
			},
		)

	case clearAllSelections:
		return executeContainerAction(ActionClearAll, &ClearAllSettings{})

	// TODO(atluq) the following buttonactions shall become container actions
	// when implemetations of individual the actions exist.
	// NOT container button-actions:

	case moveBackwardsInSelections:
		return sendReq(doc.Back)

	case moveForwardsInSelections:
		return sendReq(doc.Forward)

	case clearSelectionsInOtherFields:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Field == "" {
				return nil
			}
			field, err := fieldReq(doc.GetField).WithCache(&uplink.FieldCache)(ctx, buttonAction.Field, "")
			if err != nil {
				return errors.WithStack(err)
			}
			success, err := field.ClearAllButThis(ctx, buttonAction.SoftLock)
			return errors.Wrapf(checkSuccess(success, err), `failed action<%s> in field<%s>`, buttonAction.ActionType, buttonAction.Field)
		})

	case clearSelectionsInField:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Field == "" {
				return nil
			}
			field, err := fieldReq(doc.GetField).WithCache(&uplink.FieldCache)(ctx, buttonAction.Field, "")
			if err != nil {
				return errors.WithStack(err)
			}
			success, err := field.Clear(ctx)
			return errors.Wrapf(checkSuccess(success, err), `failed action<%s> in field<%s>`, buttonAction.ActionType, buttonAction.Field)
		})

	case selectAllValuesInField:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Field == "" {
				return nil
			}
			field, err := fieldReq(doc.GetField).WithCache(&uplink.FieldCache)(ctx, buttonAction.Field, "")
			if err != nil {
				return errors.WithStack(err)
			}
			success, err := field.SelectAll(ctx, buttonAction.SoftLock)
			return errors.Wrapf(checkSuccess(success, err), `failed action<%s> in field<%s>`, buttonAction.ActionType, buttonAction.Field)
		})

	case selectValuesInField:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Field == "" {
				return nil
			}
			values := toFieldValues(buttonAction.Value)
			if len(values) == 0 {
				return nil
			}
			field, err := fieldReq(doc.GetField).WithCache(&uplink.FieldCache)(ctx, buttonAction.Field, "")
			if err != nil {
				return errors.WithStack(err)
			}
			// GetFieldDescription here, just to mimic Sense client
			if _, err = doc.GetFieldDescription(ctx, buttonAction.Field); err != nil {
				return errors.WithStack(err)
			}
			success, err := field.SelectValues(ctx, values, false /*toggleMode*/, buttonAction.SoftLock)
			return errors.Wrapf(checkSuccess(success, err), `failed action<%s> in field<%s>`, buttonAction.ActionType, buttonAction.Field)
		})

	case selectValuesMatchingSearchCriteria:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Field == "" {
				return nil
			}
			field, err := fieldReq(doc.GetField).WithCache(&uplink.FieldCache)(ctx, buttonAction.Field, "")
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = field.Select(ctx, buttonAction.Value /*match*/, buttonAction.SoftLock, 0 /*excludedValuesMode*/)
			return errors.Wrapf(err, "selection failed in field<%s> searchcritera<%s>", buttonAction.Field, buttonAction.Value)
		})

	case selectAlternatives:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Field == "" {
				return nil
			}
			field, err := fieldReq(doc.GetField).WithCache(&uplink.FieldCache)(ctx, buttonAction.Field, "")
			if err != nil {
				return errors.WithStack(err)
			}
			success, err := field.SelectAlternative(ctx, buttonAction.SoftLock)
			return errors.Wrapf(checkSuccess(success, err), `failed action<%s> in field<%s>`, buttonAction.ActionType, buttonAction.Field)
		})

	case selectExcluded:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Field == "" {
				return nil
			}
			field, err := fieldReq(doc.GetField).WithCache(&uplink.FieldCache)(ctx, buttonAction.Field, "")
			if err != nil {
				return errors.WithStack(err)
			}
			success, err := field.SelectExcluded(ctx, buttonAction.SoftLock)
			return errors.Wrapf(checkSuccess(success, err), `failed action<%s> in field<%s>`, buttonAction.ActionType, buttonAction.Field)
		})

	case selectPossibleValuesInField:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Field == "" {
				return nil
			}
			field, err := fieldReq(doc.GetField).WithCache(&uplink.FieldCache)(ctx, buttonAction.Field, "")
			if err != nil {
				return errors.WithStack(err)
			}
			success, err := field.SelectPossible(ctx, buttonAction.SoftLock)
			return errors.Wrapf(checkSuccess(success, err), `failed action<%s> in field<%s>`, buttonAction.ActionType, buttonAction.Field)
		})

	case toggleFieldSelection:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Field == "" {
				return nil
			}
			field, err := fieldReq(doc.GetField).WithCache(&uplink.FieldCache)(ctx, buttonAction.Field, "")
			if err != nil {
				return errors.WithStack(err)
			}
			success, err := field.ToggleSelect(ctx, buttonAction.Value /*match*/, buttonAction.SoftLock, 0 /*excludedValuesMode*/)
			return errors.Wrapf(checkSuccess(success, err), `failed action<%s> in field<%s>`, buttonAction.ActionType, buttonAction.Field)
		})

	case lockAllSelections:
		return sendReq(func(ctx context.Context) error {
			return doc.LockAll(ctx, "" /*stateName*/)
		})

	case lockSpecificField:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Field == "" {
				return nil
			}
			field, err := fieldReq(doc.GetField).WithCache(&uplink.FieldCache)(ctx, buttonAction.Field, "")
			if err != nil {
				return errors.WithStack(err)
			}
			success, err := field.Lock(ctx)
			return errors.Wrapf(checkSuccess(success, err), `failed action<%s> in field<%s>`, buttonAction.ActionType, buttonAction.Field)
		})

	case unlockAllSelections:
		return sendReq(func(ctx context.Context) error {
			return doc.UnlockAll(ctx, "" /*stateName*/)
		})

	case unlockSpecificField:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Field == "" {
				return nil
			}
			field, err := fieldReq(doc.GetField).WithCache(&uplink.FieldCache)(ctx, buttonAction.Field, "")
			if err != nil {
				return errors.WithStack(err)
			}
			success, err := field.Unlock(ctx)
			return errors.Wrapf(checkSuccess(success, err), `failed action<%s> in field<%s>`, buttonAction.ActionType, buttonAction.Field)
		})

	case setVariableValue:
		return sendReq(func(ctx context.Context) error {
			if buttonAction.Variable == "" {
				return nil
			}
			variable, err := varReq(doc.GetVariableByName).WithCache(&uplink.VarCache)(ctx, buttonAction.Variable)
			if err != nil {
				return errors.WithStack(err)
			}
			return variable.SetStringValue(ctx, buttonAction.Value)
		})

	default:
		return errors.New("unexpected buttonaction")
	}
}

// execute the navigation-action contained in a Sense action-button
func (navAction *buttonNavigationAction) execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string) error {
	switch navAction.Action {
	case noneNavAction:
		return nil
	case emptyNavAction:
		return errors.New("empty button navigation action")
	case unknownNavAction:
		return errors.New("unknown button navigation action")
	}

	sheets, err := sheetIDs(sessionState, actionState)
	if err != nil {
		return errors.Wrapf(err, "error getting sheets")
	}
	if len(sheets) == 0 {
		return errors.New("no sheets in app")
	}
	currentSheet, err := GetCurrentSheet(sessionState.Connection.Sense())
	if err != nil {
		return errors.Wrapf(err, "could not get current sheet")
	}

	var newCurrentSheetID string

	switch navAction.Action {

	case firstSheet:
		newCurrentSheetID = sheets[0]

	case lastSheet:
		newCurrentSheetID = sheets[len(sheets)-1]

	case nextSheet:
		currentSheetIdx, ok := IndexOf(currentSheet.ID, sheets)
		if !ok {
			return errors.Wrapf(err, "could not index of get current sheet")
		}
		nextSheetIdx := (currentSheetIdx + 1) % len(sheets)
		newCurrentSheetID = sheets[nextSheetIdx]

	case previousSheet:
		currentSheetIdx, ok := IndexOf(currentSheet.ID, sheets)
		if !ok {
			return errors.Wrapf(err, "could not index of get current sheet")
		}
		previousSheetIdx := (currentSheetIdx - 1 + len(sheets)) % len(sheets)
		newCurrentSheetID = sheets[previousSheetIdx]

	case goToSheet, goToSheetByID:
		if navAction.Sheet == "" {
			return errors.New("empty sheet id")
		}
		newCurrentSheetID = navAction.Sheet

	default:
		return errors.Errorf(`button navigation action "%s" is not supported`, navAction.Action)
	}

	if newCurrentSheetID == currentSheet.ID {
		return nil
	}

	// Execute changeSheet Action
	changeSheetAction := Action{
		ActionCore{
			Type:  ActionChangeSheet,
			Label: fmt.Sprintf("button-navigation-%s", navAction.Action),
		},
		&ChangeSheetSettings{
			ID: newCurrentSheetID,
		},
	}
	if isAborted, err := CheckActionError(changeSheetAction.Execute(sessionState, connectionSettings)); isAborted {
		return errors.Wrap(err, "change sheet was aborted")
	} else if err != nil {
		return errors.Wrap(err, "change sheet failed")
	}
	return nil
}

func sheetIDs(sessionState *session.State, actionState *action.State) ([]string, error) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil || sessionState.Connection.Sense().CurrentApp == nil {
		return nil, errors.New("not connected to a Sense app")
	}

	sheetList, err := sessionState.Connection.Sense().CurrentApp.GetSheetList(sessionState, actionState)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	items := sheetList.Layout().AppObjectList.Items
	sheetIDs := make([]string, 0, len(items))
	for _, item := range items {
		sheetIDs = append(sheetIDs, item.Info.Id)
	}
	return sheetIDs, err
}

// toFieldValues converts ';'-separated string to fieldvalues
func toFieldValues(fields string) []*enigma.FieldValue {
	stringValues := strings.Split(fields, ";")
	values := make([]*enigma.FieldValue, 0, len(stringValues))
	for _, sv := range stringValues {
		nbr, err := strconv.ParseFloat(sv, 64)

		// if parse error which is not range error, append text value and continue
		if err != nil {
			numErr := err.(*strconv.NumError).Err
			switch numErr {
			case strconv.ErrRange:
				// do nothing, nbr == ±inf is handled by enigma.Float64()
			default:
				values = append(values, &enigma.FieldValue{
					Text:      sv,
					IsNumeric: false,
				})
				continue
			}
		}

		values = append(values, &enigma.FieldValue{
			Number:    enigma.Float64(nbr),
			IsNumeric: true,
		})
	}
	return values
}

// Enum, MutableEnum, IntegerEnum, fmt.Stringer, json.Marshaler and
// json.Unmarshaler implementations for buttonActionType

var buttonActionTypeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
	"":                     int(emptyAction),
	"unknownaction":        int(unknownAction),
	"applybookmark":        int(applyBookmark),
	"back":                 int(moveBackwardsInSelections),
	"forward":              int(moveForwardsInSelections),
	"clearall":             int(clearAllSelections),
	"clearallbutthis":      int(clearSelectionsInOtherFields),
	"clearfield":           int(clearSelectionsInField),
	"selectall":            int(selectAllValuesInField),
	"selectvalues":         int(selectValuesInField),
	"selectmatchingvalues": int(selectValuesMatchingSearchCriteria),
	"selectalternative":    int(selectAlternatives),
	"selectexcluded":       int(selectExcluded),
	"selectpossible":       int(selectPossibleValuesInField),
	"toggleselect":         int(toggleFieldSelection),
	"lockall":              int(lockAllSelections),
	"lockfield":            int(lockSpecificField),
	"unlockall":            int(unlockAllSelections),
	"unlockfield":          int(unlockSpecificField),
	"setvariable":          int(setVariableValue),
})

func (buttonActionType) GetEnumMap() *enummap.EnumMap {
	return buttonActionTypeEnumMap
}

func (value buttonActionType) Int() int {
	return int(value)
}

func (value *buttonActionType) Set(i int) {
	*value = buttonActionType(i)
}

func (value *buttonActionType) UnmarshalJSON(jsonBytes []byte) error {
	return UnmarshalJSON(value, jsonBytes)
}

func (value buttonActionType) MarshalJSON() ([]byte, error) {
	return MarshalJSON(value)
}

func (value buttonActionType) String() string {
	return String(value)
}

// Enum, MutableEnum, IntegerEnum, fmt.Stringer, json.Marshaler and
// json.Unmarshaler implementations for buttonNavigationActionType

var buttonNavActionTypeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
	"":                 int(emptyAction),
	"unknownnavaction": int(unknownNavAction),
	"none":             int(noneNavAction),
	"firstsheet":       int(firstSheet),
	"lastsheet":        int(lastSheet),
	"previoussheet":    int(previousSheet),
	"nextsheet":        int(nextSheet),
	"gotosheet":        int(goToSheet),
	"gotosheetbyid":    int(goToSheetByID),
	"gotostory":        int(goToStory),
	"openwebsite":      int(openWebsite),
})

func (buttonNavigationActionType) GetEnumMap() *enummap.EnumMap {
	return buttonNavActionTypeEnumMap
}

func (value buttonNavigationActionType) Int() int {
	return int(value)
}

func (value *buttonNavigationActionType) Set(i int) {
	*value = buttonNavigationActionType(i)
}

func (value *buttonNavigationActionType) UnmarshalJSON(jsonBytes []byte) error {
	return UnmarshalJSON(value, jsonBytes)
}

func (value buttonNavigationActionType) MarshalJSON() ([]byte, error) {
	return MarshalJSON(value)
}

func (value buttonNavigationActionType) String() string {
	return String(value)
}

func checkSuccess(success bool, err error) error {
	if err != nil {
		return errors.WithStack(err)
	}
	if !success {
		return errors.Errorf("unsuccessful operation")
	}
	return nil
}
