package helpers

import (
	"runtime/debug"

	"github.com/pkg/errors"
)

// RecoverWithError recovers from panic and sets error to panicErr
func RecoverWithError(panicErr *error) {
	if r := recover(); r != nil {
		*panicErr = errors.Errorf("PANIC: %+v Stack: %s", r, debug.Stack())
	}
}

// RecoverWithErrorFunc recovers from panic and returns panic error
func RecoverWithErrorFunc(f func()) error {
	var panicErr error
	func() {
		defer RecoverWithError(&panicErr)
		f()
	}()
	return panicErr
}
