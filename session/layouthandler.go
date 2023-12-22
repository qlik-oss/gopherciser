package session

import (
	"context"
	"fmt"
	"sync"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjdef"
)

type (
	LayoutContainerHandler struct{}

	LayoutContainerHandlerInstance struct {
		ID string

		lock sync.Mutex
	}

	LayoutContainerLayout struct {
		Objects   []LayoutContainerLayoutObject `json:"objects"`
		ChildList LayoutContainerChildList      `json:"qChildList"`
	}

	LayoutContainerLayoutObject struct {
		ChildRefId string             `json:"childRefId"`
		Condition  *helpers.FuzzyBool `json:"condition"`
	}

	LayoutContainerChildList struct {
		Items []LayoutContainerChildItem `json:"qItems"`
	}

	LayoutContainerChildItem struct {
		Info enigma.NxInfo                `json:"qInfo"`
		Meta interface{}                  `json:"qMeta"`
		Data LayoutContainerChildItemData `json:"qData"`
	}

	LayoutContainerChildItemData struct {
		Title         string `json:"title"`
		Visualization string `json:"visualization"`
		ChildRefId    string `json:"childRefId"`
		ExtendsId     string `json:"qExtendsId"`
	}
)

// Instance implements ObjectHandler  interface
func (handler *LayoutContainerHandler) Instance(id string) ObjectHandlerInstance {
	return &LayoutContainerHandlerInstance{ID: id}
}

// GetObjectDefinition implements ObjectHandlerInstance interface
func (handler *LayoutContainerHandlerInstance) GetObjectDefinition(objectType string) (string, senseobjdef.SelectType, senseobjdef.DataDefType, error) {
	if objectType != "sn-layout-container" {
		return "", senseobjdef.SelectTypeUnknown, senseobjdef.DataDefUnknown, errors.New("LayoutHandlerInstance only handles objects of type sn-layout-container")
	}
	return (&DefaultHandlerInstance{}).GetObjectDefinition("sn-layout-container")
}

// SetObjectAndEvents implements ObjectHandlerInstance interface
func (handler *LayoutContainerHandlerInstance) SetObjectAndEvents(sessionState *State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		return GetObjectProperties(sessionState, actionState, obj)
	}, actionState, true, "")

	layout := GetLayourContainerLayout(sessionState, actionState, genObj)
	if layout == nil {
		return // error occured and has been reported on actionState
	}

	if err := handler.UpdateChildren(sessionState, actionState, layout); err != nil {
		actionState.AddErrors(err)
		return
	}

	// Get layout on object changed
	event := func(ctx context.Context, as *action.State) error {
		sessionState.LogEntry.Logf(logger.DebugLevel, "Getting layout for object<%s> handle<%d> type<%s>", genObj.GenericId, genObj.Handle, genObj.GenericType)
		layout := GetLayourContainerLayout(sessionState, as, genObj)
		if as.Failed {
			return nil // error occured, but has been reported
		}

		if err := handler.UpdateChildren(sessionState, as, layout); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to update children for container object<%s>", genObj.GenericId))
		}

		return nil
	}
	sessionState.RegisterEvent(genObj.Handle, event, nil, true)
}

// UpdateChildren of the container
func (handler *LayoutContainerHandlerInstance) UpdateChildren(sessionState *State, actionState *action.State, layout *LayoutContainerLayout) error {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	// Set / add latest list to true
	activeChildRefs := make([]string, 0, len(layout.Objects))
	inactiveChildRefs := make([]string, 0, len(layout.Objects))
	for _, child := range layout.Objects {
		if child.Condition.AsBool() {
			activeChildRefs = append(activeChildRefs, child.ChildRefId)
		} else {
			inactiveChildRefs = append(activeChildRefs, child.ChildRefId)
		}
	}

	// Unsubscribe to any deactivated objects
	inactiveChildren, err := layout.ChildRefsToIDs(inactiveChildRefs)
	if err != nil {
		return errors.Wrapf(err, "layout object<%s>", handler.ID)
	}
	upLink := sessionState.Connection.Sense()
	unsubscribeObjects := make([]string, 0, len(inactiveChildren))
	for _, objId := range inactiveChildren {
		if _, err := upLink.Objects.GetObjectByID(objId); err == nil {
			unsubscribeObjects = append(unsubscribeObjects, objId)
		}
	}
	if err := sessionState.ClearSubscribedObjects(unsubscribeObjects); err != nil {
		return errors.WithStack(err)
	}

	// Subscribe to any activated objects
	activeChildren, err := layout.ChildRefsToIDs(activeChildRefs)
	if err != nil {
		return errors.Wrapf(err, "layout object<%s>", handler.ID)
	}
	for _, objId := range activeChildren {
		if _, err := upLink.Objects.GetObjectByID(objId); err != nil {
			GetAndAddObjectAsync(sessionState, actionState, objId)
		}
	}

	return nil
}

// GetLayourContainerLayout returns unmarshaled sn-layout-container layout, errors reported on action state
func GetLayourContainerLayout(sessionState *State, actionState *action.State, containerObject *enigma.GenericObject) *LayoutContainerLayout {
	rawLayout, err := sessionState.SendRequestRaw(actionState, containerObject.GetLayoutRaw)
	if err != nil {
		actionState.AddErrors(err)
		return nil
	}

	var layout LayoutContainerLayout
	if err := json.Unmarshal(rawLayout, &layout); err != nil {
		actionState.AddErrors(err)
		return nil
	}

	return &layout
}

// ChildRefsToIDs translates a list of child refs to object ID's
func (layout *LayoutContainerLayout) ChildRefsToIDs(childRefs []string) ([]string, error) {
	objectIds := make([]string, 0, len(childRefs))
	for _, refID := range childRefs {
		objId := ""
		for _, child := range layout.ChildList.Items {
			if child.Data.ChildRefId == refID {
				objId = child.Data.ExtendsId
				if objId == "" {
					objId = child.Info.Id
				}
				break
			}
		}
		if objId == "" {
			return nil, errors.Errorf("failed to find object id for child ref<%s>", refID)
		}
		objectIds = append(objectIds, objId)
	}
	return objectIds, nil
}
