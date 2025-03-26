package scenario

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjdef"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//ActionType type of action
	ActionType int

	// ActionTypeSettings settings for specific ActionType
	ActionTypeSettings struct {
		// Type of action
		Type ActionType `json:"type" doc-key:"randomaction.actions.type" displayname:"Type"`
		// Uniform likelihood for action (weight)
		Weight int `json:"weight" doc-key:"randomaction.actions.weight" displayname:"weight"`
		// Overrides override arbitrary properties of the settings corresponding to this action
		Overrides map[string]any `json:"overrides,omitempty" doc-key:"randomaction.actions.overrides" displayname:"Overrides"`

		itemSettings any
	}
)

type (
	// RandomActionSettingsCore RandomAction settings
	RandomActionSettingsCore struct {
		// List of the different actions and their weights
		ActionTypes []ActionTypeSettings `json:"actions" displayname:"Actions" doc-key:"randomaction.actions"`
		// ThinkTime in between random actions
		InterThinkTimeSettings *ThinkTimeSettings `json:"thinktimesettings,omitempty" doc-key:"randomaction.thinktimesettings" displayname:"Think time inbetween actions"`
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
	actionTypeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
		"thinktime":            int(ThinkTime),
		"sheetobjectselection": int(SheetObjectSelection),
		"changesheet":          int(ChangeSheet),
		"clearall":             int(ClearAll),
	})

	defaultThinkTimeSettings = ThinkTimeSettings{
		helpers.DistributionSettings{
			Type:      helpers.UniformDistribution,
			Delay:     0,
			Mean:      float64(35),
			Deviation: 25,
		},
	}
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

func (value ActionType) String() string {
	return actionTypeEnumMap.StringDefault(int(value), "unknown")
}

// UnmarshalJSON unmarshal ActionType
func (settings *RandomActionSettings) UnmarshalJSON(arg []byte) error {
	core := RandomActionSettingsCore{}
	err := json.Unmarshal(arg, &core)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal RandomActionSettings")
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

// Execute random action/-s. Implements ActionSetting interface
func (settings RandomActionSettings) Execute(sessionState *session.State, state *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	settings.initialize.Do(func() {
		// Go to already initialized pointer address and set a new Action value there
		if settings.InterThinkTimeSettings != nil {
			*(settings.interthinktime) = Action{ActionCore{ActionThinkTime, fmt.Sprintf("%s - inter thinktime", label), false}, settings.InterThinkTimeSettings}
		}
	})

	// logEntry for randomaction
	logEntry := sessionState.LogEntry

	// Get action weights
	weights := make([]int, 0, len(settings.ActionTypes))
	for _, actionTypeSettings := range settings.ActionTypes {
		weights = append(weights, actionTypeSettings.Weight)
	}

	for iteration := 0; iteration < settings.Iterations; iteration++ {
		sessionState.LogEntry = logEntry // reset current log entry to container action

		randomSelection, errSelection := sessionState.Randomizer().RandWeightedInt(weights)
		if errSelection != nil {
			state.AddErrors(errors.Wrapf(errSelection, "Failed to randomize action type"))
			return
		}
		selectedAction := settings.ActionTypes[randomSelection]
		state.Details = selectedAction.Type.String() // set currently selected action type to details for traceability

		sessionState.LogEntry.LogDebugf("randomaction: selected sub action: %s", selectedAction.Type)

		// Await any ongoing reconnection
		sessionState.AwaitReconnect()

		var item Action
		switch selectedAction.Type {
		case ThinkTime:
			if selectedAction.itemSettings == nil {
				thinkTimeSettings := settings.InterThinkTimeSettings
				if thinkTimeSettings == nil {
					thinkTimeSettings = &defaultThinkTimeSettings
				}

				var err error
				selectedAction.itemSettings, err = overrideSettings(*thinkTimeSettings, selectedAction.Overrides)
				if err != nil {
					state.AddErrors(errors.WithStack(err))
					return
				}
			}
			item = Action{ActionCore{ActionThinkTime, fmt.Sprintf("%s - generated thinktime", label), false}, selectedAction.itemSettings.(ThinkTimeSettings)}
		case SheetObjectSelection:
			// Get selectable objects
			selectableObjectsOnSheet, err := getSelectableObjectsOnSheet(sessionState)
			if err != nil {
				state.AddErrors(err)
				return
			}

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
			currentApp, err := sessionState.CurrentSenseApp()
			if err != nil {
				state.AddErrors(err)
				return
			}

			sheetList, errSheetList := currentApp.GetSheetList(sessionState, state)
			if errSheetList != nil {
				state.AddErrors(errors.WithStack(errSheetList))
				return
			}

			allItems := sheetList.Layout().AppObjectList.Items
			items := make([]*senseobjects.SheetNxContainerEntry, 0, len(allItems))
			for _, item := range allItems { // Only randomize between non-hidden sheets
				if item != nil && item.Data != nil && item.Data.ShowCondition {
					items = append(items, item)
				}
			}

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

		sessionState.LogEntry.LogDebugf("randomaction: executing sub action: %s", item.Type)
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

func getSelectableObjectsOnSheet(sessionState *session.State) ([]*enigmahandlers.Object, error) {
	uplink, err := sessionState.CurrentSenseUplink()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Get all objects on sheet
	handles := uplink.Objects.GetAllObjectHandles(true, enigmahandlers.ObjTypeGenericObject)
	n := len(handles)
	if n < 1 {
		sessionState.LogEntry.Log(logger.InfoLevel, "Nothing to select - no sheet objects in scope")
		return make([]*enigmahandlers.Object, 0), nil
	}
	// Determine what objects are selectable
	selectableObjects := make([]*enigmahandlers.Object, 0, n)
	var unselectableObjects []string
	if sessionState.LogEntry.ShouldLogDebug() {
		unselectableObjects = make([]string, 0, n)
	}
	for _, handle := range handles {
		obj, err := uplink.Objects.GetObject(handle)
		if err != nil {
			continue
		}
		enigmaObject, ok := obj.EnigmaObject.(*enigma.GenericObject)
		if !ok {
			if sessionState.LogEntry.ShouldLogDebug() {
				unselectableObjects = append(unselectableObjects, enigmaObject.GenericId)
			}
			continue
		}

		objInstance := sessionState.GetObjectHandlerInstance(enigmaObject.GenericId, enigmaObject.GenericType)
		_, selectType, _, err := objInstance.GetObjectDefinition(enigmaObject.GenericType)
		if err != nil {
			unselectableObjects = append(unselectableObjects, enigmaObject.GenericId)
			continue
		}

		if selectType != senseobjdef.SelectTypeUnknown && obj.HasDims() {
			selectableObjects = append(selectableObjects, obj)
		} else {
			unselectableObjects = append(unselectableObjects, enigmaObject.GenericId)
		}
	}
	if sessionState.LogEntry.ShouldLogDebug() {
		sessionState.LogEntry.LogDebugf("unselectable objects %v", unselectableObjects)
		selectableIDs := make([]string, 0, len(selectableObjects))
		for _, obj := range selectableObjects {
			if obj != nil {
				selectableIDs = append(selectableIDs, obj.ID)
			}
		}
		sessionState.LogEntry.LogDebugf("selectable objects %v", selectableIDs)
	}
	return selectableObjects, nil
}

func overrideSettings(originalSettings ActionSettings, overrideSettings map[string]interface{}) (interface{}, error) {
	// Marshal settings into json []byte
	originalSettingsJSON, err := json.Marshal(originalSettings)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Unmarshal to generic struct
	var anyJSON map[string]interface{}
	if err := json.Unmarshal(originalSettingsJSON, &anyJSON); err != nil {
		return nil, errors.WithStack(err)
	}

	// Replace matching keys
	for k, v := range overrideSettings {
		// Todo currently a check key for lower case as it's the most common, but if possible should be
		// checked case insensitive towards json tag of struct
		anyJSON[strings.ToLower(k)] = v
	}

	// Marshal it back into json []byte
	finalJSON, err := json.Marshal(anyJSON)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Instantiate a new struct of the original type and unmarshal the final object
	switch originalSettings.(type) {
	case SelectionSettings:
		newSettingsObject := SelectionSettings{}
		err = json.Unmarshal(finalJSON, &newSettingsObject)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return newSettingsObject, nil
	case ChangeSheetSettings:
		newSettingsObject := ChangeSheetSettings{}
		err = json.Unmarshal(finalJSON, &newSettingsObject)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return newSettingsObject, nil
	case ThinkTimeSettings:
		newSettingsObject := ThinkTimeSettings{}
		err = json.Unmarshal(finalJSON, &newSettingsObject)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return newSettingsObject, nil
	default:
		return nil, errors.Errorf("Type <%s> not supported by RandomAction", reflect.TypeOf(originalSettings).String())
	}
}

// Validate random action settings. Implements ActionSetting interface
func (settings RandomActionSettings) Validate() ([]string, error) {
	for _, actionTypeSettings := range settings.ActionTypes {
		probability := actionTypeSettings.Weight
		if probability <= 0 {
			return nil, errors.Errorf("Action weight (p=%d) should be at least 1", probability)
		}
	}
	totalProbability := 0
	for _, actionTypeSettings := range settings.ActionTypes {
		probability := actionTypeSettings.Weight
		totalProbability += probability
		if totalProbability < 0 {
			return nil, errors.Errorf("Summing the weights caused integer overflow!")
		}
	}
	return nil, nil
}
