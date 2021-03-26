package scenario

import (
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	UnPublishBookmarkSettings struct {
		BookMarkSettings
	}
)

// Validate PublishBookmarkSettings action (Implements ActionSettings interface)
func (settings UnPublishBookmarkSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute CreateBookmarkSettings action (Implements ActionSettings interface)
func (settings UnPublishBookmarkSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("not connected to a Sense environment"))
		return
	}
	uplink := sessionState.Connection.Sense()

	bm, err := settings.getBookmarkObject(sessionState, actionState, uplink)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	// un-publish bookmark
	if err := sessionState.SendRequest(actionState, bm.UnPublish); err != nil {
		actionState.AddErrors(err)
		return
	}
	sessionState.Wait(actionState)
}
