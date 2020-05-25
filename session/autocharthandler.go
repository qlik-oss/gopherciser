package session

import (
	"context"
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
	sessionState.QueueRequest(func(ctx context.Context) error {
		return GetObjectLayout(sessionState, actionState, obj)
	}, actionState, true, "")

	handleAutoChart(sessionState, actionState, genObj, obj)
}
