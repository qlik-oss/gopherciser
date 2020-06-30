package helpers

import (
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// FlattenMultiError return nil or original error if multierror does not have more than one error
func FlattenMultiError(mErr *multierror.Error) error {
	if mErr == nil {
		return nil
	}

	if len(mErr.Errors) == 0 {
		return nil
	}

	if len(mErr.Errors) == 1 {
		return mErr.Errors[0]
	}

	return mErr
}

// TrueCause of error, in the case of a multi error, the first error in list will be used
func TrueCause(err error) error {
	switch err.(type) {
	case nil:
		return nil
	case *multierror.Error:
		switch err.(*multierror.Error).Len() {
		case 0:
			return nil
		default:
			return TrueCause(err.(*multierror.Error).Errors[0])
		}
	default:
		return errors.Cause(err)
	}
}
