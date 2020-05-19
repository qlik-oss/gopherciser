package objecthandling

import (
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/senseobjdef"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	DefaultHandler struct{}
)

// GetObject implement ObjectHandler interface
func (handler *DefaultHandler) SetObjectAndEvents(sessionState *session.State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject) {
	setObjectDataAndEvents(sessionState, actionState, obj, genObj)

	children := obj.ChildList()
	if children != nil && children.Items != nil {
		sessionState.LogEntry.LogDebugf("object<%s> type<%s> has children", genObj.GenericId, genObj.GenericType)
		for _, child := range children.Items {
			GetAndAddObjectAsync(sessionState, actionState, child.Info.Id, child.Info.Type)
		}
	}
}

func (handler *DefaultHandler) GetObjectDefinition(objectType string) (string, senseobjdef.SelectType, senseobjdef.DataDefType, error) {
	def, defErr := senseobjdef.GetObjectDef(objectType)
	if defErr != nil {
		return "", senseobjdef.SelectTypeUnknown, senseobjdef.DataDefUnknown, errors.Wrapf(defErr, "Failed to get object<%s> selection definitions", objectType)
	}

	if validateErr := def.Validate(); validateErr != nil {
		return "", senseobjdef.SelectTypeUnknown, senseobjdef.DataDefUnknown, errors.Wrapf(validateErr, "Error validating object<%s> selection definitions<%+v>", objectType, def)
	}
	return def.Select.Path, def.Select.Type, def.DataDef.Type, nil
}
