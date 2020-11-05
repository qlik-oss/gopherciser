package scenario

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	ContainerTabMode int
	// ContainerTabSettings switches active object in container
	ContainerTabSettings struct {
		Mode        ContainerTabMode `json:"mode" displayname:"Mode" doc-key:"containertab.mode"`
		ContainerID string           `json:"containerid" displayname:"Container ID" appstructure:"active:container" doc-key:"containertab.containerid"`
		ObjectID    string           `json:"objectid,omitempty" appstructure:"children:containerid" displayname:"Object ID" doc-key:"containertab.objectid"`
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
		if settings.Index != 0 {
			return errors.Errorf("index<%d> will not be used with mode<%s>", settings.Index, settings.Mode)
		}
	case ContainerTabModeRandom:
		if settings.ObjectID != "" {
			return errors.Errorf("object ID<%s> will not be used with mode<%s>", settings.ObjectID, settings.Mode)
		}
		if settings.Index != 0 {
			return errors.Errorf("index<%d> will not be used with mode<%s>", settings.Index, settings.Mode)
		}
	case ContainerTabModeIndex:
		if settings.Index < 0 {
			return errors.Errorf("index<%d> not valid", settings.Index)
		}
		if settings.ObjectID != "" {
			return errors.Errorf("object ID<%s> will not be used with mode<%s>", settings.ObjectID, settings.Mode)
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

	childCount := len(containerInstance.Children)

	var newActive *session.ContainerChildReference
	switch settings.Mode {
	case ContainerTabModeObjectID:
		newActive = containerInstance.ChildWithID(settings.ObjectID)
	case ContainerTabModeRandom:
		if childCount < 1 {
			actionState.AddErrors(errors.Errorf("switch to random container tab defined, but container<%s> has no children", id))
			return
		}

		visibleChildren := make([]*session.ContainerChildReference, 0, childCount)
		for _, child := range containerInstance.Children {
			if child.Show {
				visibleChildren = append(visibleChildren, &child)
			}
		}

		visibleChildCount := len(visibleChildren)
		if visibleChildCount < 1 {
			sessionState.LogEntry.Logf(logger.WarningLevel, "switch to random container tab defined, but container<%s> has no visible children", id)
			sessionState.Wait(actionState)
			return
		}

		idx := sessionState.Randomizer().Rand(visibleChildCount)
		newActive = visibleChildren[idx]
	case ContainerTabModeIndex:
		if !(settings.Index < childCount) {
			actionState.AddErrors(errors.Errorf("container tab index<%d> defined, but container has only %d tabs", settings.Index, childCount))
			return
		}
		newActive = &containerInstance.Children[settings.Index]
	}

	if newActive == nil {
		actionState.AddErrors(errors.New("could not resolve an object ID to switch to"))
		return
	}

	if isExternal(sessionState, newActive) {
		sessionState.Wait(actionState)
		return
	}

	if !newActive.Show {
		actionState.AddErrors(errors.Errorf("container tab index<%d> defined, but visualization<%s> at index position is not visible", settings.Index, newActive.ObjID))
		return
	}

	actionState.Details = newActive.ObjID

	if err := containerInstance.SwitchActiveChild(sessionState, actionState, newActive); err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	sessionState.Wait(actionState) // Await all async requests, e.g. those triggered on changed objects
}

func isExternal(sessionState *session.State, child *session.ContainerChildReference) bool {
	if child.External {
		sessionState.LogEntry.Log(logger.WarningLevel, "active container child is an external reference, external references are not supported")
		return true
	}
	return false
}
