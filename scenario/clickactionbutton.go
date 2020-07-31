package scenario

import (
	"context"
	"encoding/json"
	"fmt"

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
	// ClickActionButtonSettings click action-button settings
	ClickActionButtonSettings struct {
		// ID object id
		ID string `json:"id" appstructure:"active:action-button" displayname:"Button ID" doc-key:"clickactionbutton.id"`
	}
)
type (
	// *** Sub actions of an actionButton ***
	// An actionButton can contain several buttonActions and one buttonNavigationAction

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
)

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
})

func (buttonActionType) GetEnumMap() *enummap.EnumMap {
	return buttonActionTypeEnumMap
}

func (buttonNavigationActionType) GetEnumMap() *enummap.EnumMap {
	return buttonNavActionTypeEnumMap
}

// Validate filter pane select action
func (settings ClickActionButtonSettings) Validate() error {
	if settings.ID == "" {
		return errors.Errorf("Empty object ID")
	}
	return nil
}

// IsContainerAction implements ContainerAction interface
// and sets container action logging to original action entry
func (settings ClickActionButtonSettings) IsContainerAction() {}

// Execute button action
func (settings ClickActionButtonSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment"))
		return
	}

	if sessionState.Connection.Sense().CurrentApp == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense app"))
		return
	}

	uplink := sessionState.Connection.Sense()
	objectID := sessionState.IDMap.Get(settings.ID)
	obj, err := uplink.Objects.GetObjectByID(objectID)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "Failed getting object<%s> from object list", obj.ID))
		return
	}

	buttonActions, navigationAction, err := buttonActions(sessionState, actionState, obj)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "Failed to get button actions"))
		return
	}

	for _, buttonAction := range buttonActions {
		err := buttonAction.execute(sessionState, actionState)
		if err != nil {
			actionState.AddErrors(errors.Wrapf(err, "Buttonaction type<%s> label<%s> cid<%s> failed",
				buttonAction.ActionType, buttonAction.ActionLabel, buttonAction.CID))
		}
	}

	err = navigationAction.execute(sessionState, actionState, connectionSettings, label)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "Button-navigationaction<%s> failed", navigationAction.Action))
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
			return nil, nil, errors.Wrapf(err, "Properties request failed for object<%s>", obj.ID)
		}

		// parse button actions
		actionsRaw, err := senseobjdef.NewDataPath("actions").Lookup(propsRaw)
		if err != nil {
			return nil, nil, errors.Wrapf(err, `No "actions"-property exist for object<%s>`, obj.ID)
		}

		actions := []buttonAction{}
		err = json.Unmarshal(actionsRaw, &actions)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Failed to unmarshal button actions")
		}

		// parse navigation action associated with button
		navigationRaw, err := senseobjdef.NewDataPath("navigation").Lookup(propsRaw)
		if err != nil {
			return nil, nil, errors.Wrapf(err, `No "navigation"-property exist for object<%s>`, obj.ID)
		}

		navigationAction := &buttonNavigationAction{}
		err = json.Unmarshal(navigationRaw, navigationAction)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Failed to unmarshal button navigation action")
		}

		return actions, navigationAction, nil

	default:
		return nil, nil, errors.Errorf("Expected generic object , got object type<%T>", t)
	}
}

func (buttonAction *buttonAction) execute(sessionState *session.State, actionState *action.State) error {
	doc := DocWrapper{sessionState.Connection.Sense().CurrentApp.Doc}
	sendReq := func(f func(context.Context) error) error {
		return sessionState.SendRequest(actionState, f)
	}
	// TODO sense browser client does getField only once for the same field
	// here we do multiple
	// same may apply to variables
	// where should these be stored?

	switch buttonAction.ActionType {
	case emptyAction: // do nothing
		return nil

	case unknownAction:
		return errors.New("Unknown button action")

	case applyBookmark:
		return sendReq(func(ctx context.Context) error {
			success, err := doc.ApplyBookmark(ctx, buttonAction.Bookmark /*id*/)
			if err != nil {
				return errors.Wrapf(err, "Error applying bookmark<%s>", buttonAction.Bookmark)
			}
			if !success {
				return errors.Errorf("Unsuccessful application bookmark<%s>", buttonAction.Bookmark)
			}
			return err
		})

	case moveBackwardsInSelections:
		return sendReq(doc.Back)

	case moveForwardsInSelections:
		return sendReq(doc.Forward)

	case clearAllSelections:
		return sendReq(func(ctx context.Context) error {
			return doc.ClearAll(ctx, buttonAction.SoftLock /*lockedAlso*/, "" /*stateName*/)
		})

	case clearSelectionsInOtherFields:
		return sendReq(func(ctx context.Context) error {
			field, err := doc.GetField(ctx, buttonAction.Field)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = field.ClearAllButThis(ctx, buttonAction.SoftLock)
			return errors.WithStack(err)
		})

	case clearSelectionsInField:
		return sendReq(func(ctx context.Context) error {
			field, err := doc.GetField(ctx, buttonAction.Field)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = field.Clear(ctx)
			return err
		})

	case selectAllValuesInField:
		return sendReq(func(ctx context.Context) error {
			field, err := doc.GetField(ctx, buttonAction.Field)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = field.SelectAll(ctx, buttonAction.SoftLock)
			return errors.Wrapf(err, "Could not select all values in field<%s>", buttonAction.Field)
		})

	case selectValuesInField:
		return sendReq(func(ctx context.Context) error {
			field, err := doc.GetField(ctx, buttonAction.Field)
			if err != nil {
				return errors.WithStack(err)
			}
			// TODO: OBS! not fininshed. how to distingush listvalues
			values := []*enigma.FieldValue{
				{
					Text:      buttonAction.Value,
					IsNumeric: false,
					Number:    0,
				},
			}
			_, err = field.SelectValues(ctx, values /*TODO ambigous how to do this*/, false /*toggleMode*/, buttonAction.SoftLock)
			return errors.Wrapf(err, "Could not select values in field<%s>", buttonAction.Field)
		})

	case selectValuesMatchingSearchCriteria:
		return sendReq(func(ctx context.Context) error {
			field, err := doc.GetField(ctx, buttonAction.Field)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = field.Select(ctx, buttonAction.Value /*match*/, buttonAction.SoftLock, 0 /*excludedValuesMode*/)
			return errors.Wrapf(err, "Selection failed in field<%s> searchcritera<%s>", buttonAction.Field, buttonAction.Value)
		})

	case selectAlternatives:
		return sendReq(func(ctx context.Context) error {
			field, err := doc.GetField(ctx, buttonAction.Field)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = field.SelectAlternative(ctx, buttonAction.SoftLock)
			return err
		})

	case selectExcluded:
		return sendReq(func(ctx context.Context) error {
			field, err := doc.GetField(ctx, buttonAction.Field)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = field.SelectExcluded(ctx, buttonAction.SoftLock)
			return err
		})

	case selectPossibleValuesInField:
		return sendReq(func(ctx context.Context) error {
			field, err := doc.GetField(ctx, buttonAction.Field)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = field.SelectPossible(ctx, buttonAction.SoftLock)
			return errors.Wrapf(err, "Could not select possible in field<%s>", buttonAction.Field)
		})

	case toggleFieldSelection:
		return sendReq(func(ctx context.Context) error {
			field, err := doc.GetField(ctx, buttonAction.Field)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = field.ToggleSelect(ctx, buttonAction.Value /*match*/, buttonAction.SoftLock, 0 /*excludedValuesMode*/)
			return errors.Wrapf(err, "Could not select possible in field<%s>", buttonAction.Field)
		})

	case lockAllSelections:
		return sendReq(func(ctx context.Context) error {
			return doc.LockAll(ctx, "" /*stateName*/)
		})

	case lockSpecificField:
		return sendReq(func(ctx context.Context) error {
			field, err := doc.GetField(ctx, buttonAction.Field)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = field.Lock(ctx)
			return err
		})

	case unlockAllSelections:
		return sendReq(func(ctx context.Context) error {
			return doc.UnlockAll(ctx, "" /*stateName*/)
		})

	case unlockSpecificField:
		return sendReq(func(ctx context.Context) error {
			field, err := doc.GetField(ctx, buttonAction.Field)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = field.Unlock(ctx)
			return err
		})

	case setVariableValue:
		return sendReq(func(ctx context.Context) error {
			variable, err := doc.GetVariableByName(ctx, buttonAction.Variable)
			if err != nil {
				return errors.WithStack(err)
			}
			return variable.SetStringValue(ctx, buttonAction.Value)
		})

	default:
		return errors.New("Unexpected buttonaction")
	}
}

func (navAction *buttonNavigationAction) execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string) error {
	if navAction.Action == noneNavAction {
		return nil
	}

	sheets, err := sheetIDs(sessionState, actionState)
	if err != nil {
		return errors.Wrapf(err, "Error getting sheets")
	}
	if len(sheets) == 0 {
		return errors.New("No sheets")
	}

	var sheetID string

	switch navAction.Action {

	case firstSheet:
		sheetID = sheets[0]

	case lastSheet:
		sheetID = sheets[len(sheets)-1]

	case nextSheet:
		currentSheet, err := GetCurrentSheet(sessionState.Connection.Sense())
		if err != nil {
			return errors.Wrapf(err, "Could not get current sheet")
		}
		currentSheetIdx, ok := IndexOf(currentSheet.ID, sheets)
		if !ok {
			return errors.Wrapf(err, "Could not get current sheet")
		}
		nextSheetIdx := (currentSheetIdx + 1) % len(sheets)
		if nextSheetIdx != currentSheetIdx {
			sheetID = sheets[nextSheetIdx]
		}

	case previousSheet:
		currentSheet, err := GetCurrentSheet(sessionState.Connection.Sense())
		if err != nil {
			return errors.Wrapf(err, "Could not get current sheet")
		}
		currentSheetIdx, ok := IndexOf(currentSheet.ID, sheets)
		if !ok {
			return errors.Wrapf(err, "Could not get current sheet")
		}
		previousSheetIdx := (currentSheetIdx - 1 + len(sheets)) % len(sheets)
		if previousSheetIdx != currentSheetIdx {
			sheetID = sheets[previousSheetIdx]
		}

	case goToSheet, goToSheetByID:
		if navAction.Sheet == "" {
			return errors.New("Empty sheet id")
		}
		sheetID = navAction.Sheet

	default:
		return errors.Errorf("Unknown button navigation action '%s' ", navAction.Action)
	}
	changeSheetAction := Action{
		ActionCore{
			Type:  ActionChangeSheet,
			Label: fmt.Sprintf("button-navigation-%s", navAction.Action),
		},
		&ChangeSheetSettings{
			ID: sheetID,
		},
	}
	if isAborted, err := CheckActionError(changeSheetAction.Execute(sessionState, connectionSettings)); isAborted {
		return errors.Wrapf(err, "Change sheet was aborted")
	} else if err != nil {
		return errors.Wrapf(err, "Change sheet failed")
	}
	return nil
}

func sheetIDs(sessionState *session.State, actionState *action.State) ([]string, error) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil || sessionState.Connection.Sense().CurrentApp == nil {
		return nil, errors.New("Not connected to a Sense app")
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

func (value *buttonActionType) UnmarshalJSON(jsonBytes []byte) error {
	i, err := UnmarshalJSONUsingEnum(value, jsonBytes)
	if err != nil {
		return errors.WithStack(err)
	}
	*value = buttonActionType(i)
	return nil
}

func (value buttonActionType) MarshalJSON() ([]byte, error) {
	return MarshalJSONUsingEnum(value, int(value))
}

func (value buttonActionType) String() string {
	return StringUsingEnum(value, int(value))
}

func (value *buttonNavigationActionType) UnmarshalJSON(jsonBytes []byte) error {
	i, err := UnmarshalJSONUsingEnum(value, jsonBytes)
	if err != nil {
		return errors.WithStack(err)
	}
	*value = buttonNavigationActionType(i)
	return nil
}

func (value buttonNavigationActionType) MarshalJSON() ([]byte, error) {
	return MarshalJSONUsingEnum(value, int(value))
}

func (value buttonNavigationActionType) String() string {
	return StringUsingEnum(value, int(value))
}
