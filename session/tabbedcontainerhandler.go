package session

import (
	"context"
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
	TabbedContainerHandler struct{}

	TabbedContainerHandlerInstance struct {
		ID        string
		ActiveID  string
		DefaultID string
		CachedID  string

		children []ContainerChildReference // TODO validate container with master objects
		mu       sync.Mutex
	}

	TabbedContainerLayoutObjects struct {
		ChildRefId string            `json:"childRefId"`
		Label      string            `json:"label"`
		Condition  helpers.FuzzyBool `json:"condition"`
	}

	TabbedContainerLayout struct {
		Objects      []TabbedContainerLayoutObjects `json:"objects"`
		ChildList    ContainerChildList             `json:"qChildList"`
		DefaultTabId string                         `json:"defaultTabId"`
		ShowTabs     bool                           `json:"showTabs"`
		CachedTabId  string                         `json:"cachedTabId"`
	}
)

// Instance implements ObjectHandler  interface
func (handler *TabbedContainerHandler) Instance(id string) ObjectHandlerInstance {
	return &TabbedContainerHandlerInstance{ID: id}
}

// GetObjectDefinition implements ObjectHandlerInstance interface
func (handler *TabbedContainerHandlerInstance) GetObjectDefinition(objectType string) (string, senseobjdef.SelectType, senseobjdef.DataDefType, error) {
	if objectType != "sn-tabbed-container" {
		return "", senseobjdef.SelectTypeUnknown, senseobjdef.DataDefUnknown, errors.New("TabbedContainerHandlerInstance only handles objects of type sn-tabbed-container")
	}
	return (&DefaultHandlerInstance{}).GetObjectDefinition("sn-tabbed-container")
}

// SetObjectAndEvents implements ObjectHandlerInstance interface
func (handler *TabbedContainerHandlerInstance) SetObjectAndEvents(sessionState *State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject) {

	sessionState.QueueRequest(func(ctx context.Context) error {
		return GetObjectProperties(sessionState, actionState, obj)
	}, actionState, true, "")

	layout := GetTabbedContainerLayout(sessionState, actionState, genObj)
	if layout == nil {
		return // error occured and has been reported on actionState
	}

	handler.DefaultID = layout.DefaultTabId
	handler.CachedID = layout.CachedTabId

	if err := handler.UpdateChildren(layout); err != nil {
		actionState.AddErrors(err)
		return
	}

	sessionState.RegisterEvent(genObj.Handle, func(ctx context.Context, as *action.State) error {
		sessionState.LogEntry.Logf(logger.DebugLevel, "Getting layout for object<%s> handle<%d> type<%s>", genObj.GenericId, genObj.Handle, genObj.GenericType)
		layout := GetTabbedContainerLayout(sessionState, as, genObj)
		if as.Failed {
			return nil
		}

		if err := handler.UpdateChildren(layout); err != nil {
			return errors.Wrapf(err, "failed to update children for tabbed container object<%s>", genObj.GenericId)
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
		handler.SwitchActiveChild(sessionState, as, child)
		return nil
	}, nil, true)

	child, _ := handler.FirstShowableChild()
	handler.SwitchActiveChild(sessionState, actionState, child)
}

func (handler *TabbedContainerHandlerInstance) UpdateChildren(layout *TabbedContainerLayout) error {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	refMap, err := createContainerChildRefMap(&layout.ChildList)
	if err != nil {
		return errors.WithStack(err)
	}

	// Create child array with same order as .Objects, this is the order of the tabs
	handler.children = make([]ContainerChildReference, 0, len(layout.Objects))
	for _, child := range layout.Objects {
		ccr := ContainerChildReference{RefID: child.ChildRefId, ObjID: refMap[child.ChildRefId]}
		// TODO validate master objects
		// if child.ExternalReference != nil {
		// 	ccr.External = true
		// }
		// if child.Condition == nil {
		// 	ccr.Show = true
		// } else {
		// 	ccr.Show = bool(*child.Condition)
		// }
		ccr.Show = bool(child.Condition)
		handler.children = append(handler.children, ccr)
	}

	return nil
}

// FirstShowableChild gets reference to first object from container which fulfills conditions, true if already active object
func (handler *TabbedContainerHandlerInstance) FirstShowableChild() (*ContainerChildReference, bool) {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	showMap := make(map[string]*ContainerChildReference, len(handler.children))

	var firstShow *ContainerChildReference
	showCntr := 0
	for _, child := range handler.children {
		showMap[child.ObjID] = &child
		if child.Show {
			showCntr++
			if firstShow == nil {
				firstShow = &child
			}
		}
	}

	if showCntr < 1 {
		// No object to show
		return nil, false
	}

	if showCntr == 1 {
		// Exactly one object to show
		return firstShow, firstShow.ObjID == handler.ActiveID
	}

	// check active ID
	if child := showMap[handler.ActiveID]; child != nil && child.Show {
		return child, true
	}

	// check cached ID
	if child := showMap[handler.CachedID]; child != nil && child.Show {
		return child, false
	}

	// Check default ID
	if child := showMap[handler.DefaultID]; child != nil && child.Show {
		return child, false
	}

	// Return first showable in array order
	return firstShow, false
}

func (handler *TabbedContainerHandlerInstance) SwitchActiveChild(sessionState *State, actionState *action.State, child *ContainerChildReference) {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	if sessionState.LogEntry.ShouldLogDebug() {
		defer func(current string) {
			if current != handler.ActiveID {
				sessionState.LogEntry.Logf(logger.DebugLevel, "switching container<%s> active child <%s> -> <%s>", handler.ID, current, handler.ActiveID)
			}
		}(handler.ActiveID)
	}

	if child == nil {
		// no child active
		if handler.ActiveID != "" {
			if err := sessionState.ClearSubscribedObjects([]string{handler.ActiveID}); err != nil {
				actionState.AddErrors(errors.WithStack(err))
			}
			handler.ActiveID = ""
		}
		return
	}

	if handler.ActiveID == child.ObjID {
		// Already active
		return
	}

	// TODO handle master/external object

	// Subscribe to new active ID
	if child.ObjID != "" {
		GetAndAddObjectAsync(sessionState, actionState, handler.ActiveID)
	}
}

// GetTabbedContainerLayout returns unmarshaled tabbed container layout, errors reported on action state
func GetTabbedContainerLayout(sessionState *State, actionState *action.State, containerObject *enigma.GenericObject) *TabbedContainerLayout {
	rawLayout, err := sessionState.SendRequestRaw(actionState, containerObject.GetLayoutRaw)
	if err != nil {
		actionState.AddErrors(err)
		return nil
	}

	var layout TabbedContainerLayout
	if err := json.Unmarshal(rawLayout, &layout); err != nil {
		actionState.AddErrors(err)
		return nil
	}

	return &layout
}
