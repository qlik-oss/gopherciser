package scenario

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// SetScriptVarSettings action creates/sets variables value
	SetScriptVarSettings struct {
		Name  string                 `json:"name"`
		Value session.SyncedTemplate `json:"value"`
	}
)

// Validate SetScriptVarSettings action (Implements ActionSettings interface)
func (settings SetScriptVarSettings) Validate() ([]string, error) {
	if settings.Name == "" {
		return nil, errors.New("name of variable to set not defined")
	}
	if settings.Value.String() == "" {
		return nil, errors.Errorf("value of variable<%s> not set", settings.Name)
	}
	return nil, nil
}

// Execute SetScriptVarSettings action (Implements ActionSettings interface)
func (settings SetScriptVarSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	actionState.NoResults = true // We should not log any results to log as it's not a user simulation

	value, err := sessionState.ReplaceSessionVariables(&settings.Value)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to evaluate value when trying to set variable<%s>", settings.Name))
	}

	sessionState.SetVariableValue(settings.Name, value)
	sessionState.LogEntry.LogDebug(fmt.Sprintf("setting script variable<%s> to value<%v>", settings.Name, value))
}

// AppStructureAction Implements AppStructureAction interface. It returns if this action should be included
// when doing an "get app structure" from script, IsAppAction tells the scenario
// to insert a "getappstructure" action after that action using data from
// sessionState.CurrentApp. A list of Sub action to be evaluated can also be included
// AppStructureAction returns if this action should be included when getting app structure
// and any additional sub actions which should also be included
func (settings *SetScriptVarSettings) AppStructureAction() (*AppStructureInfo, []Action) {
	return &AppStructureInfo{
		IsAppAction: false,
		Include:     true,
	}, nil
}
