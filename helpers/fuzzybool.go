package helpers

import (
	"encoding/json"

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

	tmpMap := make(map[string]interface{})
	tmpArray := []byte(`{"val":`)
	tmpArray = append(tmpArray, arg...)
	tmpArray = append(tmpArray, byte('}'))

	if err := json.Unmarshal(tmpArray, &tmpMap); err != nil {
		return errors.Wrapf(err, "Failed to unmarshal byte array<%v> as bool", arg)
	}

	if tmpMap["val"] == nil {
		return errors.Errorf("Failed to unmarshal byte array<%v> as bool", arg)
	}

	switch val := tmpMap["val"].(type) {
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
