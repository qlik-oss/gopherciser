package scenario

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// SetScriptVarSettings action creates/sets variables value
	SetScriptVarSettings struct {
		Name      string                          `json:"name" doc-key:"setscriptvar.name" displayname:"Name"`
		Type      session.SessionVariableTypeEnum `json:"type" doc-key:"setscriptvar.type" displayname:"Variable type"`
		Value     session.SyncedTemplate          `json:"value" doc-key:"setscriptvar.value" displayname:"Variable value"`
		Separator string                          `json:"sep" doc-key:"setscriptvar.sep" displayname:"Array separator"`
	}
)

const DefaultArraySeparator = ","

// Validate SetScriptVarSettings action (Implements ActionSettings interface)
func (settings SetScriptVarSettings) Validate() ([]string, error) {
	if settings.Name == "" {
		return nil, errors.New("name of variable to set not defined")
	}
	if settings.Value.String() == "" {
		return nil, errors.Errorf("value of variable<%s> not set", settings.Name)
	}
	if settings.Type == session.SessionVariableTypeUnknown {
		return nil, errors.New("variable type definition missing")
	}
	if settings.Type == session.SessionVariableTypeArray && settings.Separator == "" {
		return []string{fmt.Sprintf(`No array separator defined, using "%s"`, DefaultArraySeparator)}, nil
	}
	return nil, nil
}

// Execute SetScriptVarSettings action (Implements ActionSettings interface)
func (settings SetScriptVarSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	actionState.NoResults = true // We should not log any results to log as it's not a user simulation

	value, err := sessionState.ReplaceSessionVariables(&settings.Value)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to evaluate value when trying to set variable<%s>", settings.Name))
		return
	}

	switch settings.Type {
	case session.SessionVariableTypeString:
		sessionState.SetVariableValue(settings.Name, value)
	case session.SessionVariableTypeInt:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			actionState.AddErrors(errors.Errorf("failed to parse value<%s> to integer", value))
			return
		}
		sessionState.SetVariableValue(settings.Name, i)
	case session.SessionVariableTypeArray:
		separator := DefaultArraySeparator
		if settings.Separator != "" {
			separator = settings.Separator
		}
		sessionState.SetVariableValue(settings.Name, strings.Split(value, separator))
	default:
		actionState.AddErrors(errors.Errorf("session variable type<%s> not yet supported", settings.Type))
		return
	}

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
