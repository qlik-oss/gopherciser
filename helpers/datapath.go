package helpers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

type (
	//DataPath string which can be return as substring splitted on / (slash)
	DataPath string
	//NoStepsError no steps to take in data path
	NoStepsError string
	//NoDataFound no data found in path
	NoDataFound string
)

//Error no steps in data path
func (err NoStepsError) Error() string {
	return fmt.Sprintf("No steps to take in datapath<%s>", string(err))
}

//Error no data found in data path
func (err NoDataFound) Error() string {
	return fmt.Sprintf("No data found in datapath<%s>", string(err))
}

//NewDataPath new instance
func NewDataPath(path string) DataPath {
	return DataPath(path)
}

//String string representation of datapath
func (path *DataPath) String() string {
	return string(*path)
}

// steps path substrings splitted on / (slash)
func (path *DataPath) steps() []string {
	return strings.Split(strings.Trim(path.String(), "/"), "/")
}

// Lookup object in path, if data found is of type string it will be quoted with ""
func (path DataPath) Lookup(data json.RawMessage) (json.RawMessage, error) {
	return path.lookup(data, true)
}

// LookupNoQuotes object in path, data of type string will not be quoted
func (path DataPath) LookupNoQuotes(data json.RawMessage) (json.RawMessage, error) {
	return path.lookup(data, false)
}

func (path DataPath) lookup(data json.RawMessage, quoteString bool) (json.RawMessage, error) {
	steps := path.steps()

	if steps == nil || len(steps) < 1 {
		return nil, errors.WithStack(NoStepsError(string(path)))
	}

	v, t, _, err := jsonparser.Get([]byte(data), steps...)
	if err != nil {
		if t == jsonparser.NotExist {
			return nil, errors.WithStack(NoDataFound(path.String()))
		}
		return nil, errors.Wrapf(err, "failed parsing path<%s>", path.String())
	}
	switch t {
	case jsonparser.String:
		if quoteString {
			return json.RawMessage(fmt.Sprintf(`"%s"`, v)), nil
		}
		return json.RawMessage(string(v)), nil
	default:
		return v, nil
	}
}

// LookupAndSet look object in path and set to new object
func (path DataPath) LookupAndSet(data []byte, newValue []byte) ([]byte, error) {
	steps := path.steps()

	if steps == nil || len(steps) < 1 {
		return data, errors.WithStack(NoStepsError(string(path)))
	}

	var err error
	data, err = jsonparser.Set([]byte(data), newValue, steps...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to set data<%s> at path<%s>", newValue, path)
	}

	return data, nil
}

// LookupMulti objects with subpaths under an array in a path
func (path DataPath) LookupMulti(data json.RawMessage, separator string) ([]json.RawMessage, error) {
	// todo change to use jsonparser
	if separator == "" || !strings.Contains(string(path), separator) {
		raw, err := path.Lookup(data)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return []json.RawMessage{raw}, nil
	}

	sa := strings.Split(string(path), separator)
	if len(sa) > 2 {
		return nil, errors.Errorf("Datapath only supports one instance of separator")
	}

	arrayPath := DataPath(sa[0])

	jsonArray, err := arrayPath.Lookup(data)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var jMap interface{}

	if err := jsonit.Unmarshal(jsonArray, &jMap); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal data")
	}

	var dataArray []interface{}
	switch jMap.(type) {
	case []interface{}:
		dataArray = jMap.([]interface{})
	default:
		return nil, errors.Errorf("Expected array at path<%s> but was type<%T>", sa[0], jMap)
	}

	rawArray := make([]json.RawMessage, len(dataArray))

	hasSubPath := len(sa) > 1

	for i, v := range dataArray {
		raw, marshalErr := jsonit.Marshal(v)
		if marshalErr != nil {
			return nil, errors.Wrapf(marshalErr, "Failed to marshal array entry<%d> in path<%s>", i, sa[0])
		}

		if hasSubPath {
			subpath := DataPath(sa[1])
			raw, err = subpath.Lookup(raw)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to lookup subpath<%s> of path<%s>", sa[1], sa[0])
			}
		}

		rawArray[i] = raw
	}

	return rawArray, nil
}

//Contains Check if path contains value
func (path DataPath) Contains(val string) bool {
	str := string(path)
	if str == "" || val == "" {
		return false
	}

	return strings.Contains(str, val)
}
