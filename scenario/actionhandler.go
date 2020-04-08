package scenario

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/buildmetrics"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/globals"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/statistics"
)

type (
	// ActionSettings scenario action interface for mandatory methods
	ActionSettings interface {
		// Execute action
		Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string /* Label */, reset func() /* reset action start */)
		// Validate action
		Validate() error
	}

	// ContainerAction Implement this interface on action settings to mark an action as a
	// container action containing other actions. A container action will not log result as a normal action,
	// instead result will be logged as level=info, infotype: containeractionend
	// Returns if action is to be considered a container action.
	// ContainerAction can't be used in conjunction with StartActionOverrider interface
	ContainerAction interface {
		IsContainerAction()
	}

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
)

const (
	ActionConnectWs               = "connectws"
	ActionOpenApp                 = "openapp"
	ActionOpenHub                 = "openhub"
	ActionElasticOpenHub          = "elasticopenhub"
	ActionElasticReload           = "elasticreload"
	ActionElasticUploadApp        = "elasticuploadapp"
	ActionElasticCreateCollection = "elasticcreatecollection"
	ActionElasticDeleteCollection = "elasticdeletecollection"
	ActionElasticHubSearch        = "elastichubsearch"
	ActionElasticDeleteApp        = "elasticdeleteapp"
	ActionElasticCreateApp        = "elasticcreateapp"
	ActionElasticShareApp         = "elasticshareapp"
	ActionElasticExportApp        = "elasticexportapp"
	ActionElasticGenerateOdag     = "elasticgenerateodag"
	ActionElasticDeleteOdag       = "elasticdeleteodag"
	ActionElasticDuplicateApp     = "elasticduplicateapp"
	ActionElasticExplore          = "elasticexplore"
	ActionElasticMoveApp          = "elasticmoveapp"
	ActionGenerateOdag            = "generateodag"
	ActionDeleteOdag              = "deleteodag"
	ActionUploadData              = "uploaddata"
	ActionDeleteData              = "deletedata"
	ActionCreateSheet             = "createsheet"
	ActionCreateBookmark          = "createbookmark"
	ActionDeleteBookmark          = "deletebookmark"
	ActionApplyBookmark           = "applybookmark"
	ActionSetScript               = "setscript"
	ActionChangeSheet             = "changesheet"
	ActionStaticSelect            = "staticselect"
	ActionSelect                  = "select"
	ActionClearAll                = "clearall"
	ActionIterated                = "iterated"
	ActionThinkTime               = "thinktime"
	ActionRandom                  = "randomaction"
	ActionSheetChanger            = "sheetchanger"
	ActionDuplicateSheet          = "duplicatesheet"
	ActionReload                  = "reload"
	ActionProductVersion          = "productversion"
	ActionPublishSheet            = "publishsheet"
	ActionUnPublishSheet          = "unpublishsheet"
	ActionDisconnectApp           = "disconnectapp"
	ActionDeleteSheet             = "deletesheet"
)

// Scenario actions needs an entry in actionHandler
var (
	actionHandler map[string]ActionSettings
	ahLock        sync.Mutex
	jsonit        = jsoniter.ConfigCompatibleWithStandardLibrary
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

// RegisterAction register a custom action and override any existing with same name
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
		ActionConnectWs:               nil,
		ActionOpenApp:                 OpenAppSettings{},
		ActionOpenHub:                 OpenHubSettings{},
		ActionElasticOpenHub:          ElasticOpenHubSettings{},
		ActionElasticReload:           ElasticReloadSettings{},
		ActionElasticUploadApp:        ElasticUploadAppSettings{},
		ActionElasticCreateCollection: ElasticCreateCollectionSettings{},
		ActionElasticDeleteCollection: ElasticDeleteCollectionSettings{},
		ActionElasticHubSearch:        ElasticHubSearchSettings{},
		ActionElasticDeleteApp:        ElasticDeleteAppSettings{},
		ActionElasticCreateApp:        ElasticCreateAppSettings{},
		ActionElasticShareApp:         ElasticShareAppSettings{},
		ActionElasticExportApp:        ElasticExportAppSettings{},
		ActionElasticGenerateOdag:     ElasticGenerateOdagSettings{},
		ActionElasticDeleteOdag:       ElasticDeleteOdagSettings{},
		ActionElasticDuplicateApp:     ElasticDuplicateAppSettings{},
		ActionElasticExplore:          ElasticExploreSettings{},
		ActionElasticMoveApp:          ElasticMoveAppSettings{},
		ActionGenerateOdag:            GenerateOdagSettings{},
		ActionDeleteOdag:              DeleteOdagSettings{},
		ActionUploadData:              UploadDataSettings{},
		ActionDeleteData:              DeleteDataSettings{},
		ActionCreateSheet:             CreateSheetSettings{},
		ActionCreateBookmark:          CreateBookmarkSettings{},
		ActionDeleteBookmark:          DeleteBookmarkSettings{},
		ActionApplyBookmark:           ApplyBookmarkSettings{},
		ActionSetScript:               SetScriptSettings{},
		ActionChangeSheet:             ChangeSheetSettings{},
		ActionStaticSelect:            StaticSelectSettings{},
		ActionSelect:                  SelectionSettings{},
		ActionClearAll:                ClearAllSettings{},
		ActionIterated:                IteratedSettings{},
		ActionThinkTime:               ThinkTimeSettings{},
		ActionRandom:                  RandomActionSettings{},
		ActionSheetChanger:            SheetChangerSettings{},
		ActionDuplicateSheet:          DuplicateSheetSettings{},
		ActionReload:                  ReloadSettings{},
		ActionProductVersion:          ProductVersionSettings{},
		ActionPublishSheet:            PublishSheetSettings{},
		ActionUnPublishSheet:          UnPublishSheetSettings{},
		ActionDisconnectApp:           DisconnectAppSettings{},
		ActionDeleteSheet:             DeleteSheetSettings{},
	}
}

// RegisteredActions returns a list of currently registered actions
func RegisteredActions() []string {
	ahLock.Lock()
	defer ahLock.Unlock()

	actionList := make([]string, len(actionHandler))
	i := 0
	for action := range actionHandler {
		actionList[i] = action
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
	if err := jsonit.Unmarshal(arg, &tmpAction); err != nil {
		return errors.Wrap(err, "Failed to unmarshal action")
	}

	act.ActionCore = tmpAction.ActionCore
	act.Type = strings.ToLower(act.Type)

	settings := NewActionsSettings(act.Type)
	if settings == nil {
		return errors.Errorf("Invalid action<%s>", tmpAction.Type)
	}

	if tmpAction.Settings != nil {
		if err := jsonit.Unmarshal(tmpAction.Settings, &settings); err != nil {
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
func (act *Action) Validate() error {
	if act.Disabled {
		return nil // skip validating disabled actions
	}
	s, ok := act.Settings.(ActionSettings)
	if !ok {
		return errors.Errorf("Failed to convert action settings to ActionSettings")
	}
	return errors.WithStack(s.Validate())
}

// Execute scenario action
func (act *Action) Execute(sessionState *session.State, connectionSettings *connection.ConnectionSettings) error {

	if act.Disabled {
		return nil
	}

	actionState := &action.State{}
	sessionState.CurrentActionState = actionState

	//check if start action should be delayed
	originalActionEntry := act.startAction(sessionState)

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
		return errors.WithStack(panicErr)
	}

	return errors.WithStack(act.endAction(sessionState, actionState, originalActionEntry))
}

func (act *Action) startAction(sessionState *session.State) *logger.ActionEntry {
	actionEntry := &logger.ActionEntry{
		Action:   act.Type,
		ActionID: globals.ActionID.Inc(),
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
	if _, ok := act.Settings.(ContainerAction); ok {
		containerActionEntry = originalActionEntry
	}

	err := logResult(sessionState, actionState, actionState.Details, containerActionEntry)
	sessionState.LogEntry.LogDebugf("%s END", act.Type)
	return errors.WithStack(err)
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
		actionStats := statistics.GetOrAddGlobalActionStats(sessionState.LogEntry.Action.Action, sessionState.LogEntry.Action.Label, sessionState.LogEntry.Session.AppGUID)
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
