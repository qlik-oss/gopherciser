package session

import (
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
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
	DefaultSetObjectDataAndEvents(sessionState, actionState, obj, genObj, nil)
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
