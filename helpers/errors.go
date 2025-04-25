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
	switch e := err.(type) {
	case nil:
		return nil
	case *multierror.Error:
		switch e.Len() {
		case 0:
			return nil
		default:
			return TrueCause(e.Errors[0])
		}
	default:
		cause := errors.Cause(err)
		switch cause.(type) {
		case *multierror.Error:
			return TrueCause(cause)
		default:
			return cause
		}
	}
}

// RankedCause helper to filter through hierarchy of errors to find most suitable cause to expose
// ranker: should return the most important cause as the higher number
func RankedCause(err error, ranker func(error) int) (int, error) {
	err = errors.Cause(err)
	switch err := err.(type) {
	case *multierror.Error:
		var rankedErr error
		rank := -1 // -1 makes sure unranked gets higher rank
		for _, e := range err.Errors {
			r, cause := RankedCause(e, ranker)
			if r > rank {
				rank = r
				rankedErr = cause
			}
		}
		return rank, rankedErr
	default:
		return ranker(err), err
	}
}

// MultiErrorAppend helper for test cases to append list of errors to multi error without direct import of multi error package
func MultiErrorAppend(baseErr error, appendErr ...error) error {
	return multierror.Append(baseErr, appendErr...)
}
