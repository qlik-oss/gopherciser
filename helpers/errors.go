package helpers

import multierror "github.com/hashicorp/go-multierror"

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
