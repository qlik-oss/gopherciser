package scenario

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	OpenAppSettingsTimeouts struct {
		Connect helpers.TimeDuration `json:"connect" displayname:"Connection timeout" doc-key:"openapp.timeouts.connect"`
		Open    helpers.TimeDuration `json:"open" displayname:"Open app timeout" doc-key:"openapp.timeouts.open"`
	}
	OpenAppSettingsCore struct {
		ExternalHost  string                  `json:"externalhost" displayname:"External hostname" doc-key:"openapp.externalhost"`
		UniqueSession bool                    `json:"unique" displayname:"Make session unique" doc-key:"openapp.unique"`
		Timeouts      OpenAppSettingsTimeouts `json:"timeouts" displayname:"Timeouts" doc-key:"openapp.timeouts"`
	}
	// OpenAppSettings app and server settings
	OpenAppSettings struct {
		session.AppSelection
		OpenAppSettingsCore
	}

	connectWsSettings struct {
		ConnectFunc func(bool) (string, error)
		Timeout     time.Duration
	}
)

// UnmarshalJSON unmarshals open app settings from JSON
func (openApp *OpenAppSettings) UnmarshalJSON(arg []byte) error {
	// Check for deprecated fields
	if err := helpers.HasDeprecatedFields(arg, []string{
		"/appguid",
		"/appname",
		"/randomguids",
		"/randomapps",
		"/mode",
	}); err != nil {
		return errors.Errorf("%s %s, please remove from script", ActionOpenApp, err.Error())
	}

	if err := json.Unmarshal(arg, &openApp.OpenAppSettingsCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionOpenApp)
	}

	if err := json.Unmarshal(arg, &openApp.AppSelection); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionOpenApp)
	}
	return nil
}

// Execute open app
func (openApp OpenAppSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, setOpenStart func()) {
	actionState.FailOnDisconnect = true
	appEntry, err := openApp.Select(sessionState)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "Failed to perform app selection"))
		return
	}

	TrySetCSRFToken(sessionState, actionState, connectionSettings)

	actionState.Details = sessionState.LogEntry.Session.AppName
	var headers http.Header
	if openApp.UniqueSession {
		headers = make(http.Header, 1)
		headers.Add("X-Qlik-Session", uuid.NewString())
	}

	timeout := time.Duration(openApp.Timeouts.Connect)
	if timeout < 1 {
		timeout = sessionState.Timeout
	}

	connectFunc, err := connectionSettings.GetConnectFunc(sessionState, appEntry.ID, openApp.ExternalHost, headers, timeout)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "Failed to get connect function"))
		return
	}

	// if we have a label, add connect websocket sub-action
	var wsLabel string
	if label != "" {
		wsLabel = fmt.Sprintf("%s - WS", label)
	}

	connectWs := GetConnectWsAction(wsLabel, time.Duration(openApp.Timeouts.Connect), connectFunc)

	//Connect websocket and logs as separate action
	actionState.NoResults = true // temporary set to not report while doing sub action.
	if isAborted, err := CheckActionError(connectWs.Execute(sessionState, connectionSettings)); isAborted {
		return // action is aborted, we should not continue
	} else if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	// Update opened apps counter
	sessionState.Counters.StatisticsCollector.IncOpenedApps()

	setOpenStart()
	actionState.NoResults = false // make sure to report results for main action

	uplink := sessionState.Connection.Sense()

	DoOpenApp(sessionState, actionState, uplink, appEntry.ID, time.Duration(openApp.Timeouts.Open))
	if actionState.Failed {
		return
	}

	DoPostOpenAppRequests(sessionState, actionState, uplink, appEntry.ID)

	// setup re-connect function
	sessionState.SetReconnectFunc(connectFunc)

	sessionState.Wait(actionState)
}

// Validate open app scenario item
func (openApp OpenAppSettings) Validate() ([]string, error) {
	if err := openApp.AppSelection.Validate(); err != nil {
		return nil, err
	}

	return nil, nil
}

func openDoc(ctx context.Context, uplink *enigmahandlers.SenseUplink, appGUID string) error {
	doc, err := uplink.Global.OpenDoc(ctx, appGUID, "", "", "", false)
	if err != nil {
		return err
	}
	return uplink.SetCurrentApp(appGUID, doc)
}

func GetConnectWsAction(wsLabel string, timeout time.Duration, connectFunc func(bool) (string, error)) Action {
	connectWs := Action{
		ActionCore{
			Type:  ActionConnectWs,
			Label: wsLabel,
		},
		connectWsSettings{
			ConnectFunc: connectFunc,
			Timeout:     timeout,
		},
	}
	return connectWs
}

// AppStructureAction implements AppStructureAction interface
func (openApp OpenAppSettings) AppStructureAction() (*AppStructureInfo, []Action) {
	return &AppStructureInfo{
		IsAppAction: true,
		Include:     true,
	}, nil
}

func (connectWs connectWsSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	actionState.FailOnDisconnect = true
	appGUID, err := connectWs.ConnectFunc(false)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "Failed connecting to sense server"))
		return
	}
	actionState.Details = appGUID

	if err := sessionState.SetupChangeChan(); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to setup change channel"))
		return
	}
}

func (connectWs connectWsSettings) Validate() ([]string, error) {
	return nil, nil
}

// AffectsAppObjectsAction implements AffectsAppObjectsAction interface
func (openApp OpenAppSettings) AffectsAppObjectsAction(structure appstructure.AppStructure) ([]*appstructure.AppStructurePopulatedObjects, []string, bool) {
	newObjs := appstructure.AppStructurePopulatedObjects{
		Parent:    openApp.App.String(),
		Objects:   make([]appstructure.AppStructureObject, 0),
		Bookmarks: nil,
	}
	for _, obj := range structure.Objects {
		if obj.Type == "sheet" {
			newObjs.Objects = append(newObjs.Objects, obj)
		}
	}
	for _, v := range structure.Bookmarks {
		newObjs.Bookmarks = append(newObjs.Bookmarks, v)
	}
	return []*appstructure.AppStructurePopulatedObjects{&newObjs}, nil, true
}

// DoOpenApp is intended to be used from inside a open app action after websocket is connected
func DoOpenApp(sessionState *session.State, actionState *action.State, uplink *enigmahandlers.SenseUplink, appGUID string, timeout time.Duration) {
	if err := sessionState.SendRequestWithTimeout(actionState, timeout, func(ctx context.Context) error {
		return openDoc(ctx, uplink, appGUID)
	}); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "Failed to open app GUID<%s>", appGUID))
		return
	}

	if uplink.CurrentApp == nil {
		actionState.AddErrors(errors.New("No current app"))
		return
	}
	if uplink.CurrentApp.Doc == nil {
		actionState.AddErrors(errors.New("No current enigma doc"))
		return
	}
}

func DoPostOpenAppRequests(sessionState *session.State, actionState *action.State, uplink *enigmahandlers.SenseUplink, appID string) {
	if uplink.CurrentApp == nil {
		actionState.AddErrors(errors.New("No current app"))
		return
	}
	if uplink.CurrentApp.Doc == nil {
		actionState.AddErrors(errors.New("No current enigma doc"))
		return
	}

	doc := uplink.CurrentApp.Doc

	var authUser string
	// Ask for user synchronously to make sure it's on all subsequent log entries
	getAuthUser := func(ctx context.Context) error {
		var err error
		authUser, err = uplink.Global.GetAuthenticatedUser(ctx)
		return err
	}
	if err := sessionState.SendRequest(actionState, getAuthUser); err != nil {
		actionState.AddErrors(err)
		return
	}
	sessionState.LogEntry.LogInfo("AuthenticatedUser", authUser)

	// send another AuthenticatedUser for api compliance
	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.Global.GetAuthenticatedUser(ctx)
		return err
	}, actionState, true, "")

	sessionState.QueueRequest(func(ctx context.Context) error {
		layout, applyOutErr := doc.GetAppLayout(ctx)
		if applyOutErr != nil {
			return applyOutErr
		}
		uplink.CurrentApp.Layout = layout
		return nil
	}, actionState, true, fmt.Sprintf("Failed getting app layout for app GUID<%s>", appID))

	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.CurrentApp.GetVariableList(sessionState, actionState)
		return errors.WithStack(err)
	}, actionState, true, "")

	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.CurrentApp.GetStoryList(sessionState, actionState)
		return errors.WithStack(err)
	}, actionState, true, "")

	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.CurrentApp.GetAppsPropsList(sessionState, actionState)
		return errors.WithStack(err)
	}, actionState, true, "")

	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.Global.AllowCreateApp(ctx)
		return errors.WithStack(err)
	}, actionState, true, "")

	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.CurrentApp.Doc.GetScriptEx(ctx) // ignore err, as when not ownning app an Access denied will be returned.
		if err != nil {
			sessionState.LogEntry.LogDebugf("GetScriptEx request returned error: %v", err)
		}
		return nil
	}, actionState, true, "")

	for i := 0; i < 2; i++ {
		sessionState.QueueRequestRaw(uplink.CurrentApp.Doc.GetAppPropertiesRaw, actionState, true, "failed to get AppProperties")
	}
	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.Global.GetBaseBNFHash(ctx, "S")
		return err
	}, actionState, true, "")

	// Send GetConfiguration request 5 times
	for i := 0; i < 5; i++ {
		sessionState.QueueRequest(func(ctx context.Context) error {
			return errors.WithStack(uplink.Global.RPC(ctx, "GetConfiguration", nil))
		}, actionState, false, "GetConfiguration request failed")
	}

	sessionState.GetSheetList(actionState, uplink)
	if actionState.Failed {
		return
	}
}

func TrySetCSRFToken(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings) {
	if sessionState.TargetEnv() == "" { // No preceeding action performed to determine target environment
		host, err := connectionSettings.RestUrl()
		if err != nil {
			actionState.AddErrors(err)
			return
		}
		noContentOptions := session.DefaultReqOptions()
		noContentOptions.ExpectedStatusCode = []int{http.StatusNoContent, http.StatusOK, http.StatusNotFound}
		noContentOptions.FailOnError = false
		_, _ = sessionState.Rest.GetSyncWithCallback(fmt.Sprintf("%s/qps/csrftoken", host), actionState, sessionState.LogEntry, noContentOptions, func(err error, req *session.RestRequest) {
			if err != nil {
				return
			}
			if req.ResponseStatusCode == http.StatusNoContent {
				connectionSettings.SetCSRFToken(req.ResponseHeaders.Get("qlik-csrf-token"))
			}
		})
	}
}
