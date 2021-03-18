package scenario

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// OpenHubSettings settings for OpenHub
	OpenHubSettings struct{}
)

// Validate open app scenario item
func (openHub OpenHubSettings) Validate() error {
	return nil
}

// Execute execute the action
func (openHub OpenHubSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, setHubStart func()) {
	actionState.Details = "SenseEfW"

	connectFunc, err := connectionSettings.GetConnectFunc(sessionState, "", nil)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to get connect function"))
		return
	}

	var wslabel string
	if label != "" {
		wslabel = fmt.Sprintf("%s - WS", label)
	}

	connectWs := Action{
		ActionCore{
			Type:  ActionConnectWs,
			Label: wslabel,
		},
		connectWsSettings{
			ConnectFunc: connectFunc,
		},
	}

	//Connect websocket and logs as separate action
	actionState.NoResults = true // temporary set to not report while doing sub action.
	if isAborted, err := CheckActionError(connectWs.Execute(sessionState, connectionSettings)); isAborted {
		return // action is aborted, we should not continue
	} else if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	setHubStart()
	actionState.NoResults = false // make sure to report results for main action

	upLink := sessionState.Connection.Sense()
	defer func() {
		upLink.Disconnect()
		sessionState.Connection = nil
	}()

	sessionState.QueueRequest(func(ctx context.Context) error {
		docList, err := upLink.Global.GetDocList(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to get DocList")
		}
		if err = sessionState.ArtifactMap.FillAppsUsingDocListEntries(docList); err != nil {
			return errors.Wrap(err, "failed to populate app list")
		}
		if _, err := upLink.Global.GetStreamList(ctx); err != nil {
			return errors.Wrap(err, "failed to get stream list")
		}
		return nil
	}, actionState, true, "failed to get DocList object")
	sessionState.Wait(actionState)

	// setup re-connect function
	sessionState.SetReconnectFunc(connectFunc)

	if err := sessionState.ArtifactMap.LogMap(sessionState.LogEntry); err != nil {
		sessionState.LogEntry.Log(logger.WarningLevel, err)
	}
}

// AppStructureAction implements AppStructureAction interface
func (openHub OpenHubSettings) AppStructureAction() (*AppStructureInfo, []Action) {
	return &AppStructureInfo{
		IsAppAction: false,
		Include:     true,
	}, nil
}
