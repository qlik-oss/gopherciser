package action

import (
	"sync"

	"github.com/hashicorp/go-multierror"
	pkgerrors "github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	errors struct {
		me *multierror.Error
		mu sync.RWMutex
	}

	// State holder for an action
	State struct {
		errors
		// Failed an error occurred during execution of action
		Failed bool
		// NoResults should be reported for action
		NoResults bool
		// Details for action to log on result report
		Details string
		// NoRestartOnDisconnect in the case of using websocket reconnect logic, don't restart action when a reconnect has happened
		NoRestartOnDisconnect bool
		// FailOnDisconnect in the case of using webscoket reconnect logic, fail the action instead of trying to restart it
		FailOnDisconnect bool
	}
)

// AddErrors to error list
func (as *State) AddErrors(errs ...error) {
	as.errors.mu.Lock()
	defer as.mu.Unlock()

	as.errors.me = multierror.Append(as.errors.me, errs...)
	if as.errors.me != nil && len(as.errors.me.Errors) > 0 {
		as.Failed = true
	}
}

// NewErrorf add new action error
func (as *State) NewErrorf(format string, args ...interface{}) {
	as.AddErrors(pkgerrors.Errorf(format, args...))
}

// Errors return action error
func (as *State) Errors() error {
	if as == nil {
		return nil
	}

	as.mu.RLock()
	defer as.mu.RUnlock()

	return helpers.FlattenMultiError(as.errors.me)
}
