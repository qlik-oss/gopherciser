package scenario

import (
	"testing"

	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	customActionSettings struct {
		CustomAttrib string
	}
	customActionSettings2 struct{}
	selectOverride        struct{}
)

func (c customActionSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
}
func (c customActionSettings) Validate() ([]string, error) {
	return nil, nil
}
func (c customActionSettings2) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
}
func (c customActionSettings2) Validate() ([]string, error) {
	return nil, nil
}
func (s selectOverride) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
}
func (s selectOverride) Validate() ([]string, error) {
	return nil, nil
}

func TestActionhandler(t *testing.T) {
	// Make sure to reset overrides after test
	defer ResetDefaultActions()

	// Standard action
	rawSelect := `{
		"action" : "select",
		"label" : "select action",
		"settings" : {
			"id": "xyz"
		}
	}`

	var act Action
	if err := json.Unmarshal([]byte(rawSelect), &act); err != nil {
		t.Fatal("Unmarshal select action failed, err:", err)
	}

	validateString(t, "action", act.Type, "select")
	validateString(t, "label", act.Label, "select action")

	settings, ok := act.Settings.(*SelectionSettings)
	if !ok {
		t.Fatalf("select settings<%T> not of type SelectionSettings", settings)
	}
	validateString(t, "id", settings.ID, "xyz")

	// Custom action
	rawCustom := `{
		"action" : "customaction",
		"label" : "custom action",
		"settings" : {
			"customAttrib" : "myAttrib"
		}
	}`

	if err := json.Unmarshal([]byte(rawCustom), &act); err == nil {
		t.Fatal("Custom action not registered and did not get unmarshal error")
	}

	if err := RegisterAction("customaction", customActionSettings{}); err != nil {
		t.Fatal("Failed to register custom action")
	}
	if err := json.Unmarshal([]byte(rawCustom), &act); err != nil {
		t.Fatal("Unmarshal customAction failed, err:", err)
	}

	validateString(t, "action", act.Type, "customaction")
	validateString(t, "label", act.Label, "custom action")

	customsettings, ok := act.Settings.(*customActionSettings)
	if !ok {
		t.Fatal("custom action settings not of type customActionSettings")
	}
	validateString(t, "customAttrib", customsettings.CustomAttrib, "myAttrib")

	// test overrides
	if err := RegisterActionsOverride(map[string]ActionSettings{
		"customaction": customActionSettings2{},
		"select":       selectOverride{},
	}); err != nil {
		t.Fatal("Error registering overide actions: ", err)
	}

	if err := json.Unmarshal([]byte(rawSelect), &act); err != nil {
		t.Fatal("Unmarshal select action failed, err:", err)
	}
	_, ok = act.Settings.(*selectOverride)
	if !ok {
		t.Fatal("overidden select settings not of type selectOverride")
	}

	if err := json.Unmarshal([]byte(rawCustom), &act); err != nil {
		t.Fatal("Unmarshal customAction failed, err:", err)
	}
	_, ok = act.Settings.(*customActionSettings2)
	if !ok {
		t.Fatal("overidden custom settings not of type customActionSettings2")
	}
}

func TestDisabledInvalid(t *testing.T) {
	disabledInvalid := `{
		"action" : "thinktime",
		"disabled": true,
		"settings" : {
			"type": "static"
		}
	}`

	var act Action

	if err := json.Unmarshal([]byte(disabledInvalid), &act); err != nil {
		t.Fatal("Unmarshal thinktime with invalid settings failed, err:", err)
	}

	if _, err := act.Validate(); err != nil {
		t.Fatal("Validation failed even though action is disabled, err:", err)
	}
}
