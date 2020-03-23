package senseobjects

import (
	"context"

	"github.com/qlik-oss/gopherciser/action"
)

type (
	// SessionState temporary session interface to avoid circular dependencies
	SessionState interface {
		BaseContext() context.Context
		QueueRequest(f func(ctx context.Context) error, actionState *action.State, failOnError bool, errMsg string)
		SendRequest(actionState *action.State, f func(ctx context.Context) error) error
		RegisterEvent(handle int,
			onEvent func(ctx context.Context, actionState *action.State) error,
			onClose func(),
			failOnError bool)
		DeRegisterEvent(handle int)
	}
)
