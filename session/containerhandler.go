package session

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjdef"
)

type (
	ContainerHandler struct{}

	ContainerChildReference struct {
		RefID    string
		ObjID    string
		External bool
	}

	ContainerHandlerInstance struct {
		ID       string
		ActiveID string
		Children []ContainerChildReference
	}

	ContainerExternal struct {
		MasterID string `json:"masterId"`
		App      string `json:"app"`
		ViewID   string `json:"viewId"`
	}

	ContainerChild struct {
		RefID             string             `json:"refId"`
		Label             string             `json:"label"`
		IsMaster          bool               `json:"isMaster"`
		ExternalReference *ContainerExternal `json:"externalReference"`
		Type              string             `json:"type"`
	}

	ContainerChildItemData struct {
		Title            string `json:"title"`
		Visualization    string `json:"visualization"`
		ContainerChildId string `json:"containerChildId"`
		ExtendsId        string `json:"qExtendsId"`
		ShowCondition    string `json:"showCondition"`
	}

	ContainerChildItem struct {
		Info enigma.NxInfo          `json:"qInfo"`
		Meta interface{}            `json:"qMeta"`
		Data ContainerChildItemData `json:"qData"`
	}

	ContainerChildList struct {
		Items []ContainerChildItem `json:"qItems"`
	}

	ContainerLayout struct {
		Children  []ContainerChild   `json:"children"`
		ChildList ContainerChildList `json:"qChildList"`
	}
)

// Instance implements ObjectHandler  interface
func (handler *ContainerHandler) Instance(id string) ObjectHandlerInstance {
	return &ContainerHandlerInstance{ID: id}
}

// SetObjectAndEvents implements ObjectHandlerInstance interface
func (handler *ContainerHandlerInstance) SetObjectAndEvents(sessionState *State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		return GetObjectProperties(sessionState, actionState, obj)
	}, actionState, true, "")

	rawLayout, err := sessionState.SendRequestRaw(actionState, genObj.GetLayoutRaw)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	// TODO external objects is in children but not in qChildList
	// TODO externalReference refID not in same app

	var layout ContainerLayout
	if err := jsonit.Unmarshal(rawLayout, &layout); err != nil {
		actionState.AddErrors(err)
		return
	}

	// Create map with reference and the object id
	refMap := make(map[string]string, len(layout.ChildList.Items))
	for _, item := range layout.ChildList.Items {
		ref := item.Data.ContainerChildId
		if ref == "" {
			ref = item.Data.ExtendsId
		}
		if ref == "" {
			actionState.AddErrors(errors.Errorf("failed to find reference for object<%s>", item.Info.Id))
			continue
		}
		refMap[ref] = item.Info.Id
	}

	// Return resolving any child got an error
	if actionState.Failed {
		return
	}

	// Create child array with same order as .Children, this is the order of the tabs
	handler.Children = make([]ContainerChildReference, 0, len(layout.Children))
	for _, child := range layout.Children {
		ccr := ContainerChildReference{RefID: child.RefID, ObjID: refMap[child.RefID]}
		if child.ExternalReference != nil {
			ccr.External = true
		}
		handler.Children = append(handler.Children, ccr)
	}

	// Get layout on object changed
	event := func(ctx context.Context, as *action.State) error {
		_, err := genObj.GetLayoutRaw(ctx)
		return errors.Wrap(err, fmt.Sprintf("failed to get layout for container object<%s>", genObj.GenericId))
	}
	sessionState.RegisterEvent(genObj.Handle, event, nil, true)

	// First tab according to "Children" is the default active tab
	if len(handler.Children) > 0 {
		activeChild := handler.Children[0]
		if !activeChild.External {
			handler.ActiveID = handler.Children[0].ObjID
			// Subscribe to active object
			GetAndAddObjectAsync(sessionState, actionState, handler.ActiveID)
		} else {
			sessionState.LogEntry.Log(logger.WarningLevel, "container contains external reference, external references are not supported")
		}
	}
}

// GetObjectDefinition implements ObjectHandlerInstance interface
func (handler *ContainerHandlerInstance) GetObjectDefinition(objectType string) (string, senseobjdef.SelectType, senseobjdef.DataDefType, error) {
	if objectType != "container" {
		return "", senseobjdef.SelectTypeUnknown, senseobjdef.DataDefUnknown, errors.New("ContainerHandlerInstance only handles objects of type container")
	}
	return (&DefaultHandlerInstance{}).GetObjectDefinition("container")
}

// SwitchActiveID unsubscribes from the current activeid and subscribes to the new one
func (handler *ContainerHandlerInstance) SwitchActiveID(sessionState *State, actionState *action.State, activeID string) error {
	found := false
	for _, child := range handler.Children {
		if child.ObjID == activeID {
			found = true
		}
	}

	if !found {
		return errors.Errorf("could not find object<%s> as a child to container<%s>", activeID, handler.ID)
	}

	if handler.ActiveID != activeID {
		if handler.ActiveID != "" {
			if err := sessionState.ClearSubscribedObjects([]string{handler.ActiveID}); err != nil {
				return errors.WithStack(err)
			}
		}
		handler.ActiveID = activeID
		GetAndAddObjectAsync(sessionState, actionState, handler.ActiveID)
	}
	return nil
}
