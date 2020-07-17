package scenario

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/senseobjdef"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ListBoxSelectType type of selection
	ListBoxSelectType int
	// ListBoxSelectSettings selection settings
	ListBoxSelectSettings struct {
		// ID object id
		ID string `json:"id" appstructure:"active:listbox" displayname:"Listbox ID" doc-key:"listboxselect.id"`
		// Type selection type
		Type ListBoxSelectType `json:"type" displayname:"Selection type" doc-key:"listboxselect.type"`
		// Accept true - confirm selection. false - abort selection
		Accept bool `json:"accept" displayname:"Accept selection" doc-key:"listboxselect.accept"`
		// Wrap selection with Begin / End selection requests
		Wrap bool `json:"wrap" displayname:"Wrap selection" doc-key:"listboxselect.wrap"`
	}
)

const (
	// All select all
	All ListBoxSelectType = iota
	// Possible select possible
	Possible
	// Alternative select alternative
	Alternative
	// Excluded select excluded
	Excluded
)

var listBoxSelectTypeEnumMap, _ = enummap.NewEnumMap(map[string]int{
	"all":         int(All),
	"possible":    int(Possible),
	"alternative": int(Alternative),
	"excluded":    int(Excluded),
})

func (ListBoxSelectType) GetEnumMap() *enummap.EnumMap {
	return listBoxSelectTypeEnumMap
}

// UnmarshalJSON unmarshal filter pane selection type
func (value *ListBoxSelectType) UnmarshalJSON(json []byte) error {
	i, err := value.GetEnumMap().UnMarshal(json)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal SelectionType")
	}
	*value = ListBoxSelectType(i)
	return nil
}

// MarshalJSON marshal filter pane selection type
func (value ListBoxSelectType) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown selectiontype<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// String representation of ListBoxSelectType
func (value ListBoxSelectType) String() string {
	sType, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return strconv.Itoa(int(value))
	}
	return sType
}

// Validate filter pane select action
func (settings ListBoxSelectSettings) Validate() error {
	if settings.ID == "" {
		return errors.Errorf("Empty object ID")
	}

	return nil
}

// Execute filter pane select action
func (settings ListBoxSelectSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	uplink := sessionState.Connection.Sense()
	objectID := sessionState.IDMap.Get(settings.ID)
	obj, err := uplink.Objects.GetObjectByID(objectID)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "Failed getting object<%s> from object list", objectID))
		return
	}

	switch t := obj.EnigmaObject.(type) {
	case *enigma.GenericObject:
		genericObj := obj.EnigmaObject.(*enigma.GenericObject)
		settings.doSelect(sessionState, actionState, genericObj)
	default:
		actionState.AddErrors(errors.Errorf("Expected generic object , got object type<%T>", t))
		return
	}
	sessionState.Wait(actionState)
}

func (settings ListBoxSelectSettings) doSelect(sessionState *session.State, actionState *action.State, genericObj *enigma.GenericObject) {
	objInstance := sessionState.GetObjectHandlerInstance(genericObj.GenericId, genericObj.GenericType)

	selectPath, selectType, dataDefType, err := objInstance.GetObjectDefinition(genericObj.GenericType)
	if err != nil {
		actionState.AddErrors(err)
		return
	}
	if selectType != senseobjdef.SelectTypeListObjectValues {
		actionState.AddErrors(errors.Errorf("Object id<%s> type<%s> selecttype<%s> datadeftype<%s> must have selecttype<%s> in listboxselect-action",
			genericObj.GenericId, genericObj.Type, selectType, dataDefType, senseobjdef.SelectTypeListObjectValues))
		return
	}

	var selectFunction func(context.Context, string, bool) (bool, error)
	switch settings.Type {
	case All:
		selectFunction = genericObj.SelectListObjectAll
	case Alternative:
		selectFunction = genericObj.SelectListObjectAlternative
	case Excluded:
		selectFunction = genericObj.SelectListObjectExcluded
	case Possible:
		selectFunction = genericObj.SelectListObjectPossible
	default:
		actionState.AddErrors(errors.Wrapf(err, "Unknown filter pane selection type ListBoxSelectType<%s>", settings.Type))
		return
	}

	if settings.Wrap {
		beginSelections := func(ctx context.Context) error {
			err := genericObj.BeginSelections(ctx, []string{selectPath})
			return errors.Wrapf(err, "Failed to begin selsection in object<%s>", genericObj.GenericId)
		}
		err = sessionState.SendRequest(actionState, beginSelections)
		if err != nil {
			actionState.AddErrors(err)
			return
		}
	}

	selectListObjects := func(ctx context.Context) error {
		sessionState.LogEntry.LogDebugf("Listboxselect in object<%s> selecttype<%s> handle<%d> type<%s>", genericObj.GenericId, selectType, genericObj.Handle, genericObj.GenericType)

		success, err := selectFunction(ctx, selectPath, false)
		if err != nil {
			return errors.Wrapf(err, "Failed to perform lisboxselect in object<%s>", genericObj.GenericId)
		}
		if !success {
			return errors.Errorf("Listboxselect in object<%s> unsuccessful", genericObj.GenericId)
		}

		sessionState.LogEntry.LogDebug(fmt.Sprint("Successful select in", genericObj.GenericId))

		return err
	}

	err = sessionState.SendRequest(actionState, selectListObjects)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	if settings.Wrap {
		endSelections := func(ctx context.Context) error {
			err := genericObj.EndSelections(ctx, settings.Accept)
			return errors.Wrapf(err, "Failed to end selsection in object<%s>", genericObj.GenericId)
		}
		err = sessionState.SendRequest(actionState, endSelections)
		if err != nil {
			actionState.AddErrors(err)
			return
		}
	}
}
