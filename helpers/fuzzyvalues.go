package helpers

import (
	"strconv"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
)

type (
	// FuzzyBool resolves boolean sent as strings, integers, float64 or boolean, json unmarshal defaults to true
	FuzzyBool bool

	// FuzzyInt resolves integer sent as string or integer
	FuzzyInt int
)

// UnmarshalJSON FuzzyBool
func (fb *FuzzyBool) UnmarshalJSON(arg []byte) error {
	if fb == nil {
		return nil
	}

	if len(arg) < 1 {
		*fb = true
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
			*fb = false
		default:
			*fb = true
		}
	case float64:
		*fb = !FuzzyBool(NearlyEqual(val, 0.0))
	case string:
		switch val {
		case "false", "0":
			*fb = false
		default:
			*fb = true
		}
	case bool:
		*fb = FuzzyBool(val)
	case nil:
		*fb = true
	default:
		return errors.Errorf("Failed to unmarshal byte array<%v> as bool", arg)
	}

	return nil
}

// AsBool returns bool representation of StringBool
func (fb *FuzzyBool) AsBool() bool {
	if fb == nil {
		return false
	}

	return bool(*fb)
}

func (fi *FuzzyInt) UnmarshalJSON(arg []byte) error {
	if fi == nil {
		return nil
	}

	if len(arg) < 1 {
		*fi = 0
		return nil
	}

	var val interface{}
	if err := json.Unmarshal(arg, &val); err != nil {
		return errors.Wrapf(err, "Failed to unmarshal <%v> as integer", arg)
	}

	switch val := val.(type) {
	case int:
		*fi = FuzzyInt(val)
	case string:
		if val == "" {
			*fi = 0
		} else {
			ival, err := strconv.Atoi(val)
			if err != nil {
				return errors.Wrapf(err, "failed to convert string<%s> to integer", val)
			}
			*fi = FuzzyInt(ival)
		}
	case float64:
		*fi = FuzzyInt(val)
	default:
		return errors.Errorf("Failed to unmarshal value<%v> with type<%T> as integer", val, val)
	}
	return nil
}
