package session

import (
	"context"
	"sync"

	"github.com/goccy/go-json"
	"github.com/hashicorp/go-multierror"
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

		children []TabbedContainerChildReference
		mu       sync.Mutex
	}

	TabbedContainerLayoutObjects struct {
		ChildRefId string            `json:"childRefId"`
		Label      string            `json:"label"`
		Condition  helpers.FuzzyBool `json:"condition"`
	}

	TabbedContainerLayoutChildListItemsData struct {
		Title         string `json:"title"`
		Visualization string `json:"visualization"`
		ChildRefId    string `json:"childRefId"`
		ExtendsId     string `json:"qExtendsId"`
		ShowCondition string `json:"showCondition"`
	}

	TabbedContainerLayoutChildListItems struct {
		Info enigma.NxInfo                           `json:"qInfo"`
		Meta interface{}                             `json:"qMeta"`
		Data TabbedContainerLayoutChildListItemsData `json:"qData"`
	}

	TabbedContainerLayoutChildList struct {
		Items []TabbedContainerLayoutChildListItems `json:"qItems"`
	}

	TabbedContainerLayout struct {
		Objects      []TabbedContainerLayoutObjects `json:"objects"`
		ChildList    TabbedContainerLayoutChildList `json:"qChildList"`
		DefaultTabId string                         `json:"defaultTabId"`
		ShowTabs     bool                           `json:"showTabs"`
		CachedTabId  string                         `json:"cachedTabId"`
	}

	TabbedContainerChildReference struct {
		RefID string
		ObjID string
		Show  bool
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

	refMap, err := createTabbedContainerChildRefMap(&layout.ChildList)
	if err != nil {
		return errors.WithStack(err)
	}

	// Create child array with same order as .Objects, this is the order of the tabs
	handler.children = make([]TabbedContainerChildReference, 0, len(layout.Objects))
	for _, child := range layout.Objects {
		ccr := TabbedContainerChildReference{RefID: child.ChildRefId, ObjID: refMap[child.ChildRefId]}
		ccr.Show = bool(child.Condition)
		handler.children = append(handler.children, ccr)
	}

	return nil
}

// FirstShowableChild gets reference to first object from container which fulfills conditions, true if already active object
func (handler *TabbedContainerHandlerInstance) FirstShowableChild() (*TabbedContainerChildReference, bool) {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	showMap := make(map[string]*TabbedContainerChildReference, len(handler.children))

	var firstShow *TabbedContainerChildReference
	showCntr := 0
	for _, child := range handler.children {
		showMap[child.ObjID] = &child
		if child.Show {
			showCntr++
			if firstShow == nil {
				c := child
				firstShow = &c
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

func (handler *TabbedContainerHandlerInstance) SwitchActiveChild(sessionState *State, actionState *action.State, child *TabbedContainerChildReference) {
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
	handler.ActiveID = child.ObjID

	// Subscribe to new active ID
	if handler.ActiveID != "" {
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

func createTabbedContainerChildRefMap(childList *TabbedContainerLayoutChildList) (map[string]string, error) {
	if childList == nil {
		return map[string]string{}, nil
	}
	var mErr *multierror.Error
	refMap := make(map[string]string, len(childList.Items))
	for _, item := range childList.Items {
		ref := item.Data.ChildRefId
		if ref == "" {
			mErr = multierror.Append(mErr, errors.Errorf("failed to find reference for object<%s>", item.Info.Id))
			continue
		}
		refMap[ref] = item.Info.Id
		if item.Data.ExtendsId != "" { // if it's a masterobject, use masterobject ID instead
			refMap[ref] = item.Data.ExtendsId
		}
	}

	if mErr != nil {
		return nil, helpers.FlattenMultiError(mErr)
	}
	return refMap, nil

}
