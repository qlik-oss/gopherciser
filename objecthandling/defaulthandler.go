package objecthandling

import (
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
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

// DoSelect implement ObjectHandler interface
func (handler *DefaultHandler) DoSelect() error {
	return nil
}

// ObjectChanged implement ObjectHandler interface
func (handler *DefaultHandler) ObjectChanged() error {
	return nil
}
