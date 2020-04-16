package scenario

import (
	"context"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	PublishBookmarkSettings struct {
		BookMarkSettings
	}
)

// Validate PublishBookmarkSettings action (Implements ActionSettings interface)
func (settings PublishBookmarkSettings) Validate() error {
	return nil
}

// Execute CreateBookmarkSettings action (Implements ActionSettings interface)
func (settings PublishBookmarkSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("not connected to a Sense environment"))
		return
	}
	uplink := sessionState.Connection.Sense()

	id, _, err := settings.getBookmark(sessionState, actionState, uplink)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	// Get bookmark object
	var bm *enigma.GenericBookmark
	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		var err error
		bm, err = uplink.CurrentApp.Doc.GetBookmark(ctx, id)
		return err
	}); err != nil {
		actionState.AddErrors(err)
		return
	}

	// TODO check published / unpublished

	// publish bookmark
	if err := sessionState.SendRequest(actionState, bm.Publish); err != nil {
		actionState.AddErrors(err)
		return
	}
	sessionState.Wait(actionState)
}
