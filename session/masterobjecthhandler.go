package session

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/senseobjdef"
)

type (
	MasterObjectHandler struct{}

	MasterObjectHandlerInstance struct {
		ID            string
		Visualization string
	}

	GenericObjectLayoutWithVisualization struct {
		enigma.GenericObjectLayout
		Visualization string
	}
)

// Instance implements ObjectHandler  interface
func (handler *MasterObjectHandler) Instance(id string) ObjectHandlerInstance {
	return &MasterObjectHandlerInstance{ID: id}
}

// GetObjectDefinition implements ObjectHandlerInstance interface
func (handler *MasterObjectHandlerInstance) GetObjectDefinition(objectType string) (string, senseobjdef.SelectType, senseobjdef.DataDefType, error) {
	if objectType != "masterobject" {
		return "", senseobjdef.SelectTypeUnknown, senseobjdef.DataDefUnknown, errors.New("MasterObjectHandlerInstance only handles objects of type masterobject")
	}

	return (&DefaultHandlerInstance{}).GetObjectDefinition(handler.Visualization)
}

// SetObjectAndEvents
func (handler *MasterObjectHandlerInstance) SetObjectAndEvents(sessionState *State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		return GetObjectProperties(sessionState, actionState, obj)
	}, actionState, true, "")

	layout := GetMasterObjectLayout(sessionState, actionState, genObj)
	if layout == nil {
		return // error occured and has been reported on actionState
	}
	handler.Visualization = layout.Visualization

	def, err := senseobjdef.GetObjectDef(handler.Visualization)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	DefaultSetObjectDataAndEvents(sessionState, actionState, obj, genObj, def)
}

// GetMasterObjectLayout returns
func GetMasterObjectLayout(sessionState *State, actionState *action.State, containerObject *enigma.GenericObject) *GenericObjectLayoutWithVisualization {
	rawLayout, err := sessionState.SendRequestRaw(actionState, containerObject.GetLayoutRaw)
	if err != nil {
		actionState.AddErrors(err)
		return nil
	}

	var layout GenericObjectLayoutWithVisualization
	if err := json.Unmarshal(rawLayout, &layout); err != nil {
		actionState.AddErrors(err)
		return nil
	}

	return &layout
}
