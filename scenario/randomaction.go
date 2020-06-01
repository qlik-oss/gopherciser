package scenario

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjdef"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//ActionType type of action
	ActionType int

	// ActionTypeSettings settings for specific ActionType
	ActionTypeSettings struct {
		// Type of action
		Type ActionType `json:"type" doc-key:"randomaction.actions.type"`
		// Uniform likelihood for action (weight)
		Weight int `json:"weight" doc-key:"randomaction.actions.weight"`
		// Overrides override arbitrary properties of the settings corresponding to this action
		Overrides map[string]interface{} `json:"overrides,omitempty" doc-key:"randomaction.actions.overrides"`

		itemSettings interface{}
	}
)

type (
	// RandomActionSettingsCore RandomAction settings
	RandomActionSettingsCore struct {
		// List of the different actions and their weights
		ActionTypes []ActionTypeSettings `json:"actions" displayname:"Actions" doc-key:"randomaction.actions"`
		// ThinkTime in between random actions
		InterThinkTimeSettings *ThinkTimeSettings `json:"thinktimesettings,omitempty" doc-key:"randomaction.thinktimesettings"`
		// Number of random actions to execute
		Iterations int `json:"iterations" displayname:"Iterations" doc-key:"randomaction.iterations"`
	}

	// RandomActionSettings RandomAction settings
	RandomActionSettings struct {
		RandomActionSettingsCore
		interthinktime *Action
		initialize     *sync.Once
	}
)

const (
	// ThinkTime wait for a random duration
	ThinkTime ActionType = iota
	// SheetObjectSelection selecting in a random object on the current sheet
	SheetObjectSelection
	// ChangeSheet changing to a random sheet
	ChangeSheet
	// ClearAll clearing all selections
	ClearAll
)

var (
	actionTypeEnumMap, _ = enummap.NewEnumMap(map[string]int{
		"thinktime":            int(ThinkTime),
		"sheetobjectselection": int(SheetObjectSelection),
		"changesheet":          int(ChangeSheet),
		"clearall":             int(ClearAll),
	})
)

func (value ActionType) GetEnumMap() *enummap.EnumMap {
	return actionTypeEnumMap
}

// UnmarshalJSON unmarshal ActionType
func (value *ActionType) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal ActionType")
	}

	*value = ActionType(i)
	return nil
}

// MarshalJSON marshal ActionType
func (value ActionType) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown ActionType<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// UnmarshalJSON unmarshal ActionType
func (settings *RandomActionSettings) UnmarshalJSON(arg []byte) error {
	core := RandomActionSettingsCore{}
	err := jsonit.Unmarshal(arg, &core)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal ElasticHubSearchSettingsCore")
	}
	settings.RandomActionSettingsCore = core

	if settings.ActionTypes == nil {
		settings.ActionTypes = []ActionTypeSettings{
			{Type: ChangeSheet, Weight: 1},
			{Type: SheetObjectSelection, Weight: 1},
			{Type: ClearAll, Weight: 1},
		}
	}
	settings.initialize = &sync.Once{}
	settings.interthinktime = &Action{} // Initialize pointer which will be reused when setting the actual Action
	return nil
}

func getDefaultInterThinkTimeSettings() ThinkTimeSettings {
	defaultMean := float64(35)

	return ThinkTimeSettings{
		helpers.DistributionSettings{
			Type:      helpers.UniformDistribution,
			Delay:     0,
			Mean:      defaultMean,
			Deviation: 25,
		},
	}
}

// Execute random action/-s. Implements ActionSetting interface
func (settings RandomActionSettings) Execute(sessionState *session.State, state *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	uplink := sessionState.Connection.Sense()

	settings.initialize.Do(func() {
		// Go to already initialized pointer address and set a new Action value there
		if settings.InterThinkTimeSettings != nil {
			*(settings.interthinktime) = Action{ActionCore{ActionThinkTime, fmt.Sprintf("%s - inter thinktime", label), false}, settings.InterThinkTimeSettings}
		}
	})

	for iteration := 0; iteration < settings.Iterations; iteration++ {
		weights := make([]int, 0, len(settings.ActionTypes))
		for _, actionTypeSettings := range settings.ActionTypes {
			weights = append(weights, actionTypeSettings.Weight)
		}
		randomSelection, errSelection := sessionState.Randomizer().RandWeightedInt(weights)
		if errSelection != nil {
			state.AddErrors(errors.Wrapf(errSelection, "Failed to randomize action type"))
			return
		}
		selectedAction := settings.ActionTypes[randomSelection]

		var item Action
		switch selectedAction.Type {
		case ThinkTime:
			if selectedAction.itemSettings == nil {
				var err error
				defaultSettings := settings.InterThinkTimeSettings
				if defaultSettings == nil {
					tmp := getDefaultInterThinkTimeSettings()
					defaultSettings = &tmp
				}
				selectedAction.itemSettings, err = overrideSettings(*defaultSettings, selectedAction.Overrides)
				if err != nil {
					state.AddErrors(errors.WithStack(err))
					return
				}
			}
			item = Action{ActionCore{ActionThinkTime, fmt.Sprintf("%s - generated thinktime", label), false}, selectedAction.itemSettings.(ThinkTimeSettings)}
		case SheetObjectSelection:
			// Get selectable objects
			selectableObjectsOnSheet := getSelectableObjectsOnSheet(sessionState)
			n := len(selectableObjectsOnSheet)
			if n < 1 {
				sessionState.LogEntry.LogInfo("nosheetobjects", "Cannot select sheet object - no sheet objects")
				continue
			}

			// Randomize a selectable object
			i := sessionState.Randomizer().Rand(n)
			chosenObject := selectableObjectsOnSheet[i]

			if selectedAction.itemSettings == nil {
				// Apply override settings and perform the action
				selectedAction.itemSettings = SelectionSettings{ID: chosenObject.ID, Type: RandomFromAll, Accept: true, WrapSelections: true, Min: 1, Max: 1}
				var err error
				selectedAction.itemSettings, err = overrideSettings(selectedAction.itemSettings.(ActionSettings), selectedAction.Overrides)
				if err != nil {
					state.AddErrors(errors.WithStack(err))
					return
				}
			}
			item = Action{ActionCore{ActionSelect, fmt.Sprintf("%s - generated select", label), false}, selectedAction.itemSettings.(SelectionSettings)}
		case ChangeSheet:
			// Get the sheetlist and select a random sheet
			sheetList, errSheetList := uplink.CurrentApp.GetSheetList(sessionState, state)
			if errSheetList != nil {
				state.AddErrors(errors.WithStack(errSheetList))
				return
			}
			items := sheetList.Layout().AppObjectList.Items
			n := len(items)
			if n < 1 {
				sessionState.LogEntry.LogInfo("nosheets", "Cannot change sheets - no sheets in app")
				continue
			}
			i := sessionState.Randomizer().Rand(n)
			id := items[i].Info.Id

			// Execute the action
			itemSettings := ChangeSheetSettings{ID: id}
			item = Action{ActionCore{ActionChangeSheet, fmt.Sprintf("%s - generated changesheet", label), false}, itemSettings}
		case ClearAll:
			item = Action{ActionCore{ActionClearAll, fmt.Sprintf("%s - generated clearall", label), false}, ClearAllSettings{}}
		default:
			state.AddErrors(errors.Errorf("action type<%d> not supported", selectedAction.Type))
			return
		}

		if isAborted, err := CheckActionError(item.Execute(sessionState, connectionSettings)); isAborted {
			return // action is aborted, we should not continue
		} else if err != nil {
			state.AddErrors(errors.WithStack(err))
			return
		}

		if settings.InterThinkTimeSettings == nil {
			continue
		}
		if isAborted, err := CheckActionError(settings.interthinktime.Execute(sessionState, connectionSettings)); isAborted {
			return // action is aborted, we should not continue
		} else if err != nil {
			state.AddErrors(errors.WithStack(err))
			return
		}
	}
}

// IsContainerAction implements ContainerAction interface
// and sets container action logging to original action entry
func (settings RandomActionSettings) IsContainerAction() {}

func getSelectableObjectsOnSheet(sessionState *session.State) []*enigmahandlers.Object {
	uplink := sessionState.Connection.Sense()
	// Get all objects on sheet
	handles := uplink.Objects.GetAllObjectHandles(true, enigmahandlers.ObjTypeGenericObject)
	n := len(handles)
	if n < 1 {
		sessionState.LogEntry.Log(logger.InfoLevel, "Nothing to select - no sheet objects in scope")
		return make([]*enigmahandlers.Object, 0)
	}
	// Determine what objects are selectable
	selectableObjects := make([]*enigmahandlers.Object, 0, n)
	for _, handle := range handles {
		obj, err := uplink.Objects.GetObject(handle)
		if err != nil {
			continue
		}
		enigmaObject, ok := obj.EnigmaObject.(*enigma.GenericObject)
		if !ok {
			continue
		}
		objectDef, err := senseobjdef.GetObjectDef(enigmaObject.GenericType)
		if err != nil {
			continue
		}
		if objectDef.Select != nil && objectDef.Select.Type != senseobjdef.SelectTypeUnknown {
			selectableObjects = append(selectableObjects, obj)
		}
	}
	return selectableObjects
}

func overrideSettings(originalSettings ActionSettings, overrideSettings map[string]interface{}) (interface{}, error) {
	// Marshal settings into json []byte
	originalSettingsJSON, err := jsonit.Marshal(originalSettings)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Unmarshal to generic struct
	var anyJSON map[string]interface{}
	if err := jsonit.Unmarshal(originalSettingsJSON, &anyJSON); err != nil {
		return nil, errors.WithStack(err)
	}

	// Replace matching keys
	for k, v := range overrideSettings {
		// Todo currently a check key for lower case as it's the most common, but if possible should be
		// checked case insensitive towards json tag of struct
		anyJSON[strings.ToLower(k)] = v
	}

	// Marshal it back into json []byte
	finalJSON, err := jsonit.Marshal(anyJSON)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Instantiate a new struct of the original type and unmarshal the final object
	switch originalSettings.(type) {
	case SelectionSettings:
		newSettingsObject := SelectionSettings{}
		err = jsonit.Unmarshal(finalJSON, &newSettingsObject)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return newSettingsObject, nil
	case ChangeSheetSettings:
		newSettingsObject := ChangeSheetSettings{}
		err = jsonit.Unmarshal(finalJSON, &newSettingsObject)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return newSettingsObject, nil
	case ThinkTimeSettings:
		newSettingsObject := ThinkTimeSettings{}
		err = jsonit.Unmarshal(finalJSON, &newSettingsObject)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return newSettingsObject, nil
	default:
		return nil, errors.Errorf("Type <%s> not supported by RandomAction", reflect.TypeOf(originalSettings).String())
	}
}

// Validate random action settings. Implements ActionSetting interface
func (settings RandomActionSettings) Validate() error {
	for _, actionTypeSettings := range settings.ActionTypes {
		probability := actionTypeSettings.Weight
		if probability <= 0 {
			return errors.Errorf("Action weight (p=%d) should be at least 1", probability)
		}
	}
	totalProbability := 0
	for _, actionTypeSettings := range settings.ActionTypes {
		probability := actionTypeSettings.Weight
		totalProbability += probability
		if totalProbability < 0 {
			return errors.Errorf("Summing the weights caused integer overflow!")
		}
	}
	return nil
}
