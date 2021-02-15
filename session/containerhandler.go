package session

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjdef"
)

type (
	ContainerHandler struct{}

	ContainerChildReference struct {
		RefID    string
		ObjID    string
		External bool
		Show     bool
	}

	ContainerHandlerInstance struct {
		ID       string
		ActiveID string
		children []ContainerChildReference

		lock sync.Mutex
	}

	ContainerExternal struct {
		MasterID string `json:"masterId"`
		App      string `json:"app"`
		ViewID   string `json:"viewId"`
	}

	ContainerChild struct {
		RefID             string              `json:"refId"`
		Label             string              `json:"label"`
		IsMaster          bool                `json:"isMaster"`
		ExternalReference *ContainerExternal  `json:"externalReference"`
		Type              string              `json:"type"`
		Condition         *helpers.StringBool `json:"condition"`
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

	layout := GetContainerLayout(sessionState, actionState, genObj)
	if layout == nil {
		return // error occured and has been reported on actionState
	}

	if err := handler.UpdateChildren(layout); err != nil {
		actionState.AddErrors(err)
		return
	}

	// Get layout on object changed
	event := func(ctx context.Context, as *action.State) error {
		sessionState.LogEntry.Logf(logger.DebugLevel, "Getting layout for object<%s> handle<%d> type<%s>", genObj.GenericId, genObj.Handle, genObj.GenericType)
		layout := GetContainerLayout(sessionState, as, genObj)
		if as.Failed {
			return nil // error occured, but has been reported
		}

		if err := handler.UpdateChildren(layout); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to update children for container object<%s>", genObj.GenericId))
		}

		child, current := handler.FirstShowableChild()
		childID := ""
		if child != nil {
			childID = child.ObjID
		}
		sessionState.LogEntry.Logf(logger.DebugLevel, "container<%s> first showable child<%s> active child<%s>", handler.ID, childID, handler.ActiveID)
		if current {
			return nil
		}

		// update active child
		return errors.WithStack(handler.SwitchActiveChild(sessionState, as, child))
	}
	sessionState.RegisterEvent(genObj.Handle, event, nil, true)

	child, _ := handler.FirstShowableChild()
	if child != nil {
		_ = handler.SwitchActiveChild(sessionState, actionState, child)
	}
}

// GetContainerLayout returns unmarshaled container layout, errors reported on action state
func GetContainerLayout(sessionState *State, actionState *action.State, containerObject *enigma.GenericObject) *ContainerLayout {
	rawLayout, err := sessionState.SendRequestRaw(actionState, containerObject.GetLayoutRaw)
	if err != nil {
		actionState.AddErrors(err)
		return nil
	}

	var layout ContainerLayout
	if err := jsonit.Unmarshal(rawLayout, &layout); err != nil {
		actionState.AddErrors(err)
		return nil
	}

	return &layout
}

// UpdateChildren of the container
func (handler *ContainerHandlerInstance) UpdateChildren(layout *ContainerLayout) error {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	var mErr *multierror.Error
	// Create map with reference and the object id
	refMap := make(map[string]string, len(layout.ChildList.Items))
	for _, item := range layout.ChildList.Items {
		ref := item.Data.ContainerChildId
		if ref == "" {
			ref = item.Data.ExtendsId
		}
		if ref == "" {
			mErr = multierror.Append(mErr, errors.Errorf("failed to find reference for object<%s>", item.Info.Id))
			continue
		}
		refMap[ref] = item.Info.Id
	}

	if mErr != nil {
		return helpers.FlattenMultiError(mErr)
	}

	// Create child array with same order as .Children, this is the order of the tabs
	handler.children = make([]ContainerChildReference, 0, len(layout.Children))
	for _, child := range layout.Children {
		ccr := ContainerChildReference{RefID: child.RefID, ObjID: refMap[child.RefID]}
		if child.ExternalReference != nil {
			ccr.External = true
		}
		if child.Condition == nil {
			ccr.Show = true
		} else {
			ccr.Show = bool(*child.Condition)
		}

		handler.children = append(handler.children, ccr)
	}

	return nil
}

// FirstShowableChild gets reference to first object from container which fulfills conditions, true if already active object
func (handler *ContainerHandlerInstance) FirstShowableChild() (*ContainerChildReference, bool) {
	// First check if active ID fulfills conditions
	if handler.ActiveID != "" {
		if child := handler.ActiveChildReference(); child != nil {
			if child.Show {
				return child, true
			}
		}
	}

	// ActiveChildReference also locks, don't lock before this
	handler.lock.Lock()
	defer handler.lock.Unlock()

	// find first with no condition or condition = true
	for _, child := range handler.children {
		if child.Show {
			return &child, false
		}
	}

	// no showable child found, show nothing
	return nil, false
}

// ActiveChildReference returns reference to currently active child or nil
func (handler *ContainerHandlerInstance) ActiveChildReference() *ContainerChildReference {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	if handler.ActiveID == "" {
		return nil
	}
	for _, child := range handler.children {
		if child.ObjID == handler.ActiveID {
			return &child
		}
	}
	return nil
}

// GetObjectDefinition implements ObjectHandlerInstance interface
func (handler *ContainerHandlerInstance) GetObjectDefinition(objectType string) (string, senseobjdef.SelectType, senseobjdef.DataDefType, error) {
	if objectType != "container" {
		return "", senseobjdef.SelectTypeUnknown, senseobjdef.DataDefUnknown, errors.New("ContainerHandlerInstance only handles objects of type container")
	}
	return (&DefaultHandlerInstance{}).GetObjectDefinition("container")
}

// ChildWithID returns child reference to child with defined object ID
func (handler *ContainerHandlerInstance) ChildWithID(id string) *ContainerChildReference {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	for _, child := range handler.children {
		if child.ObjID == id {
			return &child
		}
	}
	return nil
}

// Children returns copy of children
func (handler *ContainerHandlerInstance) Children() []ContainerChildReference {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	children := make([]ContainerChildReference, 0, len(handler.children))
	return append(children, handler.children...)
}

// SwitchActiveChild to referenced child
func (handler *ContainerHandlerInstance) SwitchActiveChild(sessionState *State, actionState *action.State, child *ContainerChildReference) error {
	handler.lock.Lock()
	defer handler.lock.Unlock()

	if sessionState.LogEntry.ShouldLogDebug() {
		defer func(current string) {
			if current != handler.ActiveID {
				sessionState.LogEntry.Logf(logger.DebugLevel, "switching container<%s> active child <%s> -> <%s>", handler.ID, current, handler.ActiveID)
			}
		}(handler.ActiveID)
	}

	if child == nil {
		// set no active child
		if handler.ActiveID != "" {
			if err := sessionState.ClearSubscribedObjects([]string{handler.ActiveID}); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}

	// Check if already active
	if handler.ActiveID == child.ObjID {
		return nil
	}

	handler.ActiveID = child.ObjID
	if child.ObjID != "" && !child.External {
		GetAndAddObjectAsync(sessionState, actionState, handler.ActiveID)
	}

	if child.External {
		sessionState.LogEntry.Log(logger.WarningLevel, "container contains external reference, external references are not supported")
	}

	return nil
}
