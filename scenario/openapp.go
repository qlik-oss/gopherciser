package scenario

import (
	"context"
	"fmt"
	"github.com/qlik-oss/gopherciser/appstructure"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/statistics"
)

type (
	// OpenAppSettings app and server settings
	OpenAppSettings struct {
		session.AppSelection
	}

	connectWsSettings struct {
		ConnectFunc func() (string, error)
	}

	// Older settings no longer used, if exist in JSON, an error will be thrown
	deprecatedOpenAppSettings struct {
		AppGUID        string      `json:"appguid"`
		AppName        string      `json:"appname"`
		RandomGUIDs    []string    `json:"randomguids"`
		RandomApps     []string    `json:"randomapps"`
		ConnectionMode interface{} `json:"mode"`
	}
)

// UnmarshalJSON unmarshals open app settings from JSON
func (openApp *OpenAppSettings) UnmarshalJSON(arg []byte) error {
	var deprecated deprecatedOpenAppSettings
	if err := jsonit.Unmarshal(arg, &deprecated); err == nil { // skip check if error
		hasSettings := make([]string, 0, 5)
		if deprecated.AppGUID != "" {
			hasSettings = append(hasSettings, "appguid")
		}
		if deprecated.AppName != "" {
			hasSettings = append(hasSettings, "appname")
		}
		if len(deprecated.RandomGUIDs) > 0 {
			hasSettings = append(hasSettings, "randomguids")
		}
		if len(deprecated.RandomApps) > 0 {
			hasSettings = append(hasSettings, "randomapps")
		}
		if deprecated.ConnectionMode != nil {
			hasSettings = append(hasSettings, "mode")
		}
		if len(hasSettings) > 0 {
			return errors.Errorf("%s settings<%s> are no longer used", ActionOpenApp, strings.Join(hasSettings, ","))
		}
	}
	var appSelection session.AppSelection
	if err := jsonit.Unmarshal(arg, &appSelection); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionOpenApp)
	}
	*openApp = OpenAppSettings{appSelection}
	return nil
}

// Execute open app
func (openApp OpenAppSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, setOpenStart func()) {
	if !sessionState.LoggedIn {
		headers, err := connectionSettings.GetHeaders(sessionState)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to connect using OpenApp"))
			return
		}
		host, err := connectionSettings.GetHost()
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to extract hostname"))
			return
		}
		sessionState.HeaderJar.SetHeader(host, headers)
		sessionState.LoggedIn = true
		client, err := session.DefaultClient(connectionSettings, sessionState)
		if err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return
		}
		if sessionState.Cookies != nil {
			client.Jar = sessionState.Cookies
		}
		sessionState.Rest.SetClient(client)
	}

	appEntry, err := openApp.AppSelection.Select(sessionState)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "Failed to perform app selection"))
		return
	}

	actionState.Details = sessionState.LogEntry.Session.AppName

	connectFunc, err := connectionSettings.GetConnectFunc(sessionState, appEntry.GUID)
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

	// Update opened apps global counter
	statistics.IncOpenedApps()

	setOpenStart()
	actionState.NoResults = false // make sure to report results for main action

	uplink := sessionState.Connection.Sense()

	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		return openDoc(ctx, uplink, appEntry.GUID)
	}); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "Failed to open app GUID<%s>", appEntry.GUID))
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
	}, actionState, true, fmt.Sprintf("Failed getting app layout for app GUID<%s>", appEntry.GUID))

	sessionState.QueueRequest(func(ctx context.Context) error {
		version, versionErr := uplink.Global.EngineVersion(ctx)
		if versionErr != nil {
			return errors.Wrap(versionErr, "Failed to get engine version")
		}

		sessionState.LogEntry.LogInfo("EngineVersion", version.ComponentVersion)
		return nil
	}, actionState, false, fmt.Sprintf("Failed getting engine version"))

	sessionState.QueueRequest(func(ctx context.Context) error {
		idm, desktopErr := uplink.Global.IsDesktopMode(ctx)
		sessionState.LogEntry.LogInfo("IsDesktopMode", fmt.Sprintf("%v", idm))
		return desktopErr
	}, actionState, true, "Failed getting authenticated user")

	sheetList, err := uplink.CurrentApp.GetSheetList(sessionState, actionState)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	slLayout := sheetList.Layout()
	if slLayout == nil {
		actionState.AddErrors(errors.New("sheetlist layout is nil"))
		return
	}

	if sessionState.LogEntry.ShouldLogDebug() &&
		slLayout.AppObjectList != nil &&
		slLayout.AppObjectList.Items != nil {

		for _, v := range slLayout.AppObjectList.Items {
			sessionState.LogEntry.LogDebugf("Sheet<%s> found", v.Info.Id)
		}
	}

	sessionState.Wait(actionState)
}

// Validate open app scenario item
func (openApp OpenAppSettings) Validate() error {
	if err := openApp.AppSelection.Validate(); err != nil {
		return err
	}

	return nil
}

func openDoc(ctx context.Context, connection *enigmahandlers.SenseUplink, appGUID string) error {
	doc, err := connection.Global.OpenDoc(ctx, appGUID, "", "", "", false)
	if err != nil {
		return err
	}
	err = connection.Objects.AddObject(&enigmahandlers.Object{
		Handle:       doc.ObjectInterface.Handle,
		Type:         enigmahandlers.ObjTypeApp,
		EnigmaObject: doc,
	})
	if err != nil {
		return err
	}
	connection.CurrentApp = &senseobjects.App{
		GUID: appGUID,
		Doc:  doc,
	}
	return nil
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

func (connectWs connectWsSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	appGUID, err := connectWs.ConnectFunc()
	actionState.Details = appGUID
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "Failed connecting to sense server"))
		return
	}

	if sessionState == nil || sessionState.Connection == nil || sessionState.Connection.Sense() == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("No connection for setting up pushed messages listener"))
		return
	}

	changeChan := sessionState.Connection.Sense().Global.ChangeListsChannel(true)
	go func() {
		for {
			select {
			case cl, ok := <-changeChan:
				if !ok {
					return
				}

				if len(cl.Changed) > 0 {
					sessionState.LogEntry.LogInfo("Pushed ChangedList", fmt.Sprintf("%v", cl.Changed))
				}

				if len(cl.Closed) > 0 {
					sessionState.LogEntry.LogInfo("Pushed ClosedList", fmt.Sprintf("%v", cl.Closed))
				}

				sessionState.TriggerEvents(sessionState.CurrentActionState, cl.Changed, cl.Closed)
			case <-sessionState.BaseContext().Done():
				return
			}

		}
	}()

}

func (connectWs connectWsSettings) Validate() error {
	return nil
}

// AffectsAppObjectsAction implements AffectsAppObjectsAction interface
func (settings OpenAppSettings) AffectsAppObjectsAction(structure appstructure.AppStructure) (*appstructure.AppStructurePopulatedObjects, []string, bool) {
	newObjs := appstructure.AppStructurePopulatedObjects{
		Parent:    settings.App.String(),
		Objects:   make([]appstructure.AppStructureObject, 0),
		Bookmarks: nil,
	}
	for _, obj := range structure.Objects {
		if obj.Type == "sheet" {
			newObjs.Sheets = append(newObjs.Sheets, obj)
		}
	}
	for _, v := range structure.Bookmarks {
		newObjs.Bookmarks = append(newObjs.Bookmarks, v)
	}
	return &newObjs, nil, true
}
