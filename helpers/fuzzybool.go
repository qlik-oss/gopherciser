package helpers

import (
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
)

type (
	// FuzzyBool resolves boolean sent as strings, integers, float64 or boolean, json unmarshal defaults to true
	FuzzyBool bool
)

// UnmarshalJSON FuzzyBool
func (sb *FuzzyBool) UnmarshalJSON(arg []byte) error {
	if sb == nil {
		return nil
	}

	if len(arg) < 1 {
		*sb = true
		return nil
	}

	var val interface{}
	if err := json.Unmarshal(arg, &val); err != nil {
		return errors.Wrapf(err, "Failed to unmarshal byte array<%v> as bool", arg)
	}

	switch val := val.(type) {
	case int:
		switch val {
		case 0:
			*sb = false
		default:
			*sb = true
		}
	case float64:
		*sb = !FuzzyBool(NearlyEqual(val, 0.0))
	case string:
		switch val {
		case "false", "0":
			*sb = false
		default:
			*sb = true
		}
	case bool:
		*sb = FuzzyBool(val)
	case nil:
		*sb = true
	default:
		return errors.Errorf("Failed to unmarshal byte array<%v> as bool", arg)
	}

	return nil
}

// AsBool returns bool representation of StringBool
func (sb *FuzzyBool) AsBool() bool {
	if sb == nil {
		return false
	}

	return bool(*sb)
}
