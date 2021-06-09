package common

import (
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/qlik-oss/gopherciser/scenario"
)

type (
	// DocEntry contains strings and examples
	DocEntry struct {
		// Description Information preceding parameters
		Description string
		// Examples Information subsequent to parameters
		Examples string
	}
	// GroupsEntry definition of group of actions
	GroupsEntry struct {
		// Name of the group
		Name string
		// Title of the group (as used in documentation
		Title string
		// Actions contained in the group
		Actions []string
		DocEntry
	}
)

// Shared global variables for compile and generate documentation
var (
	// IgnoreActions list of "helper" actions to be ignored for documentation
	IgnoreActions = []string{"connectws"}

	// internal global variables
	emptyConfig  *config.Config
	configFields map[string]interface{}
)

// ActionStrings all registered actions
func ActionStrings() []string {
	return sortedKeys(Actions())
}

// Actions all registered actions with action settings
func Actions() map[string]interface{} {
	actions := scenario.RegisteredActions()
	actionMap := make(map[string]interface{}, len(actions))

	// fill all actions to map with action settings struct
	for _, action := range actions {
		actionMap[action] = scenario.NewActionsSettings(action)
	}

	// remove helper actions
	for _, ignore := range IgnoreActions {
		delete(actionMap, ignore)
	}

	return actionMap
}

// FieldsString config fields sections
func FieldsString() ([]string, error) {
	fields, err := Fields()
	if err != nil {
		return nil, err
	}
	return sortedKeys(fields), nil
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Fields config fields sections with objects
func Fields() (map[string]interface{}, error) {
	if configFields != nil {
		return configFields, nil
	}

	if emptyConfig == nil {
		var err error
		emptyConfig, err = config.NewEmptyConfig()
		if err != nil {
			return nil, err
		}
	}

	cfgValue := reflect.Indirect(reflect.ValueOf(emptyConfig))
	configFields := make(map[string]interface{}, 6)
	for i := 0; i < cfgValue.NumField(); i++ {
		handleConfigValue(cfgValue.Type().Field(i), cfgValue.Field(i), configFields)
	}

	return configFields, nil
}

func handleConfigValue(field reflect.StructField, value reflect.Value, configFields map[string]interface{}) {
	value = reflect.Indirect(value)
	if field.Anonymous {
		if value.Kind() == reflect.Struct {
			for i := 0; i < value.NumField(); i++ {
				handleConfigValue(value.Type().Field(i), value.Field(i), configFields)
			}
		}
	}

	if !value.CanSet() {
		return // not exported value
	}

	jsonTag := field.Tag.Get("json")
	if len(jsonTag) > 0 {
		jsonTag = strings.Split(jsonTag, ",")[0]
		if jsonTag == "-" {
			return // field marked to be skipped
		}
	} else {
		jsonTag = field.Name
	}

	// todo check duplicate
	configFields[jsonTag] = value.Interface()
}

// ReadFile into memory
func ReadFile(path string) ([]byte, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("failed to open file<%s>: %v\n", path, err)
	}
	return file, nil
}

// Exit prints errors message and exits program with code
func Exit(err error, code int) {
	_, _ = os.Stderr.WriteString(err.Error())
	os.Exit(code)
}

// Keys returns a sorted slice of map keys
func Keys(m map[string]DocEntry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
