package scenario

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ClearFieldSettings clear all selections in field
	ClearFieldSettings struct {
		Name string `json:"name" displayname:"name" doc-key:"clearfield.name"` // TODO add appstructure:"fields:name" when supported by GUI
	}
)

// Validate ClearField action settings
func (settings ClearFieldSettings) Validate() ([]string, error) {
	if settings.Name == "" {
		return nil, errors.New("no name defined for clear field action")
	}
	return nil, nil
}

// Execute ClearField action
func (settings ClearFieldSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment"))
		return
	}

	app := sessionState.Connection.Sense().CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense app"))
		return
	}

	fieldName := settings.Name
	sessionState.QueueRequest(func(ctx context.Context) error {
		field, err := app.Doc.GetField(ctx, settings.Name, "")
		if err != nil {
			return errors.Wrapf(err, "error getting field<%s>", fieldName)
		}

		success, err := field.Clear(ctx)
		if err != nil {
			return errors.Wrapf(err, "error clearing field<%s>", fieldName)
		}
		if !success {
			return errors.Errorf("field<%s> was not cleared", fieldName)
		}

		return nil
	}, actionState, true, "")

	sessionState.Wait(actionState)
}
