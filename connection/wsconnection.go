package connection

import (
	"net/http"
	"net/http/cookiejar"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ConnectWsSettings app and server settings using WS
	ConnectWsSettings struct{}
)

// GetConnectFunc get ws connect function
func (connectWs *ConnectWsSettings) GetConnectFunc(sessionState *session.State, connectionSettings *ConnectionSettings, appGUID string, header http.Header) func() (string, error) {
	return func() (string, error) {
		if sessionState == nil {
			return appGUID, errors.New("Session state is nil")
		}

		if sessionState.Connection == nil {
			sessionState.Connection = new(enigmahandlers.SenseConnection)
		} else {
			sessionState.Disconnect()
		}

		sense := enigmahandlers.NewSenseUplink(sessionState.BaseContext(), sessionState.LogEntry, sessionState.RequestMetrics, sessionState.TrafficLogger())
		sessionState.Connection.SetSense(sense)
		sense.OnUnexpectedDisconnect(sessionState.WSFailed)

		url, err := connectionSettings.GetURL(appGUID)
		if err != nil {
			return appGUID, errors.WithStack(err)
		}

		if sessionState.Cookies == nil {
			sessionState.Cookies, err = cookiejar.New(nil)
			if err != nil {
				return appGUID, errors.Wrap(err, "failed creating cookie jar")
			}
		}

		if err := sense.Connect(sessionState.BaseContext(), url, header, sessionState.Cookies, connectionSettings.Allowuntrusted, sessionState.Timeout); err != nil {
			return appGUID, errors.Wrap(err, "Failed connecting to sense server")
		}

		return appGUID, nil
	}
}

// Validate open app scenario item
func (connectWs *ConnectWsSettings) Validate() error {
	return nil
}
