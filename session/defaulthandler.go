package session

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/senseobjdef"
)

type (
	DefaultHandlerInstance struct {
		Id string
	}
	DefaultHandler struct{}
)

func (handler *DefaultHandler) Instance(id string) ObjectHandlerInstance {
	return &DefaultHandlerInstance{
		Id: id,
	}
}

// GetObject implement ObjectHandler interface
func (instance *DefaultHandlerInstance) SetObjectAndEvents(sessionState *State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject) {
	SetObjectDataAndEvents(sessionState, actionState, obj, genObj)

	children := obj.ChildList()
	childListItems := make(map[string]interface{})
	if children != nil && children.Items != nil {
		sessionState.LogEntry.LogDebugf("object<%s> type<%s> has children", genObj.GenericId, genObj.GenericType)
		for _, child := range children.Items {
			sessionState.LogEntry.LogDebug(fmt.Sprintf("obj<%s> child<%s> found in ChildList", obj.ID, child.Info.Id))
			childListItems[child.Info.Id] = nil
			GetAndAddObjectAsync(sessionState, actionState, child.Info.Id)
		}
	}

	if genObj.GenericType == "sheet" {
		sessionState.QueueRequest(func(ctx context.Context) error {
			sheetList, err := sessionState.Connection.Sense().CurrentApp.GetSheetList(sessionState, actionState)
			if err != nil {
				return errors.WithStack(err)
			}
			if sheetList != nil {
				entry, err := sheetList.GetSheetEntry(genObj.GenericId)
				if err != nil {
					return errors.WithStack(err)
				}
				if entry != nil && entry.Data != nil {
					for _, cell := range entry.Data.Cells {
						if _, ok := childListItems[cell.Name]; !ok {
							// Todo should this be a warning?
							sessionState.LogEntry.LogDebug(fmt.Sprintf("cell<%s> missing from sheet<%s> childlist", cell.Name, genObj.GenericId))
							GetAndAddObjectAsync(sessionState, actionState, cell.Name)
						}
					}
				}
			}
			return nil
		}, actionState, true, "")
	}
}

func (instance *DefaultHandlerInstance) GetObjectDefinition(objectType string) (string, senseobjdef.SelectType, senseobjdef.DataDefType, error) {
	def, defErr := senseobjdef.GetObjectDef(objectType)
	if defErr != nil {
		return "", senseobjdef.SelectTypeUnknown, senseobjdef.DataDefUnknown, errors.Wrapf(defErr, "Failed to get object<%s> selection definitions", objectType)
	}

	if validateErr := def.Validate(); validateErr != nil {
		return "", senseobjdef.SelectTypeUnknown, senseobjdef.DataDefUnknown, errors.Wrapf(validateErr, "Error validating object<%s> selection definitions<%+v>", objectType, def)
	}

	selectType := senseobjdef.SelectTypeUnknown
	selectPath := ""

	if def.Select != nil {
		selectPath = def.Select.Path
		selectType = def.Select.Type
	}

	return selectPath, selectType, def.DataDef.Type, nil
}
