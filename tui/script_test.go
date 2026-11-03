package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buger/jsonparser"
	"github.com/goccy/go-json"
)

const testScript = `{
  "settings": {
    "timeout": 300,
    "logs": {
      "filename": "./logs/log.tsv"
    }
  },
  "scenario": [
    {
      "label": "login",
      "action": "elasticlogin",
      "settings": {
        "mode": "jwt"
      }
    },
    {
      "action": "openapp",
      "disabled": true,
      "settings": {
        "app": "myapp"
      }
    }
  ],
  "connectionSettings": {
    "server": "localhost"
  }
}`

func writeTestScript(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "script.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}
	return path
}

func TestParseActions(t *testing.T) {
	actions, err := parseActions([]byte(testScript))
	if err != nil {
		t.Fatalf("parseActions: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	if actions[0].Type != "elasticlogin" || actions[0].Label != "login" {
		t.Errorf("unexpected first action: %+v", actions[0].ActionCore)
	}
	if !actions[1].Disabled {
		t.Error("expected second action to be disabled")
	}
	if got := actions[1].settingsOf(); got != `{"app":"myapp"}` {
		t.Errorf("unexpected settings: %s", got)
	}
	if !strings.HasPrefix(string(actions[0].raw), "{") {
		t.Errorf("action raw JSON is not an object: %s", actions[0].raw)
	}
}

func TestDocumentKeepsOtherSections(t *testing.T) {
	s, err := loadScript(writeTestScript(t, testScript))
	if err != nil {
		t.Fatalf("loadScript: %v", err)
	}

	// remove the login action and re-enable the openapp action
	s.actions = s.actions[1:]
	raw, err := setBool(s.actions[0].raw, false, "disabled")
	if err != nil {
		t.Fatalf("setBool: %v", err)
	}
	if err := s.actions[0].setRaw(raw); err != nil {
		t.Fatalf("setRaw: %v", err)
	}

	doc, err := s.document()
	if err != nil {
		t.Fatalf("document: %v", err)
	}

	actions, err := parseActions(doc)
	if err != nil {
		t.Fatalf("parseActions of document: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Type != "openapp" || actions[0].Disabled {
		t.Errorf("unexpected action: %+v", actions[0].ActionCore)
	}

	server, err := jsonparser.GetString(doc, "connectionSettings", "server")
	if err != nil || server != "localhost" {
		t.Errorf("connection settings lost: %q (%v)", server, err)
	}
	timeout, err := jsonparser.GetInt(doc, "settings", "timeout")
	if err != nil || timeout != 300 {
		t.Errorf("settings lost: %d (%v)", timeout, err)
	}
}

func TestDocumentEmptyScenario(t *testing.T) {
	s, err := loadScript(writeTestScript(t, testScript))
	if err != nil {
		t.Fatalf("loadScript: %v", err)
	}
	s.actions = nil

	doc, err := s.document()
	if err != nil {
		t.Fatalf("document: %v", err)
	}
	actions, err := parseActions(doc)
	if err != nil {
		t.Fatalf("parseActions: %v", err)
	}
	if len(actions) != 0 {
		t.Errorf("expected empty scenario, got %d actions", len(actions))
	}
}

func TestSaveAndList(t *testing.T) {
	path := writeTestScript(t, testScript)
	dir := filepath.Dir(path)

	s, err := loadScript(path)
	if err != nil {
		t.Fatalf("loadScript: %v", err)
	}
	s.actions = append(s.actions, newAction("changesheet"))
	if err := s.save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	scripts, err := listScripts(dir)
	if err != nil {
		t.Fatalf("listScripts: %v", err)
	}
	if len(scripts) != 1 {
		t.Fatalf("expected 1 script, got %d", len(scripts))
	}
	if scripts[0].actions != 3 {
		t.Errorf("expected 3 actions, got %d", scripts[0].actions)
	}
	if scripts[0].err != nil {
		t.Errorf("unexpected script error: %v", scripts[0].err)
	}
}

func TestSetRawRequiresActionType(t *testing.T) {
	act := newAction("openapp")
	if err := act.setRaw([]byte(`{"label":"no type","settings":{}}`)); err == nil {
		t.Error("expected error for action without type")
	}
	if err := act.setRaw([]byte(`{"action":"changesheet","settings":{"id":"abc"}}`)); err != nil {
		t.Fatalf("setRaw: %v", err)
	}
	if act.Type != "changesheet" {
		t.Errorf("unexpected type: %s", act.Type)
	}
}

func TestListScriptsIgnoresNonScripts(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "other.json"), []byte(`{"foo":"bar"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "script.json"), []byte(testScript), 0o644); err != nil {
		t.Fatal(err)
	}

	scripts, err := listScripts(dir)
	if err != nil {
		t.Fatalf("listScripts: %v", err)
	}
	if len(scripts) != 1 || scripts[0].name != "script.json" {
		t.Errorf("unexpected scripts: %+v", scripts)
	}
}

func TestSections(t *testing.T) {
	raw := []byte(`{
  "settings": {"timeout": 300},
  "connectionSettings": {"server": "localhost"},
  "scenario": [{"action": "thinktime"}]
}`)
	s := &script{path: "x.json", raw: raw}
	actions, err := parseActions(raw)
	if err != nil {
		t.Fatal(err)
	}
	s.actions = actions

	sections, err := s.sections()
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 2 || sections[0].name != "settings" || sections[1].name != "connectionSettings" {
		t.Fatalf("unexpected sections: %+v", sections)
	}

	if err := s.setSection("settings", []byte(`{"timeout": 10}`)); err != nil {
		t.Fatal(err)
	}
	if err := s.setSection("loginSettings", []byte(`{"type": "prefix"}`)); err != nil {
		t.Fatal(err)
	}
	s.deleteSection("connectionSettings")

	doc, err := s.document()
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(doc, &parsed); err != nil {
		t.Fatalf("document not valid JSON: %v\n%s", err, doc)
	}
	if _, ok := parsed["connectionSettings"]; ok {
		t.Fatal("connectionSettings not deleted")
	}
	if parsed["settings"].(map[string]any)["timeout"].(float64) != 10 {
		t.Fatalf("settings not updated: %s", doc)
	}
	if parsed["loginSettings"].(map[string]any)["type"] != "prefix" {
		t.Fatalf("loginSettings not added: %s", doc)
	}
	if len(parsed["scenario"].([]any)) != 1 {
		t.Fatalf("scenario lost: %s", doc)
	}

	if err := s.setSection("settings", []byte(`{broken`)); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFormFieldsAndApply(t *testing.T) {
	raw := []byte(`{"timeout": 300, "logs": {"filename": "log.tsv"}}`)
	spec := sectionSpec("settings")
	fields := spec.fields(raw)

	byName := map[string]*formField{}
	for _, f := range fields {
		byName[f.name()] = f
	}
	for _, name := range []string{"timeout", "logs.filename", "logs.format", "logs.debug", "maxerrors"} {
		if byName[name] == nil {
			t.Fatalf("missing field %s, got %v", name, names(fields))
		}
	}
	if byName["timeout"].value != "300" || byName["logs.filename"].value != "log.tsv" {
		t.Fatalf("values not loaded: %+v", byName["timeout"])
	}
	if byName["logs.format"].kind != kindEnum || !contains(byName["logs.format"].options, "tsvfile") {
		t.Fatalf("logs.format should be an enum: %+v", byName["logs.format"])
	}
	if byName["logs.debug"].kind != kindBool || byName["timeout"].kind != kindNumber {
		t.Fatal("unexpected kinds")
	}

	byName["timeout"].unset() // delete existing
	set := func(name, value string) { byName[name].value, byName[name].set = value, true }
	set("logs.debug", "true")      // set nested bool
	set("logs.format", "jsonfile") // set enum
	set("outputs.dir", "./out")    // create nested object
	updated, err := applyFields(raw, fields)
	if err != nil {
		t.Fatal(err)
	}
	if err := spec.validate(updated); err != nil {
		t.Fatalf("%v\n%s", err, updated)
	}
	var parsed map[string]any
	if err := json.Unmarshal(updated, &parsed); err != nil {
		t.Fatal(err)
	}
	if _, ok := parsed["timeout"]; ok {
		t.Fatalf("timeout not deleted: %s", updated)
	}
	logs := parsed["logs"].(map[string]any)
	if logs["debug"] != true || logs["format"] != "jsonfile" || logs["filename"] != "log.tsv" {
		t.Fatalf("logs not updated: %s", updated)
	}
	if parsed["outputs"].(map[string]any)["dir"] != "./out" {
		t.Fatalf("outputs.dir not created: %s", updated)
	}

	set("timeout", "abc")
	if _, err := applyFields(raw, fields); err == nil {
		t.Fatal("expected error for non numeric timeout")
	}
}

func TestFormSelectors(t *testing.T) {
	raw := []byte(`{"action":"openapp","label":"x","settings":{"appmode":"guid","app":"a"}}`)
	fields := actionSpec().fields(raw)
	got := names(fields)
	for _, name := range []string{"action", "label", "disabled", "settings.appmode", "settings.app"} {
		if !contains(got, name) {
			t.Fatalf("missing %s in %v", name, got)
		}
	}
	if fields[0].name() != "action" || fields[0].value != "openapp" || !contains(fields[0].options, "thinktime") {
		t.Fatalf("unexpected type field %+v", fields[0])
	}

	login := sectionSpec("loginSettings").fields([]byte(`{"type":"prefix","settings":{"prefix":"u"}}`))
	if !contains(names(login), "settings.prefix") || !contains(names(login), "settings.directory") {
		t.Fatalf("login fields: %v", names(login))
	}
	sched := sectionSpec("scheduler").fields([]byte(`{"type":"simple","settings":{"iterations":1}}`))
	if !contains(names(sched), "settings.iterations") || !contains(names(sched), "iterationtimebuffer.mode") {
		t.Fatalf("scheduler fields: %v", names(sched))
	}
	if sectionSpec("unknown") != nil {
		t.Fatal("unknown section should have no spec")
	}
}

func names(fields []*formField) []string {
	out := make([]string, len(fields))
	for i, f := range fields {
		out[i] = f.name()
	}
	return out
}

func TestFormUnsetVersusEmpty(t *testing.T) {
	raw := []byte(`{"server": "localhost", "appext": "sense/app"}`)
	fields := sectionSpec("connectionSettings").fields(raw)
	byName := map[string]*formField{}
	for _, f := range fields {
		byName[f.name()] = f
	}
	appext, mode := byName["appext"], byName["mode"]
	if appext == nil || !appext.set || appext.value != "sense/app" {
		t.Fatalf("appext not loaded: %+v", appext)
	}
	if mode == nil || mode.set {
		t.Fatalf("mode should be unset: %+v", mode)
	}

	// empty but set string is written as ""
	appext.value = ""
	updated, err := applyFields(raw, fields)
	if err != nil {
		t.Fatal(err)
	}
	if value, err := jsonparser.GetString(updated, "appext"); err != nil || value != "" {
		t.Fatalf("expected empty appext, got %q %v in %s", value, err, updated)
	}

	// unset removes the property
	appext.unset()
	updated, err = applyFields(raw, fields)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := jsonparser.Get(updated, "appext"); err == nil {
		t.Fatalf("appext should be removed: %s", updated)
	}

	// unset fields that never existed are not written, even bools
	if _, _, _, err := jsonparser.Get(updated, "allowuntrusted"); err == nil {
		t.Fatalf("unset bool written: %s", updated)
	}
	if byName["allowuntrusted"] != nil {
		byName["allowuntrusted"].value, byName["allowuntrusted"].set = "false", true
		updated, _ = applyFields(raw, fields)
		if v, err := jsonparser.GetBoolean(updated, "allowuntrusted"); err != nil || v {
			t.Fatalf("explicit false not written: %s", updated)
		}
	}
}

func TestSubActions(t *testing.T) {
	s, err := loadScript(writeTestScript(t, `{"scenario":[{"action":"iterated","settings":{"iterations":2,"actions":[{"action":"thinktime","settings":{}}]}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	app := NewApp(t.TempDir())
	v := app.scenario
	v.setScript(s)
	v.render()
	if !v.open() || len(v.cur) != 1 || v.cur[0].Type != "thinktime" {
		t.Fatalf("expected to open 1 sub action, got %v", v.cur)
	}

	v.insertAction(newAction("thinktime"), false)
	if len(v.cur) != 2 {
		t.Fatalf("expected 2 sub actions, got %d", len(v.cur))
	}
	doc, err := s.document()
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := parseActionsAt(doc, "scenario"); len(got) != 1 {
		t.Fatalf("scenario should keep 1 top level action, got %d", len(got))
	}
	if sub, err := s.actions[0].subActions(); err != nil || len(sub) != 2 {
		t.Fatalf("expected 2 sub actions in script, got %d, %v", len(sub), err)
	}

	v.undo()
	if len(v.cur) != 1 {
		t.Fatalf("undo should leave 1 sub action, got %d", len(v.cur))
	}
	if !v.up() || len(v.cur) != 1 || v.cur[0].Type != "iterated" {
		t.Fatalf("up should show the scenario, got %v", v.cur)
	}

	v.insertAction(newAction("iterated"), false)
	if !v.open() || len(v.cur) != 0 {
		t.Fatalf("expected to open an empty iterated action, got %v", v.cur)
	}
	v.insertAction(newAction("thinktime"), false)
	if sub, err := v.script.actions[1].subActions(); err != nil || len(sub) != 1 {
		t.Fatalf("expected 1 sub action in new iterated, got %d, %v", len(sub), err)
	}
}

func TestFormOpensSubActions(t *testing.T) {
	s, err := loadScript(writeTestScript(t, `{"scenario":[{"action":"iterated","settings":{"iterations":2,"actions":[{"action":"thinktime","settings":{}}]}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	app := newTestApp(t)
	v := app.scenario
	v.setScript(s)
	v.render()
	v.editAction(false)
	f, ok := app.top().(*form)
	if !ok {
		t.Fatalf("expected the action form, got %T", app.top())
	}
	for _, field := range f.fields {
		if samePath(field.path, []string{"settings", "actions"}) {
			f.drill(field)
		}
	}
	if len(v.path) != 1 || len(v.cur) != 1 || v.cur[0].Type != "thinktime" {
		t.Fatalf("expected the sub actions to be shown, path %v, actions %v", v.path, v.cur)
	}
	if app.top() == f {
		t.Fatal("form should be closed")
	}
}
