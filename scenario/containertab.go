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
		Mode        ContainerTabMode `json:"mode" displayname:"Mode" doc-key:"containertab.mode"`
		ContainerID string           `json:"containerid" displayname:"ID" appstructure:"active:container" doc-key:"containertab.containerid"`
		ObjectID    string           `json:"objectid,omitempty" appstructure:"children:id" displayname:"Active ID" doc-key:"containertab.objectid"`
		Index       int              `json:"index,omitempty" displayname:"Index" doc-key:"containertab.index"`
	}
)

// ContainerTabMode enum
const (
	ContainerTabModeObjectID ContainerTabMode = iota
	ContainerTabModeRandom
	ContainerTabModeIndex
)

var (
	containerTabMode = enummap.NewEnumMapOrPanic(map[string]int{
		"objectid": int(ContainerTabModeObjectID),
		"random":   int(ContainerTabModeRandom),
		"index":    int(ContainerTabModeIndex),
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
	if settings.ContainerID == "" {
		return errors.New("no container id defined")
	}

	switch settings.Mode {
	case ContainerTabModeObjectID:
		if settings.ObjectID == "" {
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
	id := sessionState.IDMap.Get(settings.ContainerID)

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
	case ContainerTabModeObjectID:
		activeID = settings.ObjectID
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
