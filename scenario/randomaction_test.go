package scenario

import (
	"fmt"
	"testing"

	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/stretchr/testify/assert"
)

var (
	someInterThinkTimeSettings = ThinkTimeSettings{
		helpers.DistributionSettings{
			Type:      helpers.UniformDistribution,
			Delay:     0,
			Mean:      30,
			Deviation: 25,
		},
	}

	someActionTypes = []ActionTypeSettings{
		{Type: ChangeSheet, Weight: 2},
		{Type: SheetObjectSelection, Weight: 2},
		{Type: ClearAll, Weight: 1},
	}

	someRandomActionSettingsCore = RandomActionSettingsCore{
		ActionTypes:            someActionTypes,
		Iterations:             1,
		InterThinkTimeSettings: &someInterThinkTimeSettings,
	}

	someRandomActionSettings = &RandomActionSettings{
		someRandomActionSettingsCore,
		nil,
		nil,
	}

	someRandomAction = Action{
		ActionCore{
			Type:  ActionRandom,
			Label: "randomaction",
		},
		someRandomActionSettings,
	}
)

func TestRandomactionUnmarshal(t *testing.T) {
	t.Parallel()

	raw := `{
			"label": "randomaction",
			"action": "RandomAction",
			"settings": {
				"iterations": 5,
				"actions": [
					{
						"type": "thinktime",
						"weight": 1
					},
					{
						"type": "sheetobjectselection",
						"weight": 3
					},
					{
						"type": "changesheet",
						"weight": 5
					},
					{
						"type": "clearall",
						"weight": 1
					}
				],
				"thinktimesettings": {
					"type": "static",
					"delay": 0.1
				}
			}
		}`
	var item Action
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		t.Fatal(err)
	}

	if _, err := item.Validate(); err != nil {
		t.Error(err)
	}

	validateString(t, "action", item.Type, "randomaction")
	validateString(t, "label", item.Label, "randomaction")

	settings, ok := item.Settings.(*RandomActionSettings)
	if !ok {
		t.Fatalf("Failed to cast item settings<%T> to *RandomActionSettings", item.Settings)
	}

	validateInt(t, "iterations", settings.Iterations, 5)

}

func TestRandomactionMarshal(t *testing.T) {
	raw, err := json.Marshal(someRandomAction)
	if err != nil {
		t.Fatal(err)
	}

	validateString(t, "json", string(raw), `{"action":"randomaction","label":"randomaction","disabled":false,"settings":{"actions":[{"type":"changesheet","weight":2},{"type":"sheetobjectselection","weight":2},{"type":"clearall","weight":1}],"thinktimesettings":{"type":"uniform","mean":30,"dev":25},"iterations":1}}`)
}

func TestRandomActionValidate(t *testing.T) {
	settings := someRandomActionSettings

	_, err := settings.Validate()
	assert.NoError(t, err)

	settings.ActionTypes[0].Weight = 9001
	_, err = settings.Validate()
	assert.NoError(t, err)

	settings.ActionTypes[0].Weight = -1
	_, err = settings.Validate()
	validateError(t, err, "Action weight (p=-1) should be at least 1")

	settings.ActionTypes = nil
	_, err = settings.Validate()
	assert.NoError(t, err, "Empty ActionTypes should be allowed")
}

func TestRandomActionOverrideSettings(t *testing.T) {
	t.Parallel()

	selectedAction := ActionTypeSettings{Type: SheetObjectSelection, Weight: 2}
	origSettings := SelectionSettings{ID: "id", Type: RandomFromAll, Accept: true, WrapSelections: true, Min: 1, Max: 1}
	selectedAction.itemSettings = origSettings
	overrides := make(map[string]interface{})

	// Test if interface implemented correctly
	_, ok := selectedAction.itemSettings.(ActionSettings)
	if !ok {
		t.Fatal(fmt.Sprintf("settings<%T> not of type ActionSettings", origSettings))
	}

	// Test overide select type
	overrides["type"] = "randomfromenabled"
	settings, err := overrideSettings(origSettings, overrides)
	assert.NoError(t, err)
	if settings.(SelectionSettings).Type != RandomFromEnabled {
		t.Errorf("Expected select type randomfromenabled, have %v", settings.(SelectionSettings).Type)
	}

	// Test override of integer
	overrides["max"] = 2
	settings, err = overrideSettings(origSettings, overrides)
	assert.NoError(t, err)
	if settings.(SelectionSettings).Max != 2 {
		t.Errorf("Expected max selects of 2, have %d", settings.(SelectionSettings).Max)
	}

	// Test override of integer with case insensitive
	overrides = make(map[string]interface{})
	overrides["Max"] = 3
	settings, err = overrideSettings(origSettings, overrides)
	assert.NoError(t, err)
	if settings.(SelectionSettings).Max != 3 {
		t.Errorf("Expected max selects of 3, have %d", settings.(SelectionSettings).Max)
	}
}
