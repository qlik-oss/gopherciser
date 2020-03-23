package helpers

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
)

func TestFlatten(t *testing.T) {
	var mErr *multierror.Error

	if err := FlattenMultiError(mErr); err != nil {
		t.Errorf("Expected<nil> got: %v", err)
	}

	firstError := "first error"
	mErr = multierror.Append(mErr, fmt.Errorf(firstError))

	err := FlattenMultiError(mErr)
	if err == nil || err.Error() != firstError {
		t.Errorf("Expected<%s> got: %v", firstError, err)
	}

	secondError := "second error"
	mErr = multierror.Append(mErr, fmt.Errorf(secondError))
	err = FlattenMultiError(mErr)
	expected := fmt.Sprintf("2 errors occurred:\n\t* %s\n\t* %s\n\n", firstError, secondError)
	if err == nil || err.Error() != expected {
		t.Errorf("Expected:\n<%s>\ngot:\n<%s>\n", expected, err.Error())
	}
}
