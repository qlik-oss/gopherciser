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

	fmt.Printf("%+v\n", buttonActions)
	for _, buttonAction := range buttonActions {
		fmt.Printf("%+v\n", buttonAction)
		buttonAction.execute(sessionState, actionState, connectionSettings, label, reset)
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

		println(string(actionsRaw))
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

func (buttonAction *buttonAction) execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	switch buttonAction.ActionType {
	case emptyAction: // do nothing
	case unknownAction:
		actionState.AddErrors(errors.New("Unknown button action"))

	case applyBookmark:
		// ApplyBookmarkSettings{Title: action.Bookmark}.Execute(sessionState, actionState, connectionSettings, label, reset)
		fmt.Printf("ButtonAction<%s> not implemented\n", buttonAction.ActionType)

	case moveBackwardsInSelections:
		err := sessionState.SendRequest(actionState, sessionState.Connection.Sense().CurrentApp.Doc.Back)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to move backward in selection"))
		}

	case moveForwardsInSelections:
		err := sessionState.SendRequest(actionState, sessionState.Connection.Sense().CurrentApp.Doc.Forward)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to move forward in selection"))
		}

	case clearAllSelections:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			return sessionState.Connection.Sense().CurrentApp.Doc.ClearAll(ctx, buttonAction.SoftLock /*lockedAlso*/, "" /*stateName*/)
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to clear all selections"))
		}

	case clearSelectionsInOtherFields:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			field, err := sessionState.Connection.Sense().CurrentApp.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return err
			}
			_, err = field.ClearAllButThis(ctx, buttonAction.SoftLock)
			return err
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to clear selections in other fields"))
		}

	case clearSelectionsInField:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			field, err := sessionState.Connection.Sense().CurrentApp.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return err
			}
			_, err = field.Clear(ctx)
			return err
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to clear selections in field"))
		}

	case selectAllValuesInField:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			field, err := sessionState.Connection.Sense().CurrentApp.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return err
			}
			_, err = field.SelectAll(ctx, buttonAction.SoftLock)
			return err
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to select all values in field"))
		}

	case selectValuesInField:
		fmt.Printf("ButtonAction<%s> not implemented\n", buttonAction.ActionType)

	case selectValuesMatchingSearchCriteria:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			field, err := sessionState.Connection.Sense().CurrentApp.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return err
			}
			_, err = field.Select(ctx, "" /*TODO match*/, buttonAction.SoftLock, 0 /*TODO excludedValuesMode*/)
			return err
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to select all values in field"))
		}

	case selectAlternatives:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			field, err := sessionState.Connection.Sense().CurrentApp.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return err
			}
			_, err = field.SelectAlternative(ctx, buttonAction.SoftLock)
			return err
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to select alternative values in field"))
		}

	case selectExcluded:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			field, err := sessionState.Connection.Sense().CurrentApp.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return err
			}
			_, err = field.SelectExcluded(ctx, buttonAction.SoftLock)
			return err
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to select excluded values in field"))
		}

	case selectPossibleValuesInField:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			field, err := sessionState.Connection.Sense().CurrentApp.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return err
			}
			_, err = field.SelectExcluded(ctx, buttonAction.SoftLock)
			return err
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to select possible values in field"))
		}

	case toggleFieldSelection:
		fmt.Printf("ButtonAction<%s> not implemented\n", buttonAction.ActionType)

	case lockAllSelections:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			return sessionState.Connection.Sense().CurrentApp.Doc.LockAll(ctx, "" /*alternate state name*/)
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to lock all selections"))
		}

	case lockSpecificField:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			field, err := sessionState.Connection.Sense().CurrentApp.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return err
			}
			_, err = field.Lock(ctx)
			return err
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to lock field"))
		}

	case unlockAllSelections:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			return sessionState.Connection.Sense().CurrentApp.Doc.UnlockAll(ctx, "" /*alternate state name*/)
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to unlock all selections"))
		}

	case unlockSpecificField:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			field, err := sessionState.Connection.Sense().CurrentApp.Doc.GetField(ctx, buttonAction.Field, "" /*stateName*/)
			if err != nil {
				return err
			}
			_, err = field.Unlock(ctx)
			return err
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to unlock field"))
		}

	case setVariableValue:
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			// TODO by name or id?
			variable, err := sessionState.Connection.Sense().CurrentApp.Doc.GetVariableByName(ctx, buttonAction.Variable)
			if err != nil {
				return err
			}
			// TODO set string value or int value
			return variable.SetStringValue(ctx, buttonAction.Value)
		})
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to unlock field"))
		}
	default:
		actionState.AddErrors(errors.New("Unexpected buttonaction"))
	}

	// TODO to wait or not to wait?
}
