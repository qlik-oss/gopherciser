package objecthandling

import (
	"context"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	AutoChartHandler struct{}
)

// GetObject implement ObjectHandler interface
func (handler *AutoChartHandler) SetObjectAndEvents(sessionState *session.State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		return getObjectLayout(sessionState, actionState, obj)
	}, actionState, true, "")

	handleAutoChart(sessionState, actionState, genObj, obj)
}

// DoSelect implement ObjectHandler interface
func (handler *AutoChartHandler) DoSelect() error {
	return nil
}

// ObjectChanged implement ObjectHandler interface
func (handler *AutoChartHandler) ObjectChanged() error {
	return nil
}
