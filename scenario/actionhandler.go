package scenario

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/buildmetrics"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// *** Interfaces which could be implemented on action settings ***

	// ActionSettings scenario action interface for mandatory methods
	ActionSettings interface {
		// Execute action
		Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string /* Label */, reset func() /* reset action start */)
		// Validate action []string are validation warnings to be reported to user
		Validate() ([]string, error)
	}

	// ContainerAction Implement this interface on action settings to mark an action as a
	// container action containing other actions. A container action will not log result as a normal action,
	// instead result will be logged as level=info, infotype: containeractionend
	// Returns if action is to be considered a container action.
	// ContainerAction can't be used in conjunction with StartActionOverrider interface
	ContainerAction interface {
		IsContainerAction()
	}

	// AppStructureAction returns if this action should be included
	// when doing an "get app structure" from script, IsAppAction tells the scenario
	// to insert a "getappstructure" action after that action using data from
	// sessionState.CurrentApp. A list of Sub action to be evaluated can also be included
	AppStructureAction interface {
		AppStructureAction() (*AppStructureInfo, []Action)
	}

	// AffectsAppObjectsAction is an interface that should be implemented by all actions that affect
	// the availability of selectable objects for app structure consumption. App structure of the current
	// app is passed as an argument. The return is
	// * added *config.AppStructurePopulatedObjects - objects to be added to the selectable list by this action
	// * removed []string - ids of objects that are removed (including any children) by this action
	// * clearObjects bool - clears all objects except bookmarks and sheets
	AffectsAppObjectsAction interface {
		AffectsAppObjectsAction(appstructure.AppStructure) ([]*appstructure.AppStructurePopulatedObjects, []string, bool)
	}

	DefaultValuesForGUI interface {
		DefaultValuesForGUI() ActionSettings
	}

	// ****************************************************************

	ActionCore struct {
		Type     string `json:"action" doc-key:"config.scenario.action"`
		Label    string `json:"label" doc-key:"config.scenario.label"`
		Disabled bool   `json:"disabled" doc-key:"config.scenario.disabled"`
	}

	actionTemp struct {
		ActionCore
		Settings json.RawMessage
	}

	// Action simulated user action
	Action struct {
		ActionCore
		Settings ActionSettings `json:"settings,omitempty" doc-key:"config.scenario.settings"`
	}

	// AbortedError action was aborted
	AbortedError struct{}

	// AppStructureActionContainer
	AppStructureInfo struct {
		IsAppAction bool
		Include     bool
	}
)

type (
	// *** Interfaces which could be implemented on action struct field types ***

	// Enum interface should be implemented on types used fields of action struct if:
	// 1. Type is derived from one integer type. Example: `type MyType int`
	// 2. Type has natural string representations for its values.
	// Typically you should consider implementing Enum when declaring global
	// constants of a user defined integer type.
	Enum interface {
		GetEnumMap() *enummap.EnumMap
	}
)

const (
	ActionConnectWs             = "connectws"
	ActionOpenApp               = "openapp"
	ActionOpenHub               = "openhub"
	ActionGenerateOdag          = "generateodag"
	ActionDeleteOdag            = "deleteodag"
	ActionCreateSheet           = "createsheet"
	ActionCreateBookmark        = "createbookmark"
	ActionDeleteBookmark        = "deletebookmark"
	ActionApplyBookmark         = "applybookmark"
	ActionSetScript             = "setscript"
	ActionChangeSheet           = "changesheet"
	ActionSelect                = "select"
	ActionClearAll              = "clearall"
	ActionIterated              = "iterated"
	ActionThinkTime             = "thinktime"
	ActionRandom                = "randomaction"
	ActionSheetChanger          = "sheetchanger"
	ActionDuplicateSheet        = "duplicatesheet"
	ActionReload                = "reload"
	ActionProductVersion        = "productversion"
	ActionPublishSheet          = "publishsheet"
	ActionUnPublishSheet        = "unpublishsheet"
	ActionDisconnectApp         = "disconnectapp"
	ActionDeleteSheet           = "deletesheet"
	ActionPublishBookmark       = "publishbookmark"
	ActionUnPublishBookmark     = "unpublishbookmark"
	ActionSubscribeObjects      = "subscribeobjects"
	ActionUnsubscribeObjects    = "unsubscribeobjects"
	ActionListBoxSelect         = "listboxselect"
	ActionDisconnectEnvironment = "disconnectenvironment"
	ActionClickActionButton     = "clickactionbutton"
	ActionContainerTab          = "containertab"
	ActionDoSave                = "dosave"
	ActionClearField            = "clearfield"
	ActionAskHubAdvisor         = "askhubadvisor"
	ActionSetSenseVariable      = "setsensevariable"
	ActionSetScriptVar          = "setscriptvar"
	ActionSmartSearch           = "smartsearch"
	ActionObjectSearch          = "objectsearch"
	ActionGetScript             = "getscript"
)

// Scenario actions needs an entry in actionHandler
var (
	actionHandler map[string]ActionSettings
	ahLock        sync.Mutex
)

func init() {
	ResetDefaultActions()
}

// NewActionsSettings of type
func NewActionsSettings(typ string) interface{} {
	settings := actionHandler[typ]
	if settings == nil {
		return nil
	}

	return reflect.New(reflect.TypeOf(settings)).Interface()
}

// RegisterActions register custom actions.
// This should be done as early as possible and must be done before unmarshaling actions
func RegisterActions(customActionMap map[string]ActionSettings) error {
	return errors.WithStack(registerActions(false, customActionMap))
}

// RegisterActionsOverride register custom actions and override any existing with same name
// This should be done as early as possible and must be done before unmarshaling actions
func RegisterActionsOverride(customActionMap map[string]ActionSettings) error {
	return errors.WithStack(registerActions(true, customActionMap))
}

// RegisterAction register a custom action any existing with same name
// This should be done as early as possible and must be done before unmarshaling actions
func RegisterAction(act string, settings ActionSettings) error {
	return errors.WithStack(registerAction(false, act, settings))
}

// RegisterActionOverride register a custom action and override any existing with same name
// This should be done as early as possible and must be done before unmarshaling actions
func RegisterActionOverride(act string, settings ActionSettings) error {
	return errors.WithStack(registerAction(true, act, settings))
}

func registerActions(override bool, customActionMap map[string]ActionSettings) error {
	for act, settings := range customActionMap {
		if err := registerAction(override, act, settings); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func registerAction(override bool, act string, settings ActionSettings) error {
	ahLock.Lock()
	defer ahLock.Unlock()

	act = strings.ToLower(act)
	if !override && actionHandler[act] != nil {
		return errors.Errorf("action<%s> already registered as type<%T>", act, actionHandler[act])
	}
	actionHandler[act] = settings
	return nil
}

// ResetDefaultActions reset action list to default list. Used e.g. for tests overriding default actions
func ResetDefaultActions() {
	actionHandler = map[string]ActionSettings{
		ActionConnectWs:             nil,
		ActionOpenApp:               OpenAppSettings{},
		ActionOpenHub:               OpenHubSettings{},
		ActionGenerateOdag:          GenerateOdagSettings{},
		ActionDeleteOdag:            DeleteOdagSettings{},
		ActionCreateSheet:           CreateSheetSettings{},
		ActionCreateBookmark:        CreateBookmarkSettings{},
		ActionDeleteBookmark:        DeleteBookmarkSettings{},
		ActionApplyBookmark:         ApplyBookmarkSettings{},
		ActionSetScript:             SetScriptSettings{},
		ActionChangeSheet:           ChangeSheetSettings{},
		ActionSelect:                SelectionSettings{},
		ActionClearAll:              ClearAllSettings{},
		ActionIterated:              IteratedSettings{},
		ActionThinkTime:             ThinkTimeSettings{},
		ActionRandom:                RandomActionSettings{},
		ActionSheetChanger:          SheetChangerSettings{},
		ActionDuplicateSheet:        DuplicateSheetSettings{},
		ActionReload:                ReloadSettings{},
		ActionProductVersion:        ProductVersionSettings{},
		ActionPublishSheet:          PublishSheetSettings{},
		ActionUnPublishSheet:        UnPublishSheetSettings{},
		ActionDisconnectApp:         DisconnectAppSettings{},
		ActionDeleteSheet:           DeleteSheetSettings{},
		ActionPublishBookmark:       PublishBookmarkSettings{},
		ActionUnPublishBookmark:     UnPublishBookmarkSettings{},
		ActionSubscribeObjects:      SubscribeObjectsSettings{},
		ActionUnsubscribeObjects:    UnsubscribeObjects{},
		ActionListBoxSelect:         ListBoxSelectSettings{},
		ActionDisconnectEnvironment: DisconnectEnvironment{},
		ActionClickActionButton:     ClickActionButtonSettings{},
		ActionContainerTab:          ContainerTabSettings{},
		ActionDoSave:                DoSaveSettings{},
		ActionClearField:            ClearFieldSettings{},
		ActionAskHubAdvisor:         AskHubAdvisorSettings{},
		ActionSetSenseVariable:      SetSenseVariableSettings{},
		ActionSetScriptVar:          SetScriptVarSettings{},
		ActionSmartSearch:           SmartSearchSettings{},
		ActionObjectSearch:          ObjectSearchSettings{},
		ActionGetScript:             GetscriptSettings{},
	}
}

// RegisteredActions returns a list of currently registered actions
func RegisteredActions() []string {
	ahLock.Lock()
	defer ahLock.Unlock()

	actionList := make([]string, len(actionHandler))
	i := 0
	for registerAction := range actionHandler {
		actionList[i] = registerAction
		i++
	}
	return actionList
}

// Implement error interface
func (err AbortedError) Error() string {
	return "action aborted"
}

// CheckActionError check action errors, returns true if error cause is aborted action, error containing any remaining error after flattening possible multi.Error
func CheckActionError(err error) (bool, error) {
	switch errors.Cause(err).(type) {
	case AbortedError:
		return true, err
	case *multierror.Error:
		mErr := helpers.FlattenMultiError(errors.Cause(err).(*multierror.Error))
		if _, isStillMultiError := mErr.(*multierror.Error); isStillMultiError {
			return false, mErr
		}
		return CheckActionError(mErr)
	default:
		return false, err
	}
}

// UnmarshalJSON unmarshal action
func (act *Action) UnmarshalJSON(arg []byte) error {
	var tmpAction actionTemp
	if err := json.Unmarshal(arg, &tmpAction); err != nil {
		return errors.Wrap(err, "Failed to unmarshal action")
	}

	act.ActionCore = tmpAction.ActionCore
	act.Type = strings.ToLower(act.Type)

	settings := NewActionsSettings(act.Type)
	if settings == nil {
		return errors.Errorf("Invalid action<%s>", tmpAction.Type)
	}

	if tmpAction.Settings != nil {
		if err := json.Unmarshal(tmpAction.Settings, &settings); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal action<%s> settings", tmpAction.Type)
		}

		if settings == nil {
			return errors.Errorf("Invalid action settings<%s>", string(tmpAction.Settings))
		}
	}

	var ok bool
	(*act).Settings, ok = settings.(ActionSettings)
	if !ok {
		return errors.Errorf("Settings<%T> not of type ActionSettings", settings)
	}

	return nil
}

// Validate scenario action
func (act *Action) Validate() ([]string, error) {
	if act.Disabled {
		return nil, nil // skip validating disabled actions
	}
	return act.Settings.Validate()
}

// Execute scenario action
func (act *Action) Execute(sessionState *session.State, connectionSettings *connection.ConnectionSettings) error {
	if act.Disabled {
		return nil
	}

	restart := true
	var err error
	for restart {
		sessionState.AwaitReconnect()
		restart, err = act.execute(sessionState, connectionSettings)
	}

	return err
}

func (act *Action) execute(sessionState *session.State, connectionSettings *connection.ConnectionSettings) (bool, error) {
	actionState := &action.State{}
	sessionState.CurrentActionState = actionState

	//check if start action should be delayed
	originalActionEntry := act.startAction(sessionState)
	restart := false

	// execute action
	var panicErr error
	func() {
		defer helpers.RecoverWithError(&panicErr)
		act.Settings.Execute(sessionState, actionState, connectionSettings, act.Label, func() {
			act.resetAction(sessionState, originalActionEntry)
		})
	}()
	if panicErr != nil {
		if sessionState != nil && sessionState.LogEntry != nil {
			sessionState.LogEntry.LogError(panicErr)
		}
		return false, errors.WithStack(panicErr)
	}

	if actionState.Failed {
		err := actionState.Errors()
		if !sessionState.IsAbortTriggered() && sessionState.ReconnectSettings.Reconnect && !act.IsContainerAction() && sessionState.IsSenseWebsocketDisconnected(err) {
			sessionState.AwaitReconnect()
			if sessionState.IsAbortTriggered() {
				if err = sessionState.GetReconnectError(); err != nil {
					actionState.AddErrors(errors.WithStack(err))
				} else {
					actionState.AddErrors(errors.New("Websocket unexpectedly closed"))
				}
			} else {
				// Fake an actionState for reconnect as a successful one
				restart = !actionState.NoRestartOnDisconnect
				actionState = &action.State{Details: actionState.Details}

				// rename action and label if we had a reconnect
				switch act.Settings.(type) {
				case ThinkTimeSettings, *ThinkTimeSettings:
					// Don't rename action if it's a thinktime to not affect analyzer results
				default:
					sessionState.LogEntry.Action.Action = fmt.Sprintf("Reconnect(%s)", sessionState.LogEntry.Action.Action)
				}
				sessionState.LogEntry.Action.Label = fmt.Sprintf("Reconnect(%s)", sessionState.LogEntry.Action.Label)
			}
		}
	}

	return restart, errors.WithStack(act.endAction(sessionState, actionState, originalActionEntry))
}

func (act *Action) startAction(sessionState *session.State) *logger.ActionEntry {
	actionEntry := &logger.ActionEntry{
		Action:   act.Type,
		ActionID: sessionState.Counters.ActionID.Inc(),
		Label:    act.Label,
	}
	act.setActionStart(sessionState, actionEntry, "START")
	return actionEntry
}

func (act *Action) resetAction(sessionState *session.State, actionEntry *logger.ActionEntry) {
	act.setActionStart(sessionState, actionEntry, "RESET")
}

func (act *Action) setActionStart(sessionState *session.State, actionEntry *logger.ActionEntry, typ string) {
	sessionState.LogEntry.SetActionEntry(actionEntry)

	sessionState.RequestMetrics.Reset()
	if typ == "" {
		typ = "START"
	}
	sessionState.LogEntry.LogDebugf("%s %s", act.Type, typ)
}

func (act *Action) endAction(sessionState *session.State, actionState *action.State, originalActionEntry *logger.ActionEntry) error {
	var containerActionEntry *logger.ActionEntry
	if act.IsContainerAction() {
		containerActionEntry = originalActionEntry
	}

	var errs *multierror.Error
	errs = multierror.Append(errs, logResult(sessionState, actionState, actionState.Details, containerActionEntry))
	sessionState.LogEntry.LogDebugf("%s END", act.Type)
	errs = multierror.Append(errs, logObjectRegressionData(sessionState))

	return helpers.FlattenMultiError(errs)
}

// logObjectRegressionData writes the currently subscribed objects to regression
// log if regession logging is enabled. The objects are locked and read from
// state shared shared by actions.
func logObjectRegressionData(sessionState *session.State) error {
	if !sessionState.LogEntry.ShouldLogRegression() {
		return nil
	}
	if sessionState.Connection == nil {
		// some actions do not have a connetion set up
		return nil
	}
	uplink := sessionState.Connection.Sense()
	if uplink == nil {
		return errors.New("could not log regression data: no sense connection")
	}

	// log regression analysis data for each subscribed object
	err := uplink.Objects.ForEachWithLock(
		func(obj *enigmahandlers.Object) error {
			// skip sheet and app types
			if obj.Type != enigmahandlers.ObjTypeGenericObject {
				return nil
			}
			genObj, ok := obj.EnigmaObject.(*enigma.GenericObject)
			if !ok {
				return errors.Errorf("expected object of type %T, but got %T", genObj, obj.EnigmaObject)
			}

			err := sessionState.LogEntry.LogRegression(
				// use unique id SESSION.ACTION.OBJECT
				fmt.Sprintf("%d.%d.%s", sessionState.LogEntry.Session.Session, sessionState.LogEntry.Action.ActionID, obj.ID),
				map[string]interface{}{
					"hyperCubeDataPages":  obj.HyperCubeDataPages(),
					"hyperCubeStackPages": obj.HyperCubeStackPages(),
					"hyperCubePivotPages": obj.HyperPivotPages(),
					"hyperCube":           obj.HyperCube(),
					"listObject":          obj.ListObject(),
					"listObjectDataPages": obj.ListObjectDataPages(),
				},
				map[string]interface{}{
					"actionType":  sessionState.LogEntry.Action.Action,
					"actionLabel": sessionState.LogEntry.Action.Label,
					"actionID":    sessionState.LogEntry.Action.ActionID,
					"objectID":    obj.ID,
					"objectType":  genObj.GenericType,
					"sessionID":   sessionState.LogEntry.Session.Session,
				})
			return errors.Wrap(err, "failed to log regression data")
		},
	)
	return errors.WithStack(err)
}

// AppStructureAction returns if this action should be included when getting app structure
// and any additional sub actions which should also be included
func (act *Action) AppStructureAction() (*AppStructureInfo, []Action) {
	appStruct, ok := act.Settings.(AppStructureAction)
	if !ok {
		return nil, nil
	}
	return appStruct.AppStructureAction()
}

// IsContainerAction returns true if action settings implements ContainerAction interface
func (act *Action) IsContainerAction() bool {
	if _, ok := act.Settings.(ContainerAction); ok {
		return true
	}
	return false
}

func logResult(sessionState *session.State, actionState *action.State, details string, containerActionEntry *logger.ActionEntry) error {
	var sent, received, requests uint64
	var responsetime int64
	var actionError error

	defer sessionState.EW.Reset()

	if sessionState == nil {
		return nil
	}

	// check if this is a container action
	isContainerAction := containerActionEntry != nil

	defer func() {
		if isAborted, err := CheckActionError(actionError); isAborted {
			// scenario was aborted, log result as info instead of result
			sessionState.LogEntry.LogInfo("Aborted", actionError.Error())
			return // Do not report result
		} else if err != nil {
			sessionState.LogEntry.LogError(err)
		}
		if !actionState.NoResults {
			logResults(sessionState, isContainerAction, !actionState.Failed, sent, received, requests, responsetime, details)
		}
	}()

	if helpers.IsContextTriggered(sessionState.BaseContext()) {
		actionError = AbortedError{}
		return actionError
	}

	// Set correct action entry for container action
	if isContainerAction {
		sessionState.LogEntry.SetActionEntry(containerActionEntry)
	}

	if sessionState.LogEntry.Action == nil {
		sessionState.LogEntry.LogError(actionError)
		actionError = errors.Errorf("action entry is nil")
		return actionError
	}

	if !isContainerAction && !actionState.NoResults { // Don't report metrics if container action
		var resp time.Duration
		resp, sent, received = sessionState.RequestMetrics.Metrics()

		if !actionState.Failed {
			if resp.Nanoseconds() > 0 {
				responsetime = resp.Nanoseconds()
				if sessionState.LogEntry.Session == nil {
					sessionState.LogEntry.Log(logger.WarningLevel, "Session entry is nil, unable to add prometheus metric")
				} else {
					buildmetrics.ReportSuccess(sessionState.LogEntry.Action.Action, sessionState.LogEntry.Action.Label, resp.Seconds())
				}
			}
		} else {
			buildmetrics.ReportFailure(sessionState.LogEntry.Action.Action, sessionState.LogEntry.Action.Label)
		}
	}

	if traffic := sessionState.TrafficLogger(); traffic != nil {
		requests = traffic.RequestCount()
		traffic.ResetRequestCount()
	}

	actionError = actionState.Errors()
	return actionError
}

func logResults(sessionState *session.State, isContainerAction, success bool, sent, received, requests uint64, responsetime int64, details string) {
	if isContainerAction {
		// log info instead of result for a container action
		sessionState.LogEntry.LogInfo("containeractionend", "")
	} else {
		sessionState.LogEntry.LogResult(success, sessionState.EW.Warnings(), sessionState.EW.Errors(), sent, received, requests, responsetime, details)
		actionStats := sessionState.Counters.StatisticsCollector.GetOrAddActionStats(sessionState.LogEntry.Action.Action, sessionState.LogEntry.Action.Label, sessionState.LogEntry.Session.AppGUID)
		if actionStats != nil {
			actionStats.WarnCount.Add(sessionState.EW.Warnings())
			actionStats.ErrCount.Add(sessionState.EW.Errors())
			actionStats.Sent.Add(sent)
			actionStats.Received.Add(received)
			actionStats.Requests.Add(requests)
			if success {
				actionStats.RespAvg.AddSample(uint64(responsetime))
			} else {
				actionStats.Failed.Inc()
			}
		}
	}
}
