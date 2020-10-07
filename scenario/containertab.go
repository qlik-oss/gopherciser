package scenario

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	ContainerTabMode int
	// ContainerTabSettings switches active object in container
	ContainerTabSettings struct {
		Mode     ContainerTabMode `json:"mode" displayname:"Mode" doc-key:"containertab.mode"`
		ID       string           `json:"id" displayname:"ID" appstructure:"active:!sheet" doc-key:"containertab.id"`
		ActiveID string           `json:"activeid,omitempty" displayname:"Active ID" doc-key:"containertab.activeid"`
		Index    int              `json:"index,omitempty" displayname:"Index" doc-key:"containertab.index"`
	}
)

// ContainerTabMode enum
const (
	ContainerTabModeID ContainerTabMode = iota
	ContainerTabModeRandom
	ContainerTabModeIndex
)

var (
	containerTabMode = enummap.NewEnumMapOrPanic(map[string]int{
		"id":     int(ContainerTabModeID),
		"random": int(ContainerTabModeRandom),
		"index":  int(ContainerTabModeIndex),
	})
)

func (mode ContainerTabMode) GetEnumMap() *enummap.EnumMap {
	return containerTabMode
}

// UnmarshalJSON unmarshal container tab mode
func (mode *ContainerTabMode) UnmarshalJSON(arg []byte) error {
	i, err := containerTabMode.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal ContainerTabMode")
	}

	*mode = ContainerTabMode(i)
	return nil
}

// MarshalJSON marshal container tab mode
func (mode ContainerTabMode) MarshalJSON() ([]byte, error) {
	str, err := containerTabMode.String(int(mode))
	if err != nil {
		return nil, errors.Errorf("Unknown ContainerTabMode<%d>", mode)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// String representation of ContainerTabMode
func (mode ContainerTabMode) String() string {
	cMode, err := containerTabMode.String(int(mode))
	if err != nil {
		return strconv.Itoa(int(mode))
	}
	return cMode
}

// Validate ContainerTabSettings action (Implements ActionSettings interface)
func (settings ContainerTabSettings) Validate() error {
	if settings.ID == "" {
		return errors.New("no container id defined")
	}

	switch settings.Mode {
	case ContainerTabModeID:
		if settings.ActiveID == "" {
			return errors.Errorf("no container activeid set for container tab mode<%s>", settings.Mode)
		}
	case ContainerTabModeRandom:
	case ContainerTabModeIndex:
		if settings.Index < 0 {
			return errors.Errorf("index<%d> not valid", settings.Index)
		}
	default:
		return errors.Errorf("unknown container tab mode<%v>", settings.Mode)
	}

	return nil
}

// Execute ContainerTabSettings action (Implements ActionSettings interface)
func (settings ContainerTabSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	id := sessionState.IDMap.Get(settings.ID)

	instance := sessionState.GetObjectHandlerInstance(id, "container")
	if instance == nil {
		actionState.AddErrors(errors.New("got nil instance for type container"))
		return
	}

	containerInstance, ok := instance.(*session.ContainerHandlerInstance)
	if !ok {
		actionState.AddErrors(errors.Errorf("object<%s> return object handler instance type<%T> expected<*session.ContainerHandlerInstance>", id, instance))
		return
	}

	activeID := ""
	switch settings.Mode {
	case ContainerTabModeID:
		activeID = settings.ActiveID
	case ContainerTabModeRandom:
		idx := sessionState.Randomizer().Rand(len(containerInstance.Children))
		activeID = containerInstance.Children[idx].ObjID
	case ContainerTabModeIndex:
		childCount := len(containerInstance.Children)
		if !(settings.Index < childCount) {
			actionState.AddErrors(errors.Errorf("container tab index<%d> defined, but container has only %d tabs", settings.Index, childCount))
			return
		}
		activeID = containerInstance.Children[settings.Index].ObjID
	}

	if err := containerInstance.SwitchActiveID(sessionState, actionState, activeID); err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	sessionState.Wait(actionState) // Await all async requests, e.g. those triggered on changed objects
}

// AppStructureAction Implements AppStructureAction interface. It returns if this action should be included
// when doing an "get app structure" from script, IsAppAction tells the scenario
// to insert a "getappstructure" action after that action using data from
// sessionState.CurrentApp. A list of Sub action to be evaluated can also be included
// AppStructureAction returns if this action should be included when getting app structure
// and any additional sub actions which should also be included
//func (settings *ContainerTabSettings) AppStructureAction() (*scenario.AppStructureInfo, []scenario.Action) {
//	return &scenario.AppStructureInfo{
//		IsAppAction: false,
//		Include:     true,
//	}, nil
//}

// AffectsAppObjectsAction implements AffectsAppObjectsAction interface
// Should be implemented by all actions that affect the availability of selectable objects for app structure consumption.
// App structure of the current app is passed as an argument. The return is
// * added *config.AppStructurePopulatedObjects - objects to be added to the selectable list by this action
// * removed []string - ids of objects that are removed (including any children) by this action
// * clearObjects bool - clears all objects except bookmarks and sheets
//func (settings ContainerTabSettings) AffectsAppObjectsAction(structure appstructure.AppStructure) ([]*appstructure.AppStructurePopulatedObjects, []string, bool) {
//	selectables, err := structure.GetSelectables(settings.ID)
//	if err != nil {
//		return nil, nil, false
//	}
//	newObjs := appstructure.AppStructurePopulatedObjects{
//		Parent:    settings.ID,
//		Objects:   make([]appstructure.AppStructureObject, 0),
//		Bookmarks: nil,
//	}
//	newObjs.Objects = append(newObjs.Objects, selectables...)
//	return []*appstructure.AppStructurePopulatedObjects{&newObjs}, nil, true
//}
