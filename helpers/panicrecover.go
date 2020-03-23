package helpers

import (
	"github.com/pkg/errors"
)

// RecoverWithError recovers from panic and sets error to panicErr
func RecoverWithError(panicErr *error) {
	if r := recover(); r != nil {
		var ok bool
		var err error

		if err, ok = r.(error); !ok {
			err = errors.Errorf("PANIC: %v", r)
		}
		if panicErr != nil {
			*panicErr = err
		}
	}
}
