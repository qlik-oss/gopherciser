package tui

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/qlik-oss/gopherciser/scenario"
)

type (
	// scriptInfo is a script as listed in the script browser.
	scriptInfo struct {
		name     string
		path     string
		actions  int
		size     int64
		modified time.Time
		err      error
	}

	// action is one entry of the scenario array, raw JSON kept as source of truth.
	action struct {
		scenario.ActionCore
		raw []byte
	}

	// script is a loaded scenario script, edited in memory until saved.
	script struct {
		path    string
		raw     []byte
		actions []action
	}
)

const indent = "  "

// listScripts finds the json files directly in dir containing a scenario array.
func listScripts(dir string) ([]scriptInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list scripts in <%s>", dir)
	}

	var scripts []scriptInfo
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(name), ".json") {
			continue
		}

		path := filepath.Join(dir, name)
		content, errRead := os.ReadFile(path)
		if errRead != nil {
			continue
		}
		if _, _, _, errScenario := jsonparser.Get(content, "scenario"); errScenario != nil {
			continue // not a gopherciser script
		}

		info := scriptInfo{name: name, path: path}
		if stat, errStat := entry.Info(); errStat == nil {
			info.size = stat.Size()
			info.modified = stat.ModTime()
		}
		actions, errActions := parseActions(content)
		if errActions != nil {
			info.err = errActions
		}
		info.actions = len(actions)

		scripts = append(scripts, info)
	}

	sort.Slice(scripts, func(i, j int) bool { return scripts[i].name < scripts[j].name })
	return scripts, nil
}

// loadScript reads a script from disk.
func loadScript(path string) (*script, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read <%s>", path)
	}
	actions, err := parseActions(raw)
	if err != nil {
		return nil, err
	}
	return &script{path: path, raw: raw, actions: actions}, nil
}

// parseActions extracts the scenario array as raw action entries.
func parseActions(raw []byte) ([]action, error) {
	return parseActionsAt(raw, "scenario")
}

// parseActionsAt extracts the array found under keys as raw action entries.
func parseActionsAt(raw []byte, keys ...string) ([]action, error) {
	var actions []action
	var parseErr error

	_, err := jsonparser.ArrayEach(raw, func(value []byte, dataType jsonparser.ValueType, _ int, _ error) {
		if parseErr != nil {
			return
		}
		if dataType != jsonparser.Object {
			parseErr = errors.Errorf("scenario entry %d is not an object", len(actions)+1)
			return
		}
		entry := action{raw: bytes.TrimSpace(bytes.Clone(value))}
		if err := json.Unmarshal(entry.raw, &entry.ActionCore); err != nil {
			parseErr = errors.Wrapf(err, "scenario entry %d", len(actions)+1)
			return
		}
		actions = append(actions, entry)
	}, keys...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s array", strings.Join(keys, "."))
	}

	return actions, parseErr
}

// document returns the script JSON with the in-memory scenario array applied.
func (s *script) document() ([]byte, error) {
	array, err := actionsArray(s.actions)
	if err != nil {
		return nil, err
	}
	raw, err := jsonparser.Set(s.raw, array, "scenario")
	if err != nil {
		return nil, errors.Wrap(err, "failed to set scenario array")
	}
	return raw, nil
}

// actionsArray returns actions as an indented JSON array.
func actionsArray(actions []action) ([]byte, error) {
	buf := bytes.NewBufferString("[")
	for i, act := range actions {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString("\n" + indent + indent)
		pretty, err := indentJSON(act.raw, indent+indent)
		if err != nil {
			return nil, errors.Wrapf(err, "action %d (%s)", i+1, act.Type)
		}
		buf.Write(pretty)
	}
	if len(actions) > 0 {
		buf.WriteString("\n" + indent)
	}
	buf.WriteByte(']')
	return buf.Bytes(), nil
}

// hasSubActions reports whether the action holds a list of actions in settings.actions.
func (act action) hasSubActions() bool { return act.Type == "iterated" }

// subActions returns the actions in settings.actions of an action.
func (act action) subActions() ([]action, error) {
	if _, _, _, err := jsonparser.Get(act.raw, "settings", "actions"); err != nil {
		return nil, nil
	}
	return parseActionsAt(act.raw, "settings", "actions")
}

// withSubActions returns the action with settings.actions replaced.
func (act action) withSubActions(sub []action) (action, error) {
	array, err := actionsArray(sub)
	if err != nil {
		return act, err
	}
	raw, err := jsonparser.Set(act.raw, array, "settings", "actions")
	if err != nil {
		return act, errors.Wrap(err, "failed to set sub actions")
	}
	err = act.setRaw(raw)
	return act, err
}

// replaceActionsAt returns root with the actions below the actions at path replaced by list.
func replaceActionsAt(root []action, path []int, list []action) ([]action, error) {
	if len(path) == 0 {
		return list, nil
	}
	parent := root[path[0]]
	sub, err := parent.subActions()
	if err != nil {
		return nil, err
	}
	if sub, err = replaceActionsAt(sub, path[1:], list); err != nil {
		return nil, err
	}
	if parent, err = parent.withSubActions(sub); err != nil {
		return nil, err
	}
	out := append([]action(nil), root...)
	out[path[0]] = parent
	return out, nil
}

// save writes the script back to disk.
func (s *script) save() error {
	raw, err := s.document()
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.path, raw, 0o644); err != nil {
		return errors.Wrapf(err, "failed to write <%s>", s.path)
	}
	s.raw = raw
	return nil
}

// config unmarshals and validates the current state of the script. Validation
// warnings are returned even when the config is valid.
func (s *script) config() (*config.Config, []string, error) {
	raw, err := s.document()
	if err != nil {
		return nil, nil, err
	}

	var cfg config.Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal script")
	}
	if err := cfg.Validate(); err != nil {
		return nil, cfg.ValidationWarnings, errors.Wrap(err, "script validation failed")
	}
	return &cfg, cfg.ValidationWarnings, nil
}

// settingsOf returns the settings object of an action as compact JSON.
func (act action) settingsOf() string {
	settings, dataType, _, err := jsonparser.Get(act.raw, "settings")
	if err != nil || dataType == jsonparser.NotExist {
		return ""
	}
	compact := bytes.NewBuffer(nil)
	if err := json.Compact(compact, settings); err != nil {
		return string(settings)
	}
	return compact.String()
}

// setRaw replaces the raw JSON of an action, keeping the parsed core in sync.
func (act *action) setRaw(raw []byte) error {
	var core scenario.ActionCore
	if err := json.Unmarshal(raw, &core); err != nil {
		return errors.Wrap(err, "invalid action JSON")
	}
	if core.Type == "" {
		return errors.New(`action is missing the "action" property`)
	}
	act.raw = bytes.TrimSpace(bytes.Clone(raw))
	act.ActionCore = core
	return nil
}

// indentJSON pretty prints raw JSON, prefixing all but the first line.
func indentJSON(raw []byte, prefix string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := json.Indent(buf, raw, prefix, indent); err != nil {
		return nil, errors.Wrap(err, "invalid JSON")
	}
	return buf.Bytes(), nil
}

// newAction creates a scenario entry with empty settings for an action type.
func newAction(actionType string) action {
	act := action{ActionCore: scenario.ActionCore{Type: actionType}}
	act.raw = []byte(`{"action":"` + actionType + `","label":"","settings":{}}`)
	return act
}

// setBool sets a boolean property in a JSON object.
func setBool(raw []byte, value bool, keys ...string) ([]byte, error) {
	set := []byte("false")
	if value {
		set = []byte("true")
	}
	updated, err := jsonparser.Set(raw, set, keys...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to set %s", strings.Join(keys, "/"))
	}
	return updated, nil
}

// section is a top level property of the script other than the scenario array.
type section struct {
	name string
	raw  []byte
}

// knownSections are the top level config properties offered when adding a section.
var knownSections = []string{"settings", "connectionSettings", "loginSettings", "scheduler", "hooks"}

// sections lists the top level properties of the script in document order.
func (s *script) sections() ([]section, error) {
	var out []section
	err := jsonparser.ObjectEach(s.raw, func(key, value []byte, dataType jsonparser.ValueType, _ int) error {
		if string(key) == "scenario" {
			return nil
		}
		raw := bytes.Clone(value)
		if dataType == jsonparser.String {
			raw = append(append([]byte(`"`), raw...), '"')
		}
		out = append(out, section{name: string(key), raw: raw})
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to read script properties")
	}
	return out, nil
}

// setSection replaces or adds a top level property.
func (s *script) setSection(name string, raw []byte) error {
	if !json.Valid(raw) {
		return errors.New("invalid JSON")
	}
	_, _, _, errGet := jsonparser.Get(s.raw, name)
	exists := errGet == nil

	pretty, err := indentJSON(raw, indent)
	if err != nil {
		return err
	}
	updated, err := jsonparser.Set(s.raw, pretty, name)
	if err != nil {
		return errors.Wrapf(err, "failed to set %s", name)
	}
	if !exists {
		// new properties are appended without whitespace, reformat the document
		if formatted, err := formatJSON(updated); err == nil {
			updated = formatted
		}
	}
	s.raw = updated
	return nil
}

// deleteSection removes a top level property.
func (s *script) deleteSection(name string) {
	s.raw = jsonparser.Delete(s.raw, name)
}

// clone returns an independent copy of the script state.
func (s *script) clone() *script {
	return &script{
		path:    s.path,
		raw:     bytes.Clone(s.raw),
		actions: append([]action(nil), s.actions...),
	}
}
