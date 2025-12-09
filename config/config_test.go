package config_test

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/runid"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/scheduler"
)

func TestConfig(t *testing.T) {
	tmpKeyFilePath := "../docs/examples/mock.pem"

	JSONConfigFile := `{
		"settings" : {
			"timeout" : 300
		},
		"scheduler" : {
			"type" : "simple",
			"settings" : {
				"executionTime" : -1,
				"iterations" : 2,
				"rampupDelay" : 7.0,
				"concurrentUsers" : 1
			}
		},
		"connectionSettings": {
			"server": "myserver",
			"mode": "jwt",
			"virtualproxy": "myvp",
			"security": true,
			"jwtsettings": {
				"keypath" : "` + tmpKeyFilePath + `",
				"claims" : "{\"user\":\"{{.UserName}}\",\"directory\":\"{{.Directory}}\"}"
			}
		},
		"loginSettings": {
			"type": "prefix",
			"settings": {
				"prefix": "gopher"
			}
		},
		"scenario" : [
			{
				"action": "OpenApp",
				"settings": {
					"appmode": "name",
					"app": "\"{{.UserName}}\""
				}
			},
			{
				"label" : "Select1",
				"action" : "select",
				"settings" : {
					"id" : "objid1",
					"type" : "values",
					"values" : [4,5],
					"accept" : true
				}
			},
			{
				"label" : "Select2",
				"action" : "select",
				"settings" : {
					"id" : "objid2",
					"type" : "values",
					"values" : [2],
					"accept" : false
				}
			}
		]
	}`

	var cfg config.Config
	if err := json.Unmarshal([]byte(JSONConfigFile), &cfg); err != nil {
		t.Fatal(err)
	}

	_, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Settings.Timeout != 300 {
		t.Errorf("Timout mismatch, expexted<300>, got<%d>", cfg.Settings.Timeout)
	}

	err = cfg.Validate()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Scheduler == nil {
		t.Fatal("nil scheduler")
	}

	sched, ok := cfg.Scheduler.(*scheduler.SimpleScheduler)
	if !ok {
		t.Errorf("Unexpected scheduler type<%T> expected<*scheduler.SimpleScheduler>", cfg.Scheduler)
	}
	schedSettings := sched.Settings

	if schedSettings.ConcurrentUsers != 1 {
		t.Errorf("Unexpected ConcurrentUsers<%d>, expected<1>", schedSettings.ConcurrentUsers)
	}

	if schedSettings.ExecutionTime != -1 {
		t.Errorf("Unexpected ExecutionTime<%d>, expected<-1>", schedSettings.ExecutionTime)
	}

	if schedSettings.RampupDelay != 7.0 {
		t.Errorf("Unexpected RampupDelay<%f>, expected<7.0>", schedSettings.RampupDelay)
	}

	if schedSettings.Iterations != 2 {
		t.Errorf("Unexpected Iterations<%d>, expected<2>", schedSettings.Iterations)
	}

	scenarioLength := len(cfg.Scenario)
	if scenarioLength != 3 {
		t.Fatalf("Unexpected scenario length, expected<2> got<%d>", scenarioLength)
	}

	if cfg.Scenario[1].Label != "Select1" {
		t.Errorf("incorrect label, expected<Select1> got<%s>", cfg.Scenario[0].Label)
	}

	if cfg.Scenario[2].Label != "Select2" {
		t.Errorf("incorrect label, expected<Select2> got<%s>", cfg.Scenario[1].Label)
	}

	if cfg.Scenario[1].Type != scenario.ActionSelect {
		t.Errorf("incorrect action(0), expected<Select> got<%s>", cfg.Scenario[0].Type)
	}

	if cfg.Scenario[2].Type != scenario.ActionSelect {
		t.Errorf("incorrect action(1), expected<Select> got<%s>", cfg.Scenario[1].Type)
	}

	settings, ok := cfg.Scenario[1].Settings.(*scenario.SelectionSettings)
	if !ok {
		t.Fatal("Failed to cast action(0) settings to SelectSettings")
	}

	if settings.ID != "objid1" {
		t.Errorf("Action(0): unexpected id<%s>, expected<objid1>", settings.ID)
	}

	if settings.Type != scenario.Values {
		t.Errorf("Action(0): unexpected type <%d>, expected<%d>", settings.Type, scenario.Values)
	}

	if !settings.Accept {
		t.Errorf("Action(0): unexpected accept <%v>, expected<true>", settings.Accept)
	}

	valuesLength := len(settings.Values)
	if valuesLength != 2 {
		t.Fatalf("Action(0): unexpected values length<%d>, expected<2>", valuesLength)
	}

	if settings.Values[0] != 4 {
		t.Errorf("Action(0): unexpected value1<%d>, expected<4>", settings.Values[0])
	}

	if settings.Values[1] != 5 {
		t.Errorf("Action(0): unexpected value2<%d>, expected<5>", settings.Values[1])
	}

	settings, ok = cfg.Scenario[2].Settings.(*scenario.SelectionSettings)
	if !ok {
		t.Fatal("Failed to cast action(1) settings to SelectSettings")
	}

	if settings.ID != "objid2" {
		t.Errorf("Action(1): unexpected id<%s>, expected<objid2>", settings.ID)
	}

	if settings.Type != scenario.Values {
		t.Errorf("Action(1): unexpected type <%d>, expected<%d>", settings.Type, scenario.Values)
	}

	if settings.Accept {
		t.Errorf("Action(1): unexpected accept <%v>, expected<false>", settings.Accept)
	}

	valuesLength = len(settings.Values)
	if valuesLength != 1 {
		t.Fatalf("Action(1): unexpected values length<%d>, expected<1>", valuesLength)
	}

	if settings.Values[0] != 2 {
		t.Errorf("Action(1): unexpected value1<%d>, expected<2>", settings.Values[0])
	}

	err = settingsTest(&cfg, tmpKeyFilePath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestScenario(t *testing.T) {
	JSONConfigFile := `{
		"scheduler" : {
			"type" : "simple"
		},
		"scenario" : [
			{
				"label" : "Open my app",
				"action" : "OpenApp",
				"settings" : {
					"appmode" : "guid",
					"app" : "1a8859ad-643c-49be-85cd-17f54ffa7aa4"
				}
			}
		]
	}`

	t.Logf("config file: %v", JSONConfigFile)

	var cfg config.Config
	if err := json.Unmarshal([]byte(JSONConfigFile), &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Scenario == nil {
		t.Fatal("Nil scenario")
	}

	if len(cfg.Scenario) != 1 {
		t.Errorf("Invalid scenario item length<%d>", len(cfg.Scenario))
	}

	if cfg.Scenario[0].Type != scenario.ActionOpenApp {
		t.Fatalf("Invalid action expected<OpenApp> got<%s>", cfg.Scenario[0].Type)
	}

	_, ok := cfg.Scenario[0].Settings.(*scenario.OpenAppSettings)
	if !ok {
		t.Fatalf("Failed to cast settings to OpenAppSettings is Type<%T>", cfg.Scenario[0].Settings)
	}
}

func TestScheduler(t *testing.T) {
	JSONConfigFile := `{
		"settings" : {
			"timeout" : 300
		},
		"scheduler" : {
			"type" : "simple",
			"settings" : {
				"executionTime" : -1,
				"iterations" : 2,
				"rampupDelay" : 10.0,
				"concurrentUsers" : 2
			},
			"iterationtimebuffer" : {
				"mode" : 1,
				"duration": "30s"
			}
		},
		"loginSettings": {
			"type": "prefix",
			"settings": {
				"prefix": "gopher"
			}
		},
		"scenario" : [
			{
				"action" : "OpenApp",
				"settings" : {}
			}
		]
	}`

	var cfg config.Config
	if err := json.Unmarshal([]byte(JSONConfigFile), &cfg); err != nil {
		t.Fatal(err)
	}

	sched, ok := cfg.Scheduler.(*scheduler.SimpleScheduler)
	if !ok {
		t.Errorf("Unexpected scheduler type<%T> expected<*scheduler.SimpleScheduler>", cfg.Scheduler)
	}

	// validate settings
	settings := sched.Settings

	if settings.ExecutionTime != -1 {
		t.Errorf("Unexpected execution time<%d> expected<-1>", settings.ExecutionTime)
	}
	if settings.Iterations != 2 {
		t.Errorf("Unexpected iterations<%d> expected<2>", settings.Iterations)
	}
	if settings.RampupDelay != 10.0 {
		t.Errorf("Unexpected RampupDelay<%f> expected<10.0>", settings.RampupDelay)
	}
	if settings.ConcurrentUsers != 2 {
		t.Errorf("Unexpected ConcurrentUsers<%d> expected<2>", settings.ConcurrentUsers)
	}

	if _, err := json.Marshal(&cfg); err != nil {
		t.Fatal(err)
	}

	marshJSON, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	expectedJSON := `{"executionTime":-1,"iterations":2,"rampupDelay":10,"concurrentUsers":2,"reuseUsers":false,"onlyinstanceseed":false}`

	if string(marshJSON) != expectedJSON {
		t.Errorf("Unexpected marshal result.\nHave:\n%s\nExpected:\n%s", marshJSON, expectedJSON)
	}

	// validate time buffer
	timeBuf := sched.TimeBuf
	if timeBuf.Mode != scheduler.TimeBufConstant {
		t.Errorf("iterationtimebuffer mode<%v>, expected<%v>", timeBuf.Mode, scheduler.TimeBufConstant)
	}

	if time.Duration(timeBuf.Duration) != 30*time.Second {
		t.Errorf("iterationtimebuffer duration<%v>, expected<%v>", timeBuf.Duration, 30*time.Second)
	}
}

func settingsTest(cfg *config.Config, keypath string) error {
	if cfg.ConnectionSettings.Server != "myserver" {
		return fmt.Errorf("expected open app server<myserver> got <%s>", cfg.ConnectionSettings.Server)
	}

	if cfg.ConnectionSettings.VirtualProxy != "myvp" {
		return fmt.Errorf("expected open app virtual proxy<myvp> got <%s>", cfg.ConnectionSettings.VirtualProxy)
	}

	if !cfg.ConnectionSettings.Security {
		return fmt.Errorf("expected open app security<true> got <false>")
	}

	if cfg.ConnectionSettings.JwtSettings.KeyPath != keypath {
		return fmt.Errorf("expected key path<%s> got<%s>", keypath, cfg.ConnectionSettings.JwtSettings.KeyPath)
	}

	engineUrl, err := cfg.ConnectionSettings.GetEngineUrl("", "")
	if err != nil {
		return fmt.Errorf("error getting open app url, err: %v", err)
	}

	if engineUrl.String() != "wss://myserver:443/myvp/app" {
		return fmt.Errorf("expected open app url<wss://myserver:443/myvp/app> got<%s>", engineUrl.String())
	}

	engineUrl, err = cfg.ConnectionSettings.GetEngineUrl("1a8859ad-643c-49be-85cd-17f54ffa7aa4", "")
	if err != nil {
		return fmt.Errorf("error getting open app url, err: %v", err)
	}

	if engineUrl.String() != "wss://myserver:443/myvp/app/1a8859ad-643c-49be-85cd-17f54ffa7aa4" {
		return fmt.Errorf("expected open app url<wss://myserver:443/myvp/app/1a8859ad-643c-49be-85cd-17f54ffa7aa4> got<%s>", engineUrl.String())
	}

	return nil
}

var mwCount int

type (
	writerEntry struct {
		level    logger.LogLevel
		details  string
		msg      string
		infoType string
		user     string
	}
	myMessageWriter struct {
		curLevel logger.LogLevel
		entries  map[int]logger.LogChanMsg
		lock     sync.Mutex
	}
)

func (w *myMessageWriter) WriteMessage(msg *logger.LogChanMsg) error {
	if w == nil {
		return errors.New("writer is nil")
	}

	if msg == nil {
		return errors.New("msg is nil")
	}

	if w.curLevel < msg.Level {
		return nil // we should not log
	}

	w.lock.Lock()
	defer w.lock.Unlock()

	if w.entries == nil {
		w.entries = make(map[int]logger.LogChanMsg, 10)
	}
	w.entries[mwCount] = *msg

	mwCount++

	return nil
}

func (w *myMessageWriter) Level(lvl logger.LogLevel) {
	w.curLevel = lvl
}

func TestCustomLogger(t *testing.T) {
	JSONConfigFile := `{
		"settings" : {
			"logs" : {
				"format" : "no"
			}
		},
		"scheduler" : {
			"type" : "simple",
			"settings" : {
				"executionTime" : 60,
				"iterations" : 2,
				"rampupDelay" : 0.01,
				"concurrentUsers" : 2
			}
		},
		"loginSettings": {
			"type": "userlist",
			"settings": {
				"userlist":  [
					{"username": "user1"},
					{"username": "user2"},
					{"username": "user3"},
					{"username": "user4"}
				]
			}
		},
		"scenario" : [
			{
				"action" : "thinktime",
				"settings" : {
					"type" : "static",
					"delay" : 0.002
				}
			}
		],
		"connectionSettings": {
			"server": "myserver",
			"mode": "ws",
			"security": true
		}
	}`

	// todo turn off status output

	// expected log rows on default level
	expectedInfoLevel := []writerEntry{
		{logger.ResultLevel, "delay:0.002", "", "", "user1"},
		{logger.ResultLevel, "delay:0.002", "", "", "user2"},
		{logger.ResultLevel, "delay:0.002", "", "", "user3"},
		{logger.ResultLevel, "delay:0.002", "", "", "user4"},
		{logger.WarningLevel, "", "Not existing", "", ""},
		{logger.InfoLevel, "", "0", "TotRequests", ""},
		{logger.InfoLevel, "", "0", "SequenceSummary", "user1"},
		{logger.InfoLevel, "", "0", "SequenceSummary", "user2"},
		{logger.InfoLevel, "", "0", "SequenceSummary", "user3"},
		{logger.InfoLevel, "", "0", "SequenceSummary", "user4"},
		{logger.InfoLevel, "", "", "GopherciserVersion", ""},
		{logger.InfoLevel, "", runid.Get(), "RunID", ""},
		{logger.InfoLevel, "", "0", "TotErrWarn", ""},
		{logger.InfoLevel, "", "$!IGNORE", "TestDuration", ""},
	}

	// Can't do parallel due to global variables (session, thread etc)
	t.Run("infoLevel", func(t *testing.T) {
		expected := append(expectedInfoLevel, writerEntry{logger.InfoLevel, "", "4", "TotActions", ""})
		cfg, err := setupConfig(JSONConfigFile)
		if err != nil {
			t.Fatal(err)
		}

		script, err := json.Marshal(cfg)
		if err != nil {
			t.Fatal("failed marshaling config:", err)
		}
		expected = append(expected, writerEntry{logger.InfoLevel, "", string(script), "Script", ""})

		runCustomLogTest(t, cfg, expected, logger.InfoLevel, nil)
	})

	t.Run("debugLevel", func(t *testing.T) {
		expected := append(expectedInfoLevel, []writerEntry{
			{logger.InfoLevel, "", "4", "TotActions", ""},
			{logger.DebugLevel, "", "Disconnect session", "", "user1"},
			{logger.DebugLevel, "", "Disconnect session", "", "user2"},
			{logger.DebugLevel, "", "Disconnect session", "", "user3"},
			{logger.DebugLevel, "", "Disconnect session", "", "user4"},
			{logger.DebugLevel, "", "thinktime START", "", "user1"},
			{logger.DebugLevel, "", "thinktime START", "", "user2"},
			{logger.DebugLevel, "", "thinktime START", "", "user3"},
			{logger.DebugLevel, "", "thinktime START", "", "user4"},
			{logger.DebugLevel, "", "thinktime END", "", "user1"},
			{logger.DebugLevel, "", "thinktime END", "", "user2"},
			{logger.DebugLevel, "", "thinktime END", "", "user3"},
			{logger.DebugLevel, "", "thinktime END", "", "user4"},
		}...)
		cfg, err := setupConfig(JSONConfigFile)
		if err != nil {
			t.Fatal(err)
		}
		cfg.Settings.LogSettings.Debug = true
		script, err := json.Marshal(cfg)
		if err != nil {
			t.Fatal("failed marshaling config:", err)
		}
		expected = append(expected, writerEntry{logger.InfoLevel, "", string(script), "Script", ""})
		runCustomLogTest(t, cfg, expected, logger.DebugLevel, nil)
	})

	t.Run("trafficLog", func(t *testing.T) {
		expected := append(expectedInfoLevel, []writerEntry{
			{logger.InfoLevel, "", "4", "TotActions", ""},
			{logger.TrafficLevel, "Sent", "traffic message", "", "user1"},
			{logger.MetricsLevel, "WS1GetListObjectData", "", "", "user1"},
		}...)
		addTraffic := func(lgr *logger.Logger) error {
			logChanMsg := logger.NewEmptyLogChanMsg()
			logChanMsg.Level = logger.TrafficLevel
			logChanMsg.Details = "Sent"
			logChanMsg.Message = "traffic message"
			logChanMsg.User = "user1"
			if err := lgr.Writer.WriteMessage(logChanMsg); err != nil {
				return err
			}
			logChanMsg = logger.NewEmptyLogChanMsg()
			logChanMsg.Level = logger.MetricsLevel
			logChanMsg.Details = "WS1GetListObjectData"
			logChanMsg.User = "user1"
			if err := lgr.Writer.WriteMessage(logChanMsg); err != nil {
				return err
			}
			return nil
		}
		cfg, err := setupConfig(JSONConfigFile)
		if err != nil {
			t.Fatal(err)
		}
		cfg.SetTrafficLogging()
		cfg.SetTrafficMetricsLogging()

		script, err := json.Marshal(cfg)
		if err != nil {
			t.Fatal("failed marshaling config:", err)
		}
		expected = append(expected, writerEntry{logger.InfoLevel, "", string(script), "Script", ""})
		runCustomLogTest(t, cfg, expected, logger.InfoLevel, addTraffic)
	})

	t.Run("nilLogger", func(t *testing.T) {
		cfg, err := setupConfig(JSONConfigFile)
		if err != nil {
			t.Fatal(err)
		}
		cfg.CustomLoggers = append(cfg.CustomLoggers, nil)
		if err := cfg.Execute(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
	})
}

func runCustomLogTest(t *testing.T, cfg *config.Config, expected []writerEntry, level logger.LogLevel, addLogs func(*logger.Logger) error) {
	t.Helper()

	mw := myMessageWriter{
		curLevel: logger.UnknownLevel,
	}
	customLogger := &logger.Logger{
		Writer: &mw,
	}
	cfg.CustomLoggers = append(cfg.CustomLoggers, customLogger)

	customLogger.Writer.Level(level)
	if mw.curLevel != level {
		t.Errorf("setting log level %s failed", level)
	}

	if err := cfg.Execute(context.Background(), nil); err != nil {
		t.Fatal(err)
	}

	if addLogs != nil {
		if err := addLogs(customLogger); err != nil {
			t.Fatal(err)
		}
	}

	validateExpected(t, expected, &mw)
}

func setupConfig(script string) (*config.Config, error) {
	var cfg config.Config
	if err := json.Unmarshal([]byte(script), &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validateExpected(t *testing.T, expected []writerEntry, mw *myMessageWriter) {
	t.Helper()

	expectedNotFound := make([]writerEntry, 0, 10)

expectedFor:
	for _, expectedEntry := range expected {
		for key, entry := range mw.entries {
			if isEntryEquatable(entry, expectedEntry) {
				delete(mw.entries, key)
				continue expectedFor
			}
		}
		expectedNotFound = append(expectedNotFound, expectedEntry)
	}

	for _, e := range expectedNotFound {
		if e.level == logger.WarningLevel && e.msg == "Not existing" && e.details == "" && e.infoType == "" { // make sure our assertion logic works
			continue
		}

		t.Errorf("Expected log row not found level<%s> msg<%s> details<%s> infotype<%s>", e.level, e.msg, e.details, e.infoType)
	}

	// re-use the jsonLogger to be able to see entire line
	buf := bytes.NewBuffer(nil)
	jLogger := logger.CreateJSONLogger(buf, nil)
	jLogger.Writer.Level(logger.DebugLevel) //make sure everything is written
	for _, entry := range mw.entries {
		if err := jLogger.Writer.WriteMessage(&entry); err != nil {
			t.Log("warning: json logger failed, using backup error printing")
			t.Error("Unexpected log row entry:", buf.String())
		} else {
			t.Errorf("Unexpected log row entry: level<%s> msg<%s> details<%s> infotype<%s>", entry.Level, entry.Message, entry.Details, entry.InfoType)
		}
		buf.Reset()
	}
}

func isEntryEquatable(msg logger.LogChanMsg, wrt writerEntry) bool {
	if msg.Level != wrt.level {
		return false
	}
	if wrt.msg != "$!IGNORE" && msg.Message != wrt.msg {
		return false
	}
	if msg.Details != wrt.details {
		return false
	}
	if msg.InfoType != wrt.infoType {
		return false
	}
	if msg.User != wrt.user {
		return false
	}
	return true
}
