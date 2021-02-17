package helpers

import (
	"strings"

	"github.com/pkg/errors"
)

type (
	// StringBool resolves boolean sent as strings, json unmarshal defaults to true
	StringBool bool
)

// UnmarshalJSON StringBool
func (sb *StringBool) UnmarshalJSON(arg []byte) error {
	var s string
	if err := jsonit.Unmarshal(arg, &s); err != nil {
		return errors.Wrap(err, "failed to unmarshal StringBool")
	}

	switch strings.ToLower(s) {
	case "false":
	case "0":
		*sb = false
	default:
		*sb = true
	}
	return nil
}
