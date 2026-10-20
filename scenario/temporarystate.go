package scenario

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

const (
	ActionStoreTemporaryState   = "storetemporarystate"
	ActionRestoreTemporaryState = "restoretemporarystate"
)

// actions used as sub actions for clicking a chained app button transfering selection state between apps
type (
	storeTemporaryStateSettings struct {
		id *string
	}
	restoreTemporaryStateSettings struct {
		id string
	}
)

// Validate TemporaryStateSettings action (Implements ActionSettings interface)
func (settings storeTemporaryStateSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute TemporaryStateSettings action (Implements ActionSettings interface)
func (settings storeTemporaryStateSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {

	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		// Save temp selection state
		id, success, err := sessionState.Connection.Sense().CurrentApp.Doc.StoreTempSelectionState(ctx, 0)
		if err != nil {
			return errors.Wrap(err, "failed to store temp selection state")
		}
		if !success {
			return errors.New("store temp selection state unsuccessful")
		}
		*settings.id = id

		return nil
	}); err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	sessionState.Wait(actionState) // Await all async requests, e.g. those triggered on changed objects
}

// Validate TemporaryStateSettings action (Implements ActionSettings interface)
func (settings restoreTemporaryStateSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute TemporaryStateSettings action (Implements ActionSettings interface)
func (settings restoreTemporaryStateSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {

	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		// Restore temp selection state
		success, err := sessionState.Connection.Sense().CurrentApp.Doc.RestoreTempSelectionState(ctx, settings.id)
		if err != nil {
			return errors.Wrap(err, "failed to restore temp selection state")
		}
		if !success {
			return errors.New("restore temp selection state unsuccessful")
		}
		return nil
	}); err != nil {
		actionState.AddErrors(errors.WithStack(err))
	}

	sessionState.Wait(actionState) // Await all async requests, e.g. those triggered on changed objects
}
