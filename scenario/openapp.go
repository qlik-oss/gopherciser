package scenario

import (
	"context"
	"fmt"
	"net/http"

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
	OpenAppSettingsCore struct {
		UniqueSession bool `json:"unique" displayname:"Make session unique" doc-key:"openapp.unique"`
	}
	// OpenAppSettings app and server settings
	OpenAppSettings struct {
		session.AppSelection
		OpenAppSettingsCore
	}

	connectWsSettings struct {
		ConnectFunc func() (string, error)
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

	if err := jsonit.Unmarshal(arg, &openApp.OpenAppSettingsCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionOpenApp)
	}

	if err := jsonit.Unmarshal(arg, &openApp.AppSelection); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionOpenApp)
	}
	return nil
}

// Execute open app
func (openApp OpenAppSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, setOpenStart func()) {
	appEntry, err := openApp.AppSelection.Select(sessionState)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "Failed to perform app selection"))
		return
	}

	actionState.Details = sessionState.LogEntry.Session.AppName
	var headers http.Header
	if openApp.UniqueSession {
		headers = make(http.Header, 1)
		headers.Add("X-Qlik-Session", uuid.NewString())
	}
	connectFunc, err := connectionSettings.GetConnectFunc(sessionState, appEntry.ID, headers)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "Failed to get connect function"))
		return
	}

	// if we have a label, add connect websocket sub-action
	var wsLabel string
	if label != "" {
		wsLabel = fmt.Sprintf("%s - WS", label)
	}

	connectWs := openApp.GetConnectWsAction(wsLabel, connectFunc)

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

	openApp.doOpen(sessionState, actionState, uplink, appEntry.ID)
	if actionState.Failed {
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

	sessionState.QueueRequest(func(ctx context.Context) error {
		layout, applyOutErr := doc.GetAppLayout(ctx)
		if applyOutErr != nil {
			return applyOutErr
		}
		uplink.CurrentApp.Layout = layout
		return nil
	}, actionState, true, fmt.Sprintf("Failed getting app layout for app GUID<%s>", appEntry.ID))

	sessionState.QueueRequest(func(ctx context.Context) error {
		version, versionErr := uplink.Global.EngineVersion(ctx)
		if versionErr != nil {
			return errors.Wrap(versionErr, "Failed to get engine version")
		}

		sessionState.LogEntry.LogInfo("EngineVersion", version.ComponentVersion)
		return nil
	}, actionState, false, "Failed getting engine version")

	sessionState.QueueRequest(func(ctx context.Context) error {
		idm, desktopErr := uplink.Global.IsDesktopMode(ctx)
		sessionState.LogEntry.LogInfo("IsDesktopMode", fmt.Sprintf("%v", idm))
		return desktopErr
	}, actionState, true, "Failed getting authenticated user")

	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.CurrentApp.GetVariableList(sessionState, actionState)
		return errors.WithStack(err)
	}, actionState, true, "")

	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.CurrentApp.GetStoryList(sessionState, actionState)
		return errors.WithStack(err)
	}, actionState, true, "")

	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.CurrentApp.GetLoadModelList(sessionState, actionState)
		return errors.WithStack(err)
	}, actionState, true, "")

	sessionState.GetSheetList(actionState, uplink)
	if actionState.Failed {
		return
	}

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

func (openApp OpenAppSettings) GetConnectWsAction(wsLabel string, connectFunc func() (string, error)) Action {
	connectWs := Action{
		ActionCore{
			Type:  ActionConnectWs,
			Label: wsLabel,
		},
		connectWsSettings{
			ConnectFunc: connectFunc,
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
	appGUID, err := connectWs.ConnectFunc()
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

func (openApp OpenAppSettings) doOpen(sessionState *session.State, actionState *action.State, uplink *enigmahandlers.SenseUplink, appGUID string) {
	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
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
