package helpers

import (
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
)

type (
	// StringBool resolves boolean sent as strings, json unmarshal defaults to true
	StringBool bool
)

// UnmarshalJSON StringBool
func (sb *StringBool) UnmarshalJSON(arg []byte) error {
	if sb == nil {
		return nil
	}

	// Is integer
	var i int
	if err := json.Unmarshal(arg, &i); err == nil {
		switch i {
		case 0:
			*sb = false
		default:
			*sb = true
		}
		return nil
	}

	// Is string
	var s string
	if err := json.Unmarshal(arg, &s); err != nil {
		return errors.Wrapf(err, "Failed to unmarshal byte array<%v>", arg)
	}

	switch s {
	case "false", "0":
		*sb = false
	default:
		*sb = true
	}

	return nil
}

// AsBool returns bool representation of StringBool
func (sb *StringBool) AsBool() bool {
	if sb == nil {
		return false
	}

	return bool(*sb)
}
