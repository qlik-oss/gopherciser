package scenario

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// CycleDimensionSettings Cycledimension Cycle a step in a cyclic dimension
	CycleDimensionSettings struct {
		Id string `json:"id" doc-key:"cycledimension.id"`
	}
)

// Validate CycledimensionSettings action (Implements ActionSettings interface)
func (settings CycleDimensionSettings) Validate() ([]string, error) {
	if settings.Id == "" {
		return nil, errors.Errorf("Id not set for %s", ActionCycleDimension)
	}
	return nil, nil
}

// Execute CycledimensionSettings action (Implements ActionSettings interface)
func (settings CycleDimensionSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment"))
		return
	}

	app := sessionState.Connection.Sense().CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense app"))
		return
	}

	dim, err := app.GetDimension(sessionState, actionState, settings.Id)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	err = sessionState.SendRequest(actionState, func(ctx context.Context) error {
		return dim.StepCycle(ctx, 1)
	})
	if err != nil {
		actionState.AddErrors(err)
	}

	sessionState.Wait(actionState) // Await all async requests, e.g. those triggered on changed objects
}
