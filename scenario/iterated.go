package scenario

import (
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// IteratedSettings parallel action settings
	IteratedSettings struct {
		Iterations int      `json:"iterations" displayname:"Iterations" doc-key:"iterated.iterations"`
		Actions    []Action `json:"actions" displayname:"Actions" doc-key:"iterated.actions"`
	}
)

// Execute iterated actions
func (action IteratedSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	actionCount := 0

	for i := 0; i < action.Iterations || action.Iterations == -1; i++ {
		if sessionState.IsAbortTriggered() {
			break
		}
		for idx := range action.Actions {
			if sessionState.IsAbortTriggered() {
				break
			}
			actionCount++

			if isAborted, err := CheckActionError(action.Actions[idx].Execute(sessionState, connection)); isAborted {
				return // action is aborted, we should not continue
			} else if err != nil {
				actionState.AddErrors(errors.WithStack(err))
				return
			}
		}
		if actionState.Failed {
			break
		}
	}
}

// Validate iterated actions
func (action IteratedSettings) Validate() ([]string, error) {

	if action.Iterations < 1 && action.Iterations != -1 {
		return nil, errors.Errorf("illegal iterations count<%d>", action.Iterations)
	}

	warnings := make([]string, 0)
	// Validate all actions before executing
	for _, v := range action.Actions {
		if w, err := v.Validate(); err != nil {
			return nil, errors.WithStack(err)
		} else if len(w) > 0 {
			warnings = append(warnings, w...)
		}
	}

	return warnings, nil
}

// IsContainerAction implements ContainerAction interface
// and sets container action logging to original action entry
func (action IteratedSettings) IsContainerAction() {}

// AppStructureAction implements AppStructureAction interface
func (action IteratedSettings) AppStructureAction() (*AppStructureInfo, []Action) {
	return &AppStructureInfo{
		IsAppAction: false,
		Include:     false,
	}, action.Actions
}
