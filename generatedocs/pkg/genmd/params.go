package genmd

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
)

const (
	defaultIndent = "  "
)

func MarkdownParams(obj interface{}) string {
	buf := bytes.NewBuffer(nil)
	if err := handleValue(reflect.ValueOf(obj), buf, ""); err != nil {
		common.Exit(err, ExitCodeFailedHandleParams)
	}
	return buf.String()
}

func handleValue(value reflect.Value, buf *bytes.Buffer, indent string) error {
	switch value.Kind() {
	case reflect.Ptr:
		elem := value.Elem()
		if value.IsNil() && value.CanInterface() {
			elem = reflect.New(value.Type().Elem())
		}
		if !elem.IsValid() {
			return nil
		}
		return errors.WithStack(handleValue(elem, buf, indent))
	case reflect.Interface:
		elem := value.Elem()

		if !elem.IsValid() {
			return nil
		}
		return errors.WithStack(handleValue(elem, buf, indent))
	case reflect.Struct:
		if len(indent) > 20*len(defaultIndent) {
			return errors.Errorf("Error: recursive generation of paramater docs: add struct tag `rec:\"true\"` to recursive struct member ")
		}
		for i := 0; i < value.NumField(); i++ {
			field := reflect.Indirect(value).Type().Field(i)
			// stop on recursive data type
			if field.Tag.Get("rec") != "" {
				continue
			}
			// Template is recursive
			if value.Type().String() == "session.SyncedTemplate" {
				continue
			}

			if value.Field(i).CanInterface() {
				handleFields(field, buf, indent)
			}
			innerIndent := indent
			if !field.Anonymous {
				innerIndent += defaultIndent
			}

			if err := handleValue(value.Field(i), buf, innerIndent); err != nil {
				return errors.WithStack(err)
			}
		}
	case reflect.Array, reflect.Slice:
		if value.CanInterface() {
			if err := handleValue(reflect.New(value.Type().Elem()), buf, indent); err != nil {
				return errors.WithStack(err)
			}
		}

	// This Kinds are "end-of-line"
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Bool:
	case reflect.Chan:
	case reflect.Complex64:
	case reflect.Complex128:
	case reflect.Float32:
	case reflect.Float64:
	case reflect.String:
	case reflect.Map:
	case reflect.Func:

	default:
		// Default with error on following cases, if ever needed we need to add special handling of them
		// Uintptr
		// UnsafePointer
		return errors.Errorf("Kind<%v> not supported:", value.Kind())
	}
	return nil
}

// handleFields
func handleFields(field reflect.StructField, buf *bytes.Buffer, indent string) {
	if field.Anonymous {
		return
	}
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		jsonTag = field.Name
	}
	jsonTag = strings.Split(jsonTag, ",")[0]
	if jsonTag == "-" {
		return // this field should be ignored
	}

	buf.WriteString(fmt.Sprintf("%s* `%s`: ", indent, jsonTag))
	docKey := field.Tag.Get("doc-key")
	defaultString := func() {
		buf.WriteString("*Missing documentation*\n")
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Warning: parameter %s is missing documentation\n", field.Name))
	}
	if docKey != "" {
		params, ok := compiledDocsGlobal.Params[docKey]
		if !ok || len(params) < 1 {
			defaultString()
		} else {
			buf.WriteString(params[0])
			buf.WriteString("\n")
			indent += "    * "
			for i := 1; i < len(params); i++ {
				buf.WriteString(indent)
				buf.WriteString(params[i])
				buf.WriteString("\n")
			}
		}
	} else {
		defaultString()
	}
}

// UngroupedActions filter grouped actions from all actions
func UngroupedActions(groups []common.GroupsEntry) []string {
	// fill map with all actions
	actionMap := make(map[string]struct{})
	for _, action := range common.ActionStrings() {
		actionMap[action] = struct{}{}
	}

	// Remove grouped actions
	for _, group := range groups {
		for _, action := range group.Actions {
			delete(actionMap, action)
		}
	}
	actions := make([]string, 0, len(actionMap))
	for action := range actionMap {
		actions = append(actions, action)
	}
	sort.Strings(actions)
	return actions
}
