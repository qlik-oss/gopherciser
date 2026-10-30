package tui

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/generatedocs/generated"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/scheduler"
	"github.com/qlik-oss/gopherciser/users"
)

type fieldKind int

const (
	kindText fieldKind = iota
	kindTextArea
	kindNumber
	kindBool
	kindEnum
	kindJSON // slices, maps and other values edited as compact JSON
)

type (
	// formField is one editable property, generated from a Go struct field.
	formField struct {
		path    []string
		docKey  string
		display string
		kind    fieldKind
		goKind  reflect.Kind
		options []string
		// elemType is the struct type backing a compound field, when known,
		// so it can be edited as a real sub-form instead of raw JSON.
		elemType reflect.Type

		value    string // current text as shown in the form
		orig     string // text when the form was built
		exist    bool   // property existed in the JSON when the form was built
		set      bool   // property is present, an empty set string is written as ""
		compound bool   // value is a JSON object or array, too complex for a single line

		input  textinput.Model // widget of text, number and JSON kinds
		area   textarea.Model  // widget of the textarea kind
		opts   []string        // dropdown entries of the enum kind, the first clears the property
		open   bool            // dropdown list is open
		cursor int             // highlighted dropdown entry
	}

	// typeSelector describes the "type" property of a polymorphic object whose
	// remaining properties depend on the chosen type.
	typeSelector struct {
		path    []string
		docKey  string
		options []string
		// resolve returns the struct describing the type specific properties
		// and the path they live under, nil when the type is unknown.
		resolve func(name string) (reflect.Type, []string)
	}

	// formSpec describes how to build a form for a JSON object.
	formSpec struct {
		title    string
		root     reflect.Type  // properties always present, may be nil
		selector *typeSelector // polymorphic part, may be nil
		// openActions, when set, is called instead of editing settings.actions
		// as JSON; the form is then saved and closed
		openActions func()
	}

	enumMapper interface {
		GetEnumMap() *enummap.EnumMap
	}
)

var (
	jsonUnmarshaler = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()

	// actionCoreForm holds the action properties besides type and settings.
	actionCoreForm = reflect.TypeOf(struct {
		Label    string `json:"label" displayname:"Label"`
		Disabled bool   `json:"disabled" displayname:"Disabled"`
	}{})
)

// sectionSpec returns the form specification for a top level script section,
// nil when the section has no known structure.
func sectionSpec(name string) *formSpec {
	switch name {
	case "settings":
		return &formSpec{title: "Settings", root: reflect.TypeOf(config.Settings{})}
	case "connectionSettings":
		return &formSpec{title: "Connection settings", root: reflect.TypeOf(connection.ConnectionSettings{})}
	case "hooks":
		return &formSpec{title: "Hooks", root: reflect.TypeOf(config.Hooks{})}
	case "loginSettings":
		enum := users.Type(0).GetEnumMap()
		return &formSpec{title: "Login settings", selector: &typeSelector{
			path:    []string{"type"},
			docKey:  "config.loginSettings.type",
			options: sortedKeys(enum),
			resolve: func(name string) (reflect.Type, []string) {
				i, err := enum.Int(name)
				if err != nil {
					return nil, nil
				}
				handler := users.UserGenHandler(users.Type(i))
				if handler == nil {
					return nil, nil
				}
				return reflect.TypeOf(handler), []string{"settings"}
			},
		}}
	case "scheduler":
		options := scheduler.RegisteredSchedulers()
		sort.Strings(options)
		return &formSpec{title: "Scheduler", selector: &typeSelector{
			path:    []string{"type"},
			docKey:  "config.scheduler.type",
			options: options,
			resolve: func(name string) (reflect.Type, []string) {
				handler := scheduler.SchedHandler(name)
				if handler == nil {
					return nil, nil
				}
				return reflect.TypeOf(handler), nil
			},
		}}
	}
	return nil
}

// actionSpec returns the form specification for a scenario action.
func actionSpec() *formSpec {
	options := scenario.RegisteredActions()
	sort.Strings(options)
	return &formSpec{title: "Action", root: actionCoreForm, selector: &typeSelector{
		path:    []string{"action"},
		options: options,
		resolve: func(name string) (reflect.Type, []string) {
			settings := scenario.NewActionsSettings(name)
			if settings == nil {
				return nil, nil
			}
			return reflect.TypeOf(settings), []string{"settings"}
		},
	}}
}

// fields generates the form fields for raw according to the spec.
func (spec *formSpec) fields(raw []byte) []*formField {
	var fields []*formField
	var skip []string
	if spec.selector != nil {
		skip = spec.selector.path
	}
	if spec.root != nil {
		describeStruct(spec.root, raw, nil, skip, &fields)
	}
	if sel := spec.selector; sel != nil {
		typeField := &formField{path: sel.path, docKey: sel.docKey, display: "Type", kind: kindEnum, goKind: reflect.String, options: sel.options}
		typeField.load(raw)
		fields = append([]*formField{typeField}, fields...)
		if typ, path := sel.resolve(typeField.value); typ != nil {
			describeStruct(typ, raw, path, skip, &fields)
		}
	}
	return fields
}

// validate checks that raw unmarshals into the types behind the spec.
func (spec *formSpec) validate(raw []byte) error {
	if spec.root != nil {
		if err := json.Unmarshal(raw, reflect.New(spec.root).Interface()); err != nil {
			return errors.Wrap(err, "invalid values")
		}
	}
	if sel := spec.selector; sel != nil {
		name, err := jsonparser.GetString(raw, sel.path...)
		if err != nil {
			return errors.Errorf("%s is required", strings.Join(sel.path, "."))
		}
		typ, path := sel.resolve(name)
		if typ == nil {
			return errors.Errorf("unknown %s %q", strings.Join(sel.path, "."), name)
		}
		part := raw
		if len(path) > 0 {
			part, _, _, err = jsonparser.Get(raw, path...)
			if err != nil {
				part = []byte("{}")
			}
		}
		if err := json.Unmarshal(part, reflect.New(deref(typ)).Interface()); err != nil {
			return errors.Wrapf(err, "invalid %s", name)
		}
	}
	return nil
}

func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func sortedKeys(enum *enummap.EnumMap) []string {
	keys := enum.Keys()
	sort.Strings(keys)
	return keys
}

// enumOf returns the enum map of t when it is a gopherciser enum type.
func enumOf(t reflect.Type) *enummap.EnumMap {
	if t.Kind() == reflect.Interface {
		return nil
	}
	if e, ok := reflect.New(deref(t)).Interface().(enumMapper); ok {
		return e.GetEnumMap()
	}
	return nil
}

// hasJSONFields reports whether a struct exposes any JSON properties.
func hasJSONFields(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.IsExported() && jsonName(f) != "-" {
			return true
		}
	}
	return false
}

func jsonName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return ""
	}
	return strings.Split(tag, ",")[0]
}

// classify maps a Go type to the kind of form item editing it, nested structs
// return true in the second value.
func classify(t reflect.Type, element string) (fieldKind, bool) {
	t = deref(t)
	switch {
	case enumOf(t) != nil:
		return kindEnum, false
	case element == "textarea":
		return kindTextArea, false
	case t.Kind() == reflect.Struct && hasJSONFields(t) && !reflect.PointerTo(t).Implements(jsonUnmarshaler):
		return 0, true
	case t.Kind() == reflect.Struct:
		return kindText, false // opaque type such as a template, edited as text
	case t.Kind() == reflect.Bool:
		return kindBool, false
	case t.Kind() == reflect.String:
		return kindText, false
	case reflect.PointerTo(t).Implements(jsonUnmarshaler):
		return kindText, false // e.g. durations accepting "30s"
	case t.Kind() >= reflect.Int && t.Kind() <= reflect.Float64:
		return kindNumber, false
	default:
		return kindJSON, false
	}
}

// describeStruct appends a field for every JSON property of t, recursing into
// nested structs.
func describeStruct(t reflect.Type, raw []byte, path, skip []string, fields *[]*formField) {
	t = deref(t)
	if t.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := jsonName(f)
		if name == "-" || (!f.IsExported() && !f.Anonymous) {
			continue
		}
		if f.Anonymous && name == "" {
			describeStruct(f.Type, raw, path, skip, fields)
			continue
		}
		if name == "" {
			name = f.Name
		}
		fieldPath := append(append([]string(nil), path...), name)
		if samePath(fieldPath, skip) {
			continue
		}

		kind, nested := classify(f.Type, f.Tag.Get("displayelement"))
		if nested {
			describeStruct(f.Type, raw, fieldPath, skip, fields)
			continue
		}
		goKind := deref(f.Type).Kind()
		field := &formField{
			path:    fieldPath,
			docKey:  f.Tag.Get("doc-key"),
			display: f.Tag.Get("displayname"),
			kind:    kind,
			goKind:  goKind,
			// kindJSON is always a slice/map too complex for a single line;
			// kindText additionally covers structs edited as opaque text
			// because they implement json.Unmarshaler (e.g. a legacy hook
			// format) but still marshal to a JSON object, not a string.
			compound: kind == kindJSON || (kind == kindText && goKind == reflect.Struct),
		}
		if kind == kindText && goKind == reflect.Struct {
			// a real Go struct type is known, so it can be edited as a
			// sub-form of its own instead of raw JSON.
			field.elemType = deref(f.Type)
		}
		if kind == kindEnum {
			field.options = sortedKeys(enumOf(f.Type))
		}
		field.load(raw)
		*fields = append(*fields, field)
	}
}

func samePath(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (f *formField) name() string { return strings.Join(f.path, ".") }

// load reads the current value from raw.
func (f *formField) load(raw []byte) {
	value, dataType, _, err := jsonparser.Get(raw, f.path...)
	f.exist = err == nil && dataType != jsonparser.NotExist && dataType != jsonparser.Null
	if f.exist {
		switch dataType {
		case jsonparser.String:
			if s, err := jsonparser.ParseString(value); err == nil {
				f.value = s
			} else {
				f.value = string(value)
			}
		case jsonparser.Object, jsonparser.Array:
			f.value = compact(value)
		default:
			f.value = string(value)
		}
	}
	if f.kind == kindBool && f.value == "" {
		f.value = "false"
	}
	f.orig = f.value
	f.set = f.exist
}

// changed reports whether the field differs from the loaded state.
func (f *formField) changed() bool { return f.set != f.exist || f.value != f.orig }

// unset marks the property for removal.
func (f *formField) unset() {
	f.set = false
	f.value = ""
	if f.kind == kindBool {
		f.value = "false"
	}
	f.sync()
}

// encode returns the JSON for the current value.
func (f *formField) encode() ([]byte, error) {
	switch f.kind {
	case kindBool:
		return []byte(f.value), nil
	case kindEnum:
		return json.Marshal(f.value)
	case kindNumber:
		var n json.Number
		if err := json.Unmarshal([]byte(f.value), &n); err != nil {
			return nil, errors.Errorf("%s: not a number", f.name())
		}
		return []byte(f.value), nil
	case kindJSON:
		if !json.Valid([]byte(f.value)) {
			return nil, errors.Errorf("%s: invalid JSON", f.name())
		}
		return compactBytes([]byte(f.value)), nil
	default:
		if f.goKind != reflect.String && json.Valid([]byte(f.value)) {
			return []byte(f.value), nil // e.g. a duration given as a number
		}
		return json.Marshal(f.value)
	}
}

// applyFields writes the changed fields into raw, removing unset properties.
func applyFields(raw []byte, fields []*formField) ([]byte, error) {
	for _, f := range fields {
		if !f.changed() {
			continue
		}
		if !f.set {
			if f.exist {
				raw = jsonparser.Delete(raw, f.path...)
			}
			continue
		}
		value, err := f.encode()
		if err != nil {
			return nil, err
		}
		if raw, err = jsonparser.Set(raw, value, f.path...); err != nil {
			return nil, errors.Wrapf(err, "%s", f.name())
		}
	}
	return raw, nil
}

// docs returns the documentation lines for a field.
func (f *formField) docs() []string {
	if f.docKey == "" {
		return nil
	}
	lines := generated.Params[f.docKey]
	if len(lines) < 1 {
		if entry, ok := generated.Config[f.docKey]; ok {
			lines = []string{entry.Description}
		}
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.ReplaceAll(strings.TrimSpace(line), "`", "")
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
