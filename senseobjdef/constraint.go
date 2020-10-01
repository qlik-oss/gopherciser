package senseobjdef

import (
	"encoding/json"
	"fmt"
	"github.com/qlik-oss/gopherciser/helpers"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enummap"
)

type (
	// ConstraintValue value to evaluate. First character needs to be one of [<,>,=,!],
	// followed by a number or one of the words [true,false]
	ConstraintValue string
	// EmptyConstraintValueError ConstraintValue is empty
	EmptyConstraintValueError struct{}
	// MalformedConstraintValueError ConstraintValue is malformed
	MalformedConstraintValueError ConstraintValue

	// Constraint defining if to send get data requests
	Constraint struct {
		// Path to value to evaluate
		Path helpers.DataPath `json:"path"`
		// Value constraint definition, first character must be <,>,= or !
		// followed by number or the words true/false
		Value ConstraintValue `json:"value"`
		// Required require constraint to evaluate, error if evaluation fails
		Required bool `json:"required,omitempty"`

		validate sync.Once

		operator constraintOperator
		value    string
	}

	constraintOperator int
)

const (
	unknownOperator constraintOperator = iota
	lessThanOperator
	largerThanOperator
	equalOperator
	notOperator
	containsOperator
)

var (
	constraintOperatorEnum = enummap.NewEnumMapOrPanic(map[string]int{
		"<": int(lessThanOperator),
		">": int(largerThanOperator),
		"=": int(equalOperator),
		"!": int(notOperator),
		"~": int(containsOperator),
	})
)

//Error constraint value is empty
func (err EmptyConstraintValueError) Error() string {
	return "constraint value is empty"
}

//Error ConstraintValue is malformed
func (err MalformedConstraintValueError) Error() string {
	return fmt.Sprintf("constraint value<%s> is malformed", string(err))
}

// Validate constraint
func (constraint *Constraint) Validate() error {
	var validateErr error
	constraint.validate.Do(func() {
		if string(constraint.Value) == "" {
			validateErr = EmptyConstraintValueError{}
			return
		}

		runes := []rune(string(constraint.Value))

		if len(runes) < 2 {
			validateErr = MalformedConstraintValueError(constraint.Value)
			return
		}

		operator, err := constraintOperatorEnum.Int(string(runes[0]))
		if err != nil {
			validateErr = MalformedConstraintValueError(constraint.Value)
			return
		}

		//TODO validate constraint.Path

		constraint.operator = constraintOperator(operator)
		constraint.value = string(runes[1:])
	})

	return errors.WithStack(validateErr)
}

//Evaluate constraint value in data
func (constraint *Constraint) Evaluate(data json.RawMessage) (bool, error) {
	if err := constraint.Validate(); err != nil {
		return false, errors.WithStack(err)
	}

	rawValue, err := constraint.Path.Lookup(data)
	if err != nil {
		err = errors.Wrapf(err, "error evaluating constraint<%s> in data path<%s>",
			string(constraint.Value), string(constraint.Path))

		switch errors.Cause(err).(type) {
		case helpers.NoDataFound:
			if !constraint.Required {
				err = nil
			}
		}

		return false, err
	}

	var value interface{}
	if err = jsonit.Unmarshal(rawValue, &value); err != nil {
		return false, errors.Wrapf(err, "error unmarshaling value in path<%s>", string(constraint.Path))
	}

	switch value.(type) {
	case float64:
		floatValue, errParse := strconv.ParseFloat(constraint.value, 64)
		if errParse != nil {
			return false, errors.Wrapf(errParse, "error parsing constraint value as float64")
		}
		return constraint.operator.evalFloat64(value.(float64), floatValue)
	case bool:
		boolValue, errParse := strconv.ParseBool(constraint.value)
		if errParse != nil {
			return false, errors.Wrapf(errParse, "error parsing constraint value as bool")
		}
		return constraint.operator.evalBool(value.(bool), boolValue)
	case string:
		return constraint.operator.evalString(value.(string), constraint.value)
	case []interface{}:
		return constraint.operator.evalArray(value.([]interface{}), constraint.value)
	default:
		return false, errors.Errorf("value type<%T> not supported", value)
	}
}

func (operator constraintOperator) evalFloat64(val float64, constraint float64) (bool, error) {
	switch operator {
	case lessThanOperator:
		return val < constraint, nil
	case largerThanOperator:
		return val > constraint, nil
	case equalOperator:
		return val > constraint-0.0000000000001 && val < constraint+0.0000000000001, nil
	case notOperator:
		return val < constraint-0.0000000000001 || val > constraint+0.0000000000001, nil
	case unknownOperator:
		fallthrough // use unknownOperator somewhere to avoid lint warning...
	default:
		str, _ := constraintOperatorEnum.String(int(operator))
		if str == "" {
			str = strconv.Itoa(int(operator))
		}
		return false, errors.Errorf("operator<%s> not supported for float64 constraint evaluation", str)
	}
}

func (operator constraintOperator) evalBool(val bool, constraint bool) (bool, error) {
	switch operator {
	case equalOperator:
		return val == constraint, nil
	case notOperator:
		return val != constraint, nil
	default:
		str, _ := constraintOperatorEnum.String(int(operator))
		if str == "" {
			str = strconv.Itoa(int(operator))
		}
		return false, errors.Errorf("operator<%s> not supported for bool constraint evaluation", str)
	}
}

func (operator constraintOperator) evalString(val string, constraint string) (bool, error) {
	switch operator {
	case equalOperator:
		return val == constraint, nil
	case notOperator:
		return val != constraint, nil
	default:
		str, _ := constraintOperatorEnum.String(int(operator))
		if str == "" {
			str = strconv.Itoa(int(operator))
		}
		return false, errors.Errorf("operator<%s> not supported for string constraint evaluation", str)
	}
}

func (operator constraintOperator) evalArray(val []interface{}, constraint string) (bool, error) {
	boolValue, errParseBool := strconv.ParseBool(constraint)
	floatValue, errParseFloat := strconv.ParseFloat(constraint, 64)

	switch operator {
	case containsOperator:
		for _, v := range val {
			switch v.(type) {
			case float64:
				if errParseFloat != nil {
					return false, errors.Wrapf(errParseFloat, "error parsing constraint value as float64")
				}
				if v.(float64) > floatValue-0.0000000000001 && v.(float64) < floatValue+0.0000000000001 {
					return true, nil
				}
			case bool:
				if errParseBool != nil {
					return false, errors.Wrapf(errParseBool, "error parsing constraint value as bool")
				}
				if v.(bool) == boolValue {
					return true, nil
				}
				return operator.evalBool(v.(bool), boolValue)
			case string:
				if v.(string) == constraint {
					return true, nil
				}
			default:
				return false, errors.Errorf("array element value type<%T> not supported", v)
			}
		}
		return false, nil
	default:
		str, _ := constraintOperatorEnum.String(int(operator))
		if str == "" {
			str = strconv.Itoa(int(operator))
		}
		return false, errors.Errorf("operator<%s> not supported for array constraint evaluation", str)
	}
}
