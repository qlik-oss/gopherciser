package session

import (
	"context"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
)

type (
	AutoChartInstance struct {
		DefaultHandlerInstance
	}

	AutoChartHandler struct {
		DefaultHandler
	}
)

func (handler *AutoChartHandler) Instance(id string) ObjectHandlerInstance {
	return &AutoChartInstance{DefaultHandlerInstance{Id: id}}
}

// GetObject implement ObjectHandler interface
func (instance *AutoChartInstance) SetObjectAndEvents(sessionState *State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject) {
	instance.GetObjectLayout(sessionState, actionState, obj)
	handleAutoChart(sessionState, actionState, genObj, obj)
}

// GetObjectLayout for auto-chart main object
func (instance *AutoChartInstance) GetObjectLayout(sessionState *State, actionState *action.State, obj *enigmahandlers.Object) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		enigmaObject, ok := obj.EnigmaObject.(*enigma.GenericObject)
		if !ok {
			return errors.Errorf("Failed to cast object<%s> to *enigma.GenericObject", obj.ID)
		}
		sessionState.LogEntry.LogDebugf("Getting layout for object<%s> handle<%d> type<%s> START", obj.ID, obj.Handle, enigmaObject.GenericType)

		rawLayout, layoutErr := sessionState.SendRequestRaw(actionState, enigmaObject.GetLayoutRaw)
		if layoutErr != nil {
			return errors.Wrapf(layoutErr, "object<%s>.GetLayout", enigmaObject.GenericId)
		}

		if err := SetChildList(rawLayout, obj); err != nil {
			return errors.Wrapf(err, "failed to get childlist for object<%s> type<%s>", obj.ID, enigmaObject.GenericType)
		}
		if err := SetChildren(rawLayout, obj); err != nil {
			return errors.Wrapf(err, "failed to get children for object<%s> type<%s>", obj.ID, enigmaObject.GenericType)
		}

		sessionState.LogEntry.LogDebugf("Getting layout for object<%s> handle<%d> type<%s> END", obj.ID, obj.Handle, enigmaObject.GenericType)

		return nil
	}, actionState, true, "")
}
