package doccompiler

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
)

type check func(data *docData) error

var checks = []check{
	checkAllActionsDocumented,
	checkAllConfigFieldsDocumented,
	checkAllActionsInGroup,
	checkAllActionTags,
}

func checkAll(data *docData) error {
	var findings error
	for _, check := range checks {
		findings = errors.Join(findings, check(data))
	}
	return findings
}

func checkAllActionsDocumented(data *docData) error {
	var findings error
	allActions := common.ActionStrings()
	for _, a := range allActions {
		docEntry, ok := data.ActionMap[a]
		if !ok {
			findings = errors.Join(findings, fmt.Errorf(`action "%s" is not documented`, a))
			continue
		}

		if docEntry.Description == "" {
			findings = errors.Join(findings, fmt.Errorf(`action "%s" has no description`, a))
		}
		if docEntry.Examples == "" {
			findings = errors.Join(findings, fmt.Errorf(`action "%s" has no examples`, a))
		}
	}
	return findings
}

func checkAllConfigFieldsDocumented(data *docData) error {
	var findings error
	// Get all config fields
	expectedConfigFields, err := common.FieldsString()
	if err != nil {
		common.Exit(err, ExitCodeFailedConfigFields)
	}
	// Add documentation wrapping entire document as "main" entry into config map
	expectedConfigFields = append(expectedConfigFields, "main")
	for _, field := range expectedConfigFields {
		docEntry, ok := data.ConfigMap[field]
		if !ok {
			findings = errors.Join(findings, fmt.Errorf(`config field "%s" is not documented`, field))
			continue
		}

		if docEntry.Description == "" {
			findings = errors.Join(findings, fmt.Errorf(`config field "%s" has no description`, field))
		}
		if docEntry.Examples == "" {
			findings = errors.Join(findings, fmt.Errorf(`config field "%s" has no examples`, field))
		}
	}

	return findings
}

func checkAllActionsInGroup(data *docData) error {
	var findings error

	// map actions to groups
	actionToGroups := map[string][]string{}
	for _, action := range data.Actions {
		actionToGroups[action] = []string{}
	}
	for _, group := range data.Groups {
		for _, action := range group.Actions {
			actionToGroups[action] = append(actionToGroups[action], group.Name)
		}
	}

	// check action belongs to one and only one group
	for action, groups := range actionToGroups {
		lenGroups := len(groups)
		switch {
		case lenGroups == 0:
			findings = errors.Join(findings, fmt.Errorf(`action "%s" does not belong to a group`, action))
		case lenGroups > 1:
			findings = errors.Join(findings, fmt.Errorf(`action "%s" belong to %d groups %v`, action, lenGroups, groups))
		}
	}

	return findings
}

func checkAllActionTags(data *docData) error {
	actionSettings := common.Actions()
	var tagErrors error
	for _, action := range data.Actions {
		actionParams, exists := actionSettings[action]
		if !exists {
			tagErrors = errors.Join(tagErrors, fmt.Errorf("action<%s> couldn't be found in action list", action))
			continue
		}

		if err := checkActionTags(action, reflect.ValueOf(actionParams), data.ParamMap, 0); err != nil {
			tagErrors = errors.Join(tagErrors, err)
		}
	}
	return tagErrors
}

func checkActionTags(action string, value reflect.Value, paramDocs map[string][]string, level int) error {
	switch value.Kind() {
	case reflect.Ptr:
		elem := value.Elem()
		if value.IsNil() && value.CanInterface() {
			elem = reflect.New(value.Type().Elem())
		}
		if !elem.IsValid() {
			return nil
		}
		return checkActionTags(action, elem, paramDocs, level+1)
	case reflect.Interface:
		elem := value.Elem()

		if !elem.IsValid() {
			return nil
		}
		return checkActionTags(action, elem, paramDocs, level+1)
	case reflect.Struct:
		if level > 20 {
			return fmt.Errorf("action<%s> recursive generation of parameter docs: add \"recursive\" to struct tag `doc-key:\"a.doc.key,recursive\"` of recursive struct member", action)
		}
		var findings error
	fieldLoop:
		for i := range value.NumField() {
			field := reflect.Indirect(value).Type().Field(i)
			if _, ignore := common.JsonTagName(field.Tag); ignore { // should not be checked
				continue fieldLoop
			}

			// stop if recursive data type
			docKeyTag := field.Tag.Get("doc-key")
			for _, flag := range strings.Split(docKeyTag, ",")[1:] {
				if strings.TrimSpace(flag) == "recursive" {
					continue fieldLoop
				}
			}
			// Template is recursive
			if value.Type().String() == "synced.Template" {
				continue fieldLoop
			}

			if value.Field(i).CanInterface() {
				if err := checkFieldTags(action, field, paramDocs); err != nil {
					findings = errors.Join(findings, err)
				}
			}

			if err := checkActionTags(action, value.Field(i), paramDocs, level+1); err != nil {
				findings = errors.Join(findings, err)
			}
		}
		return findings
	case reflect.Array, reflect.Slice:
		if value.CanInterface() {
			if err := checkActionTags(action, reflect.New(value.Type().Elem()), paramDocs, level+1); err != nil {
				return err
			}
		}

	// This Kinds are "end-of-line"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Chan, reflect.Complex64, reflect.Complex128,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Map, reflect.Func:
	default:
		// Default with error on following cases, if ever needed we need to add special handling of them
		// Uintptr
		// UnsafePointer
		return fmt.Errorf("action<%s> kind<%v> not supported", action, value.Kind())
	}
	return nil
}

// checkFieldTags
func checkFieldTags(action string, field reflect.StructField, fieldDocs map[string][]string) error {
	if field.Anonymous {
		return nil
	}

	if _, ignore := common.JsonTagName(field.Tag); ignore {
		return nil
	}

	docKey := strings.Split(field.Tag.Get("doc-key"), ",")[0]
	if docKey == "-" {
		return nil
	}

	if docKey == "" {
		return fmt.Errorf("action<%s> field<%s> does not have a doc-key", action, field.Name)
	}

	params, ok := fieldDocs[docKey]
	if !ok || len(params) < 1 {
		return fmt.Errorf("action<%s> field<%s> doc-key<%s> not found in params.json", action, field.Name, docKey)
	}

	displayName := field.Tag.Get("displayname")
	if displayName == "-" {
		return nil
	}

	if displayName == "" {
		fmt.Printf("WARNING: action<%s> field<%s> does not have a displayName.\n", action, field.Name)
	}

	return nil
}
