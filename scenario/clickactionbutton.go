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
	// ClickActionButtonSettings click action-button settings
	ClickActionButtonSettings struct {
		// ID object id
		ID string `json:"id" appstructure:"active:action-button" displayname:"Button ID" doc-key:"clickactionbutton.id"`
	}
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

	buttonActionType int
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

func (buttonActionType) getEnumMap() *enummap.EnumMap {
	return buttonActionTypeEnumMap
}

func (value *buttonActionType) UnmarshalJSON(jsonBytes []byte) error {
	var actionStr string
	if err := json.Unmarshal(jsonBytes, &actionStr); err != nil {
		return err
	}
	i, ok := value.getEnumMap().AsInt()[strings.ToLower(actionStr)]
	if !ok {
		*value = unknownAction
	}
	*value = buttonActionType(i)
	return nil
}

func (value buttonActionType) MarshalJSON() ([]byte, error) {
	str, err := value.getEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown selectiontype<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// String representation of ListBoxSelectType
func (value buttonActionType) String() string {
	sType, err := value.getEnumMap().String(int(value))
	if err != nil {
		return strconv.Itoa(int(value))
	}
	return sType
}

// Validate filter pane select action
func (settings ClickActionButtonSettings) Validate() error {
	if settings.ID == "" {
		return errors.Errorf("Empty object ID")
	}

	return nil
}

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
	buttonActions, err := buttonActions(sessionState, actionState, obj)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "Could not get button actions"))
		return
	}

	for _, buttonAction := range buttonActions {
		if err := buttonAction.execute(sessionState, actionState); err != nil {
			actionState.AddErrors(errors.Wrapf(err, "Buttonaction type<%s> label<%s> cid<%s> failed",
				buttonAction.ActionType, buttonAction.ActionLabel, buttonAction.CID))
		}
	}

	sessionState.Wait(actionState)
}

func buttonActions(sessionState *session.State, actionState *action.State, obj *enigmahandlers.Object) ([]buttonAction, error) {
	// TODO(atluq): Add buttonActions to enigma.GenericObjectProperties such
	// that obj.Properties() could be used, instead of
	// obj.EnigmaObject.GetpropertiesRaw()
	switch t := obj.EnigmaObject.(type) {
	case *enigma.GenericObject:
		genericObj := obj.EnigmaObject.(*enigma.GenericObject)

		propsRaw, err := sessionState.SendRequestRaw(actionState, genericObj.GetPropertiesRaw)
		if err != nil {
			return nil, errors.Wrapf(err, "Properties request failed for object<%s>", obj.ID)
		}

		actionsRaw, err := senseobjdef.NewDataPath("actions").Lookup(propsRaw)
		if err != nil {
			return nil, errors.Wrapf(err, `No "actions"-property exist for object<%s>`, obj.ID)
		}

		// println(string(actionsRaw))
		actions := []buttonAction{}
		err = json.Unmarshal(actionsRaw, &actions)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not unmarshal button actions")
		}

		println("actions unmarshaled")
		return actions, nil

	default:
		return nil, errors.Errorf("Expected generic object , got object type<%T>", t)
	}
}

func (buttonAction *buttonAction) execute(sessionState *session.State, actionState *action.State) error {
	app := sessionState.Connection.Sense().CurrentApp
	sendReq := func(f func(context.Context) error) error {
		return sessionState.SendRequest(actionState, f)
	}
	// TODO sense browser client does getField only once for the same field
	// here we do multiple
	// same may apply to variables
	// where are these stored?

	switch buttonAction.ActionType {
	case emptyAction: // do nothing
		return nil

	case unknownAction:
		return errors.New("Unknown button action")

	case applyBookmark:
		return sendReq(func(ctx context.Context) error {
			success, err := app.Doc.ApplyBookmark(ctx, buttonAction.Bookmark /*id*/)
			if err != nil {
				return errors.Wrapf(err, "Error applying bookmark<%s>", buttonAction.Bookmark)
			}
			if !success {
				return errors.Errorf("Unsuccessful application bookmark<%s>", buttonAction.Bookmark)
			}
			return err
		})

	case moveBackwardsInSelections:
		return sendReq(app.Doc.Back)

	case moveForwardsInSelections:
		return sendReq(app.Doc.Forward)

	case clearAllSelections:
		return sendReq(func(ctx context.Context) error {
			return app.Doc.ClearAll(ctx, buttonAction.SoftLock /*lockedAlso*/, "" /*stateName*/)
		})

	case clearSelectionsInOtherFields:
		return sendReq(func(ctx context.Context) error {
			field, err := app.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return errors.Wrapf(err, "Could not get field<%s>", buttonAction.Field)
			}
			_, err = field.ClearAllButThis(ctx, buttonAction.SoftLock)
			return err
		})

	case clearSelectionsInField:
		return sendReq(func(ctx context.Context) error {
			field, err := app.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return errors.Wrapf(err, "Could not get field<%s>", buttonAction.Field)
			}
			_, err = field.Clear(ctx)
			return err
		})

	case selectAllValuesInField:
		return sendReq(func(ctx context.Context) error {
			field, err := app.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return errors.Wrapf(err, "Failed to get field<%s>", buttonAction.Field)
			}
			_, err = field.SelectAll(ctx, buttonAction.SoftLock)
			return errors.Wrapf(err, "Could not select all values in field<%s>", buttonAction.Field)
		})

	case selectValuesInField:
		return sendReq(func(ctx context.Context) error {
			field, err := app.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return errors.Wrapf(err, "Failed to get field<%s>", buttonAction.Field)
			}
			_, err = field.SelectValuesRaw(ctx, buttonAction.Value /*TODO unambigous how to do this*/, false /*toggleMode*/, buttonAction.SoftLock)
			return errors.Wrapf(err, "Could not select values in field<%s>", buttonAction.Field)
		})

	case selectValuesMatchingSearchCriteria:
		return sendReq(func(ctx context.Context) error {
			field, err := app.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return errors.Wrapf(err, "Could not get field<%s>", buttonAction.Field)
			}

			_, err = field.Select(ctx, buttonAction.Value /*match*/, buttonAction.SoftLock, 0 /*TODO excludedValuesMode*/)
			return errors.Wrapf(err, "Selection failed in field<%s> searchcritera<%s>", buttonAction.Field, buttonAction.Value)
		})

	case selectAlternatives:
		return sendReq(func(ctx context.Context) error {
			field, err := app.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return errors.Wrapf(err, "Could not get field<%s>", buttonAction.Field)
			}
			_, err = field.SelectAlternative(ctx, buttonAction.SoftLock)
			return err
		})

	case selectExcluded:
		return sendReq(func(ctx context.Context) error {
			field, err := app.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return errors.Wrapf(err, "Could not get field<%s>", buttonAction.Field)
			}
			_, err = field.SelectExcluded(ctx, buttonAction.SoftLock)
			return err
		})

	case selectPossibleValuesInField:
		return sendReq(func(ctx context.Context) error {
			field, err := app.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return errors.Wrapf(err, "Could not get field<%s>", buttonAction.Field)
			}
			_, err = field.SelectPossible(ctx, buttonAction.SoftLock)
			return errors.Wrapf(err, "Could not select possible in field<%s>", buttonAction.Field)
		})

	case toggleFieldSelection:
		fmt.Printf("ButtonAction<%s> not implemented\n", buttonAction.ActionType)
		return nil

	case lockAllSelections:
		return sendReq(func(ctx context.Context) error {
			return app.Doc.LockAll(ctx, "" /*alternate state name*/)
		})

	case lockSpecificField:
		return sendReq(func(ctx context.Context) error {
			field, err := app.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return errors.Wrapf(err, "Could not get field<%s>", buttonAction.Field)
			}
			_, err = field.Lock(ctx)
			return err
		})

	case unlockAllSelections:
		return sendReq(func(ctx context.Context) error {
			return app.Doc.UnlockAll(ctx, "" /*alternate state name*/)
		})

	case unlockSpecificField:
		return sendReq(func(ctx context.Context) error {
			field, err := app.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return errors.Wrapf(err, "Could not get field<%s>", buttonAction.Field)
			}
			_, err = field.Unlock(ctx)
			return err
		})

	case setVariableValue:
		return sendReq(func(ctx context.Context) error {
			variable, err := app.Doc.GetVariableByName(ctx, buttonAction.Variable)
			println("var type " + buttonAction.Variable + " " + variable.Type)
			if err != nil {
				return errors.Wrapf(err, "Could not get variable<%s>", buttonAction.Variable)
			}
			return variable.SetStringValue(ctx, buttonAction.Value)
		})

	default:
		return errors.New("Unexpected buttonaction")
	}
}
