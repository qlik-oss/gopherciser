package common

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/scheduler"
)

type (
	// DocEntry contains strings and examples
	DocEntry struct {
		// Description Information preceding parameters
		Description string
		// Examples Information subsequent to parameters
		Examples string
	}

	GroupsEntries []GroupsEntry

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

// Implements Sort interface
func (entries GroupsEntries) Len() int           { return len(entries) }
func (entries GroupsEntries) Less(i, j int) bool { return entries[i].Name < entries[j].Name }
func (entries GroupsEntries) Swap(i, j int)      { entries[i], entries[j] = entries[j], entries[i] }

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

func Schedulers() map[string]interface{} {
	scheds := scheduler.RegisteredSchedulers()
	schedulerMap := make(map[string]interface{}, len(scheds))

	for _, sched := range scheds {
		schedulerMap[sched] = scheduler.SchedHandler(sched)
	}

	return schedulerMap
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
		emptyConfig.Scheduler = DummyScheduler{}
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

	jsonTag, ignore := JsonTagName(field.Tag)
	if ignore {
		return // field marked to be skipped
	}
	if jsonTag == "" {
		jsonTag = field.Name
	}

	// todo check duplicate
	configFields[jsonTag] = value.Interface()
}

func JsonTagName(tag reflect.StructTag) (string, bool) {
	jsonTag := tag.Get("json")
	if len(jsonTag) > 0 {
		jsonTag = strings.Split(jsonTag, ",")[0]
		if jsonTag == "-" {
			return "", true // field marked to be skipped
		}
		return jsonTag, false
	}
	return "", false
}

// ReadFile into memory
func ReadFile(path string) ([]byte, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("failed to open file<%s>: %v\n", path, err)
	}
	return file, nil
}

// Exit prints errors message and exits program with code
func Exit(err error, code int) {
	_, _ = fmt.Fprintf(os.Stderr, "%v\n", err.Error())
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
