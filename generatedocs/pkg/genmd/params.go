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

const defaultIndent = "  "

func MarkdownParams(obj interface{}, paramDocs map[string][]string) string {
	buf := bytes.NewBuffer(nil)
	if err := handleValue(reflect.ValueOf(obj), paramDocs, buf, ""); err != nil {
		common.Exit(err, ExitCodeFailedHandleParams)
	}
	return buf.String()
}

func handleValue(value reflect.Value, paramDocs map[string][]string, buf *bytes.Buffer, indent string) error {
	switch value.Kind() {
	case reflect.Ptr:
		elem := value.Elem()
		if value.IsNil() && value.CanInterface() {
			elem = reflect.New(value.Type().Elem())
		}
		if !elem.IsValid() {
			return nil
		}
		return errors.WithStack(handleValue(elem, paramDocs, buf, indent))
	case reflect.Interface:
		elem := value.Elem()

		if !elem.IsValid() {
			return nil
		}
		return errors.WithStack(handleValue(elem, paramDocs, buf, indent))
	case reflect.Struct:
		if len(indent) > 20*len(defaultIndent) {
			return errors.Errorf("Error: recursive generation of paramater docs: add \"rec\" to struct tag `doc-key:\"a.doc.key,rec\"` of recursive struct member ")
		}
	fieldLoop:
		for i := 0; i < value.NumField(); i++ {
			field := reflect.Indirect(value).Type().Field(i)

			// stop if recursive data type
			for _, flag := range strings.Split(field.Tag.Get("doc-key"), ",")[1:] {
				if strings.TrimSpace(flag) == "recursive" {
					continue fieldLoop
				}
			}
			// Template is recursive
			if value.Type().String() == "synced.Template" {
				continue fieldLoop
			}

			if value.Field(i).CanInterface() {
				handleFields(field, paramDocs, buf, indent)
			}
			innerIndent := indent
			if !field.Anonymous {
				innerIndent += defaultIndent
			}

			if err := handleValue(value.Field(i), paramDocs, buf, innerIndent); err != nil {
				return errors.WithStack(err)
			}
		}
	case reflect.Array, reflect.Slice:
		if value.CanInterface() {
			if err := handleValue(reflect.New(value.Type().Elem()), paramDocs, buf, indent); err != nil {
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
func handleFields(field reflect.StructField, fieldDocs map[string][]string, buf *bytes.Buffer, indent string) {
	if field.Anonymous {
		return
	}
	// Write jsontag to buffer
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		jsonTag = field.Name
	} else if jsonTag = strings.Split(jsonTag, ",")[0]; jsonTag == "-" {
		return // this field should be ignored
	}

	writeFieldDoc := func(docString string) {
		// Write docs to buffer
		fmt.Fprintf(buf, "%s* `%s`: %s\n", indent, jsonTag, docString)
	}

	// handle unexisting docs
	defaultString := func() {
		if unitTestMode {
			return
		}
		writeFieldDoc("*Missing documentation*")
		fmt.Fprintf(os.Stderr, "Warning: parameter %s is missing documentation\n", field.Name)
	}

	docKey := strings.Split(field.Tag.Get("doc-key"), ",")[0]
	if docKey == "" || docKey == "-" {
		defaultString()
		return
	}

	params, ok := fieldDocs[docKey]
	if !ok || len(params) < 1 {
		defaultString()
		return
	}
	// write docs
	writeFieldDoc(params[0])
	indent += "    "
	for _, param := range params[1:] {
		fmt.Fprintf(buf, "%s* %s\n", indent, param)
	}
}

// UngroupedActions filter grouped actions from all actions
func UngroupedActions(groups []common.GroupsEntry) []string {
	// fill map with all actions
	actionSet := make(map[string]struct{})
	for _, action := range common.ActionStrings() {
		actionSet[action] = struct{}{}
	}

	// Remove grouped actions
	for _, group := range groups {
		for _, action := range group.Actions {
			delete(actionSet, action)
		}
	}
	actions := make([]string, 0, len(actionSet))
	for action := range actionSet {
		actions = append(actions, action)
	}
	sort.Strings(actions)
	return actions
}
