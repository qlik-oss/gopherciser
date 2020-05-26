package scenario

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// StaticSelectionType type of selection
	StaticSelectionType int
	// StaticSelectSettings selection settings
	StaticSelectSettings struct {
		//ID object id
		ID string `json:"id" displayname:"Object ID" doc-key:"staticselect.id" appstructure:"active:!sheet"`
		//Path object selection path
		Path string `json:"path" displayname:"Object selection path" doc-key:"staticselect.path"`
		//Rows to select
		Rows []int `json:"rows" displayname:"Rows to select" doc-key:"staticselect.rows"`
		//Cols columns to select
		Cols []int `json:"cols" displayname:"Columns to select" doc-key:"staticselect.cols"`
		//Type selection type
		Type StaticSelectionType `json:"type" displayname:"Selection type" doc-key:"staticselect.type"`
		//Accept true - confirm selection. false - abort selection
		Accept bool `json:"accept" displayname:"Accept selection" doc-key:"staticselect.accept"`
		//WrapSelections
		WrapSelections bool `json:"wrap" displayname:"Wrap selections" doc-key:"staticselect.wrap"`
	}
)

const (
	// HyperCubeCells select in hypercube
	HyperCubeCells StaticSelectionType = iota
	// ListObjectValues select in listbox
	ListObjectValues
)

func (value StaticSelectionType) GetEnumMap() *enummap.EnumMap {
	enumMap, _ := enummap.NewEnumMap(map[string]int{
		"hypercubecells":   int(HyperCubeCells),
		"listobjectvalues": int(ListObjectValues),
	})
	return enumMap
}

// UnmarshalJSON unmarshal selection type
func (value *StaticSelectionType) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal StaticSelectionType")
	}

	*value = StaticSelectionType(i)

	return nil
}

// MarshalJSON marshal selection type
func (value StaticSelectionType) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown StaticSelectionType<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// Validate static selection settings
func (settings StaticSelectSettings) Validate() error {
	if settings.ID == "" {
		return errors.Errorf("Empty object ID")
	}
	if settings.Path == "" {
		return errors.Errorf("Empty selection path")
	}
	if len(settings.Rows) < 1 && len(settings.Cols) < 1 {
		return errors.Errorf("Nothing selected")
	}
	return nil
}

// Execute static selection
func (settings StaticSelectSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	actionState.Details = settings.ID
	uplink := sessionState.Connection.Sense()
	gob, err := uplink.Objects.GetObjectByID(sessionState.IDMap.Get(settings.ID))
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "Failed getting object<%s> from object list", settings.ID))
		return
	}

	switch t := gob.EnigmaObject.(type) {
	case *enigma.GenericObject:
		genObj := gob.EnigmaObject.(*enigma.GenericObject)

		if settings.WrapSelections {
			// Start selections
			sessionState.QueueRequest(func(ctx context.Context) error {
				return genObj.BeginSelections(ctx, []string{settings.Path})
			}, actionState, false, fmt.Sprintf("Begin selection error: %v", err))
		}

		sessionState.Wait(actionState)
		if actionState.Errors() != nil {
			return
		}

		var selectFunc func(ctx context.Context) (bool, error)
		switch settings.Type {
		case HyperCubeCells:
			selectFunc = func(ctx context.Context) (bool, error) {
				return genObj.SelectHyperCubeCells(ctx, settings.Path, settings.Rows, settings.Cols,
					false, true)
			}
		case ListObjectValues:
			selectFunc = func(ctx context.Context) (bool, error) {
				return genObj.SelectListObjectValues(ctx, settings.Path, settings.Rows, true, false)
			}
		default:
			actionState.AddErrors(errors.Errorf("Unknown select type %v", settings.Type))
			return
		}

		// Select
		sessionState.QueueRequest(func(ctx context.Context) error {
			sessionState.LogEntry.LogDebugf("Select in object<%s> h<%d> type<%s>",
				genObj.GenericId, genObj.Handle, genObj.GenericType)
			success, err := selectFunc(ctx)
			if err != nil {
				return errors.Wrapf(err, "Failed to select in object<%s>", genObj.GenericId)
			}
			if !success {
				return errors.Errorf("Select in object<%s> unsuccessful", genObj.GenericId)
			}
			sessionState.LogEntry.LogDebug(fmt.Sprint("Successful select in", genObj.GenericId))

			if settings.WrapSelections {
				//End Selections
				sessionState.QueueRequest(func(ctx context.Context) error {
					return genObj.EndSelections(ctx, settings.Accept)
				}, actionState, true, "End selections failed")
			}

			return nil
		}, actionState, true, fmt.Sprintf("Failed to select in %s", genObj.GenericId))
	default:
		actionState.AddErrors(errors.Errorf("Unknown object type<%T>", t))
		return
	}

	sessionState.Wait(actionState)
}
