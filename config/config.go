package config

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/InVisionApp/tabular"
	"github.com/buger/jsonparser"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/runid"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/scheduler"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/synced"
	"github.com/qlik-oss/gopherciser/users"
	"github.com/qlik-oss/gopherciser/version"
	"github.com/shiena/ansicolor"
)

type (
	// LogFormatType one of: LogFormatTSVFile, LogFormatTSVConsole,
	//  LogFormatJSONFile, LogFormatJSONConsole, LogFormatColorConsole
	LogFormatType int

	SummaryType int

	// SummaryHeaderEntry is used to calculate summary column sizes
	SummaryHeaderEntry struct {
		FullName string
		ColSize  int
	}

	SummaryHeader map[string]*SummaryHeaderEntry

	// SummaryActionDataEntry data entry for action summary table
	SummaryActionDataEntry struct {
		Action      string
		Label       string
		AppGUID     string
		SuccessRate string
		AvgResp     string
		Requests    string
		Errs        string
		Warns       string
		Sent        string
		Received    string
	}

	// SummaryRequestDataEntry data entry for request summary table
	SummaryRequestDataEntry struct {
		Method   string
		Path     string
		AvgResp  string
		Requests string
		Sent     string
		Received string
	}

	// LogSettings settings for logging
	LogSettings struct {
		Traffic         bool            `json:"traffic,omitempty" displayname:"Traffic log" doc-key:"config.settings.logs.traffic"`
		Debug           bool            `json:"debug,omitempty" displayname:"Debug log" doc-key:"config.settings.logs.debug"`
		TrafficMetrics  bool            `json:"metrics,omitempty" displayname:"Traffic metrics log" doc-key:"config.settings.logs.metrics"`
		Regression      bool            `json:"regression,omitempty" displayname:"Regression log" doc-key:"config.settings.logs.regression"`
		FileName        synced.Template `json:"filename" displayname:"Log filename" displayelement:"savefile" doc-key:"config.settings.logs.filename"`
		Format          LogFormatType   `json:"format,omitempty" displayname:"Log format" doc-key:"config.settings.logs.format"`
		Summary         SummaryType     `json:"summary,omitempty" displayname:"Summary type" doc-key:"config.settings.logs.summary"`
		SummaryFileName string          `json:"summaryFilename,omitempty" displayname:"Name of summary file" doc-key:"config.settings.logs.summaryfile"`
	}

	// OutputsSettings settings for produced outputs (if any)
	OutputsSettings struct {
		Dir string `json:"dir" displayname:"Output directory" doc-key:"config.settings.outputs.dir"`
	}

	// Settings Config settings struct
	Settings struct {
		Timeout         int             `json:"timeout" displayname:"Request timeout" doc-key:"config.settings.timeout"` // Timeout in seconds
		LogSettings     LogSettings     `json:"logs" doc-key:"config.settings.logs"`
		OutputsSettings OutputsSettings `json:"outputs,omitempty" doc-key:"config.settings.outputs"`
		MaxErrorCount   uint64          `json:"maxerrors,omitempty" doc-key:"config.settings.maxerrors" displayname:"Max errors"`
	}

	hookData struct {
		Vars               map[string]interface{}
		Scheduler          map[string]interface{}
		ConnectionSettings *connection.ConnectionSettings
		Counters           *statistics.ExecutionCounters
	}

	Hooks struct {
		Pre  *Hook `json:"preexecute" doc-key:"config.hooks.preexecute" displayname:"Pre-execution Hook"`
		Post *Hook `json:"postexecute" doc-key:"config.hooks.postexecute" displayname:"Post-execution Hook"`

		data hookData
	}

	cfgCore struct {
		Scenario           []scenario.Action             `json:"scenario"`
		Settings           Settings                      `json:"settings"`
		LoginSettings      users.UserGenerator           `json:"loginSettings"`
		ConnectionSettings connection.ConnectionSettings `json:"connectionSettings"`
		Hooks              Hooks                         `json:"hooks"`
	}

	// Config setup and scenario to execute
	Config struct {
		*cfgCore
		schedType string
		Scheduler scheduler.IScheduler `json:"scheduler"`

		// CustomLoggers list of custom loggers.
		CustomLoggers []*logger.Logger `json:"-"`
		// Counters statistics for execution
		Counters statistics.ExecutionCounters `json:"-"`
		// ValidationWarnings list of script validation warnings
		ValidationWarnings []string `json:"-"`
		// Options alter the behavior of unmarshal and validate
		Options struct {
			// AcceptNoScheduler produces no error scheduler does not exist in json
			AcceptNoScheduler bool
		} `json:"-"`

		// Cancel execution, should be set to function triggering context cancel
		Cancel func(msg string) `json:"-"`
	}

	//SummaryEntry title, value and color combo for summary printout
	SummaryEntry struct {
		LongTitle  string `json:"longTitle"`  // used in extended and full summary
		ShortTitle string `json:"shortTitle"` // used in simple summary
		Value      string `json:"value"`
		Color      string `json:"-"`
	}
)

// ansi color codes
const (
	ansiReset      = "\x1b[0m"
	ansiStatus     = "\x1b[21;30;47m"
	ansiBoldRed    = "\x1b[1;31m"
	ansiBoldYellow = "\x1b[1;33m"
	ansiBoldBlue   = "\x1b[1;34m"
	ansiBoldWhite  = "\x1b[1;37m"
)

// LogFormat enum
const (
	// LogFormatTSVFile log to tsv file, and status to console
	LogFormatTSVFile LogFormatType = iota
	// LogFormatTSVConsole log tsv to console and no status output
	LogFormatTSVConsole
	// LogFormatJSONFile log to json file, and status to console
	LogFormatJSONFile
	// LogFormatJSONConsole log json to console and no status output
	LogFormatJSONConsole
	// LogFormatColorConsole log to console color formatted and no status output
	LogFormatColorConsole
	// LogFormatTSVFileJSONConsole log to  console in json format and to file in TSV format
	LogFormatTSVFileJSONConsole
	// LogFormatNoLogs turns off all default logging, custom loggers will still be used
	LogFormatNoLogs
	// LogFormatOnlyStatus turns off all default logging except status, custom loggers will still be used
	LogFormatOnlyStatus
)

// SummaryType enum
const (
	SummaryTypeDefault SummaryType = iota
	SummaryTypeNone
	SummaryTypeSimple
	SummaryTypeExtended
	SummaryTypeFull
	SummaryTypeFile
)

const DefaultSummaryFilename = "summary.json"

var (
	ansiWriter = ansicolor.NewAnsiColorWriter(os.Stdout)
)

func (value LogFormatType) GetEnumMap() *enummap.EnumMap {
	logFormatEnum, _ := enummap.NewEnumMap(map[string]int{
		"tsvfile":     int(LogFormatTSVFile),
		"tsvconsole":  int(LogFormatTSVConsole),
		"jsonfile":    int(LogFormatJSONFile),
		"jsonconsole": int(LogFormatJSONConsole),
		"console":     int(LogFormatColorConsole),
		"combined":    int(LogFormatTSVFileJSONConsole),
		"no":          int(LogFormatNoLogs),
		"onlystatus":  int(LogFormatOnlyStatus),
	})

	return logFormatEnum
}

func (value SummaryType) GetEnumMap() *enummap.EnumMap {
	summaryTypeEnum, _ := enummap.NewEnumMap(map[string]int{
		"default":  int(SummaryTypeDefault), // e.g. default value when un-marshaling from JSON etc
		"none":     int(SummaryTypeNone),
		"simple":   int(SummaryTypeSimple),
		"extended": int(SummaryTypeExtended),
		"full":     int(SummaryTypeFull),
		"file":     int(SummaryTypeFile),
	})

	return summaryTypeEnum
}

// UnmarshalJSON unmarshal LogFormatType
func (value *SummaryType) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.WithStack(err)
	}

	*value = SummaryType(i)

	return nil
}

// MarshalJSON marshal LogFormatType
func (value SummaryType) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	return []byte(fmt.Sprintf(`"%s"`, str)), errors.Wrapf(err, "failed to marshal SummaryType<%d>", value)
}

// UnmarshalJSON unmarshal LogFormatType
func (format *LogFormatType) UnmarshalJSON(arg []byte) error {
	i, err := format.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.WithStack(err)
	}

	*format = LogFormatType(i)

	return nil
}

// MarshalJSON marshal LogFormatType
func (format LogFormatType) MarshalJSON() ([]byte, error) {
	str, err := format.GetEnumMap().String(int(format))
	return []byte(fmt.Sprintf(`"%s"`, str)), errors.Wrapf(err, "failed to marshal LogFormatType<%d>", format)
}

// NewExampleConfig creates an example configuration populated with example data
func NewExampleConfig() (*Config, error) {

	// open hub action
	openHub := scenario.Action{}
	openHub.Type = scenario.ActionOpenHub
	openHub.Label = "open hub"
	openHub.Settings = scenario.OpenHubSettings{}

	// open app action
	appSelection, err := session.NewAppSelection(session.AppModeName, "myapp", nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	openApp := scenario.Action{
		Settings: scenario.OpenAppSettings{
			AppSelection: *appSelection,
		},
	}
	openApp.Type = scenario.ActionOpenApp
	openApp.Label = "open app"

	// think time action
	think := scenario.Action{
		Settings: scenario.ThinkTimeSettings{
			DistributionSettings: helpers.DistributionSettings{
				Type:      helpers.UniformDistribution,
				Mean:      15.0,
				Deviation: 5.0,
			},
		},
	}
	think.Type = scenario.ActionThinkTime
	think.Label = "think for 10-15s"

	// change sheet action
	changeSheet := scenario.Action{
		Settings: scenario.ChangeSheetSettings{ID: "QWERTY"},
	}
	changeSheet.Type = scenario.ActionChangeSheet
	changeSheet.Label = "change sheet to analysis sheet"

	// select action
	selectAction := scenario.Action{
		Settings: scenario.SelectionSettings{
			ID:   "uvxyz",
			Type: scenario.RandomFromEnabled,
			Min:  1,
			Max:  10,
		},
	}
	selectAction.Type = scenario.ActionSelect
	selectAction.Label = "select 1-10 values in object uvxyz"

	logFileName, err := synced.New("scenarioresult.tsv")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		cfgCore: &cfgCore{
			ConnectionSettings: connection.ConnectionSettings{
				Mode:           connection.WS,
				WsSettings:     nil,
				Server:         "localhost",
				VirtualProxy:   "header",
				Security:       true,
				Allowuntrusted: true,
				Headers: map[string]string{
					"Qlik-User-Header": "{{.UserName}}",
				},
			},
			LoginSettings: users.NewUserGeneratorPrefix("testuser"),
			Settings: Settings{
				Timeout: 300,
				LogSettings: LogSettings{
					FileName: *logFileName,
				},
			},
			Scenario: []scenario.Action{
				openHub,
				think,
				openApp,
				think,
				changeSheet,
				think,
				selectAction,
			},
		},
		Scheduler: &scheduler.SimpleScheduler{
			Scheduler: scheduler.Scheduler{
				SchedType: scheduler.SchedSimple,
				TimeBuf: scheduler.TimeBuffer{
					Mode:     scheduler.TimeBufOnError,
					Duration: helpers.TimeDuration(10 * time.Second),
				},
				InstanceNumber: 1,
			},
			Settings: scheduler.SimpleSchedSettings{
				ExecutionTime:   -1,
				Iterations:      10,
				RampupDelay:     7.0,
				ConcurrentUsers: 10,
				ReuseUsers:      false,
			},
		},
		CustomLoggers:      nil,
		Counters:           statistics.ExecutionCounters{},
		ValidationWarnings: nil,
	}

	return cfg, nil
}

// NewEmptyConfig creates an empty config
func NewEmptyConfig() (*Config, error) {

	logFileName, err := synced.New("scenarioresult.tsv")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		cfgCore: &cfgCore{
			ConnectionSettings: connection.ConnectionSettings{
				Mode:           connection.WS,
				WsSettings:     nil,
				Server:         "localhost",
				VirtualProxy:   "",
				Security:       true,
				Allowuntrusted: true,
				Headers:        map[string]string{},
			},
			LoginSettings: users.NewUserGeneratorPrefix("testuser"),
			Settings: Settings{
				Timeout: 300,
				LogSettings: LogSettings{
					FileName: *logFileName,
				},
			},
			Scenario: []scenario.Action{},
		},
		Scheduler: &scheduler.SimpleScheduler{
			Scheduler: scheduler.Scheduler{
				SchedType: scheduler.SchedSimple,
				TimeBuf: scheduler.TimeBuffer{
					Mode:     scheduler.TimeBufOnError,
					Duration: helpers.TimeDuration(10 * time.Second),
				},
				InstanceNumber: 1,
			},
			Settings: scheduler.SimpleSchedSettings{
				ExecutionTime:   -1,
				Iterations:      1,
				RampupDelay:     1,
				ConcurrentUsers: 1,
				ReuseUsers:      false,
			},
		},
		CustomLoggers:      nil,
		Counters:           statistics.ExecutionCounters{},
		ValidationWarnings: nil,
	}

	return cfg, nil
}

// UnmarshalJSON unmarshal Config. Only unmarshals scheduler if scheduler is nil.
func (cfg *Config) UnmarshalJSON(arg []byte) error {
	var core cfgCore
	if err := json.Unmarshal(arg, &core); err != nil {
		return errors.Wrap(err, "Failed unmarshaling config")
	}
	cfg.cfgCore = &core
	if cfg.Settings.LogSettings.Regression {
		cfg.Options.AcceptNoScheduler = true
	}

	rawsched, _, _, err := jsonparser.Get(arg, "scheduler")
	if err != nil {
		if cfg.Options.AcceptNoScheduler {
			return nil
		}
		return errors.Wrap(err, "no scheduler in config")
	}

	cfg.Scheduler, cfg.schedType, err = scheduler.UnmarshalScheduler(rawsched)
	if err != nil {
		return errors.Wrap(err, "failed unmarhaling scheduler")
	}

	return nil
}

// SetTrafficLogging override function to set traffic logging
func (cfg *Config) SetTrafficLogging() {
	cfg.Settings.LogSettings.Traffic = true
}

// SetTrafficMetricsLogging override function to set traffic logging
func (cfg *Config) SetTrafficMetricsLogging() {
	cfg.Settings.LogSettings.TrafficMetrics = true
}

// SetDebugLogging override function to set debug logging
func (cfg *Config) SetDebugLogging() {
	cfg.Settings.LogSettings.Debug = true
}

// SetRegressionLogging override function to set regression logging
func (cfg *Config) SetRegressionLogging() {
	cfg.Settings.LogSettings.Regression = true
}

func (cfg *Config) validateScheduler() error {
	if cfg.Scheduler == nil {
		if cfg.Options.AcceptNoScheduler {
			return nil
		} else {
			return errors.Errorf("No scheduler defined")
		}
	}

	w, err := cfg.Scheduler.Validate()
	cfg.ValidationWarnings = append(cfg.ValidationWarnings, w...)
	if err != nil {
		return errors.Wrap(err, "Scheduler settings validation failed")
	}

	if cfg.Scheduler.RequireScenario() {
		if cfg.Scenario == nil || len(cfg.Scenario) < 1 {
			return errors.Errorf("No scenario items defined")
		}
		for _, act := range cfg.Scenario {
			if schedValidate, ok := act.Settings.(scenario.ValidateActionForScheduler); ok {
				warnings, err := schedValidate.ValidateActionForScheduler(cfg.schedType)
				if err != nil {
					return errors.WithStack(err)
				}
				cfg.ValidationWarnings = append(cfg.ValidationWarnings, warnings...)
			}
		}

	}
	return nil
}

// Validate scenario
func (cfg *Config) Validate() error {
	cfg.ValidationWarnings = make([]string, 0)
	if err := cfg.validateScheduler(); err != nil {
		return err
	}

	if cfg.LoginSettings.Settings == nil {
		return errors.Errorf("No LoginSettings defined")
	}
	if err := cfg.LoginSettings.Settings.Validate(); err != nil {
		return errors.Wrap(err, "LoginSettings validation failed")
	}

	if cfg.ConnectionSettings.Server == "" {
		return errors.Errorf("Empty server name, server name is required")
	}
	if err := cfg.ConnectionSettings.Validate(); err != nil {
		return errors.Wrap(err, "ConnectionSettings validation failed")
	}

	// Validate all actions before executing
	for _, v := range cfg.Scenario {
		if w, err := v.Validate(); err != nil {
			return errors.WithStack(err)
		} else if len(w) > 0 {
			cfg.ValidationWarnings = append(cfg.ValidationWarnings, w...)
		}
	}

	// Validate hooks
	if w, err := cfg.Hooks.Validate(); err != nil {
		return errors.WithStack(err)
	} else if len(w) > 0 {
		cfg.ValidationWarnings = append(cfg.ValidationWarnings, w...)
	}

	return nil
}

func (cfg *Config) TestConnection(ctx context.Context) error {
	user := cfg.LoginSettings.GetNext(&cfg.Counters)
	cfg.Settings.LogSettings.Format = LogFormatNoLogs
	log, err := setupLogging(ctx, cfg.Settings.LogSettings, cfg.CustomLoggers, nil, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	if log == nil {
		return errors.New("setup logging returned nil logger")
	}
	sessionState := session.New(ctx, "", time.Duration(cfg.Settings.Timeout)*time.Second, user, 1, 1,
		cfg.ConnectionSettings.VirtualProxy, false, &cfg.Counters)
	logEntry := log.NewLogEntry()
	sessionState.SetLogEntry(logEntry)
	sessionState.LogEntry.Session = &logger.SessionEntry{}

	headers, err := cfg.ConnectionSettings.GetHeaders(sessionState, "")
	if err != nil {
		return errors.Wrap(err, "failed to generate authentication headers")
	}
	host, err := cfg.ConnectionSettings.GetHost()
	if err != nil {
		return errors.Wrap(err, "failed to extract hostname")
	}
	sessionState.HeaderJar.SetHeader(host, headers)

	client, err := session.DefaultClient(cfg.ConnectionSettings.Allowuntrusted, sessionState)
	if err != nil {
		return errors.Wrap(err, "failed to set up REST client")
	}
	sessionState.Rest.SetClient(client)

	actionState := &action.State{}
	sessionState.CurrentActionState = actionState

	errs := make([]error, 0)
	for _, connFunc := range scenario.GetConnTestFuncs() {
		if err = connFunc(&cfg.ConnectionSettings, sessionState, actionState); err == nil {
			break
		}
		errs = append(errs, err)
	}

	// At least one function succeeded
	if len(errs) < len(scenario.GetConnTestFuncs()) {
		return nil
	}

	errorString := fmt.Sprintf("Failed to connect using %d connection test functions.", len(errs))
	var i int
	for i, err = range errs {
		errorString = fmt.Sprintf("%s\n\nError #%d: %s", errorString, i+1, err)
	}

	return errors.New(errorString)
}

// Execute scenario (will be replaced by scheduler)
func (cfg *Config) Execute(ctx context.Context, templateData interface{}) error {
	timeout := time.Duration(cfg.Settings.Timeout) * time.Second
	// Setup logging
	log, err := setupLogging(ctx, cfg.Settings.LogSettings, cfg.CustomLoggers, templateData, &cfg.Counters)
	if err != nil {
		return errors.WithStack(err)
	}
	if log == nil {
		return errors.New("setup logging returned nil logger")
	}
	defer func() {
		if err := log.Close(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error closing log: ", err)
		}
	}()

	// Log version information at the start of the log
	entry := logger.NewLogEntry(log)
	entry.LogInfo("GopherciserVersion", version.Version)
	entry.LogInfo("RunID", runid.Get())

	// Log script validation warnings
	for _, warning := range cfg.ValidationWarnings {
		entry.LogInfo("ScriptValidationWarning", warning)
	}

	// Log script to be executed
	script, err := json.Marshal(cfg)
	if err != nil {
		entry.Logf(logger.WarningLevel, "failed to Marshal config for logging: %v", err)
	} else {
		entry.LogInfo("Script", string(script))
	}

	// Setup outputs folder
	outputsDir, err := setupOutputs(cfg.Settings.OutputsSettings)
	if err != nil {
		return errors.WithStack(err)
	}

	// start statistics collection if summarylevel high enough
	summaryType := cfg.Settings.LogSettings.getSummaryType()
	if err := cfg.SetupStatistics(summaryType); err != nil {
		return errors.WithStack(err)
	}

	// Log test summary after test is done
	defer summary(log, summaryType, time.Now(), &cfg.Counters, cfg.Settings.LogSettings.SummaryFileName)

	if cfg.Settings.MaxErrorCount > 0 {
		var once sync.Once
		cfg.Counters.SetMaxErrors(cfg.Settings.MaxErrorCount, func(msg string) {
			once.Do(func() {
				if cfg.Cancel == nil {
					entry.LogError(errors.Errorf("max errors surpassed, but no function to cancel execution is set!"))
					return
				}
				entry.Log(logger.ErrorLevel, msg)
				cfg.Cancel(msg)
			})
		})
	}

	cfg.PopulateHookData()

	// Execute pre execution hook
	if cfg.Hooks.Pre != nil {
		hookCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Settings.Timeout*int(time.Second)))
		defer cancel()
		if err := cfg.Hooks.Pre.Execute(hookCtx, entry, &cfg.Hooks.data, cfg.ConnectionSettings.Allowuntrusted); err != nil {
			return err
		}
	}

	if cfg.Hooks.Post != nil {
		defer func() {
			// Execute post execution hook
			hookCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Settings.Timeout*int(time.Second)))
			defer cancel()
			if err := cfg.Hooks.Post.Execute(hookCtx, entry, &cfg.Hooks.data, cfg.ConnectionSettings.Allowuntrusted); err != nil {
				entry.LogError(err)
			}
		}()
	}

	execErr := cfg.Scheduler.Execute(
		ctx, log, timeout, cfg.Scenario, outputsDir, cfg.LoginSettings, &cfg.ConnectionSettings, &cfg.Counters,
	)

	if execErr != nil {
		return errors.WithStack(execErr)
	}

	return nil
}

func (cfg *Config) PopulateHookData() {
	cfg.Hooks.data.Vars = make(map[string]interface{})
	cfg.Hooks.data.Scheduler = make(map[string]interface{})
	cfg.Scheduler.PopulateHookData(cfg.Hooks.data.Scheduler)
	cfg.Hooks.data.ConnectionSettings = &cfg.ConnectionSettings
	cfg.Hooks.data.Counters = &cfg.Counters
}

func (cfg *Config) SetupStatistics(summary SummaryType) error {
	switch summary {
	case SummaryTypeExtended, SummaryTypeFile:
		cfg.Counters.StatisticsCollector = statistics.NewCollector()
		return errors.WithStack(cfg.Counters.StatisticsCollector.SetLevel(statistics.StatsLevelOn))
	case SummaryTypeFull:
		cfg.Counters.StatisticsCollector = statistics.NewCollector()
		return errors.WithStack(cfg.Counters.StatisticsCollector.SetLevel(statistics.StatsLevelFull))
	}
	return nil
}

func (settings *LogSettings) shouldLogStatus() bool {
	if settings == nil {
		return true
	}
	switch settings.Format {
	case LogFormatNoLogs:
		return false
	case LogFormatTSVConsole:
		return false
	case LogFormatJSONConsole:
		return false
	case LogFormatColorConsole:
		return false
	case LogFormatTSVFileJSONConsole:
		return false
	default:
		return true
	}
}

func (settings *LogSettings) getSummaryType() SummaryType {
	switch settings.Summary {
	case SummaryTypeDefault:
		return SummaryTypeSimple
	default:
		if _, err := settings.Summary.GetEnumMap().String(int(settings.Summary)); err != nil {
			// illegal value used, default to none or simple
			if !settings.shouldLogStatus() {
				return SummaryTypeNone
			}
			return SummaryTypeSimple
		}
		return settings.Summary
	}
}

// Title returns long or short form of title depending on summary type
func (entry *SummaryEntry) Title(summary SummaryType) string {
	switch summary {
	case SummaryTypeSimple:
		return entry.ShortTitle
	default:
		return fmt.Sprintf("%-20s", entry.LongTitle)
	}
}

// ValueString returns value string as ": value" or "<value>" depending on summary type
func (entry *SummaryEntry) ValueString(summary SummaryType) string {
	switch summary {
	case SummaryTypeSimple:
		return "<" + entry.Value + ">"
	default:
		return ": " + entry.Value
	}
}

// EntryEnd returns new row or space depending on summary type
func (summary *SummaryType) EntryEnd() string {
	switch *summary {
	case SummaryTypeSimple:
		return " "
	default:
		return "\n"
	}
}

func summary(log *logger.Log, summary SummaryType, startTime time.Time, counters *statistics.ExecutionCounters, summaryFilename string) {
	testDuration := time.Since(startTime)

	entry := logger.NewLogEntry(log)

	errs := counters.Errors.Current()
	warnings := counters.Warnings.Current()
	actions := strconv.FormatUint(counters.ActionID.Current(), 10)
	requests := strconv.FormatUint(counters.Requests.Current(), 10)

	entry.LogErrorReport("TotErrWarn", errs, warnings)
	entry.LogInfo("TotActions", actions)
	entry.LogInfo("TotRequests", requests)
	entry.LogInfo("TestDuration", strconv.FormatInt(testDuration.Nanoseconds(), 10))

	buf := helpers.NewBuffer()
	defer func() {
		if buf.Error != nil {
			fmt.Printf("Summary: Errors<%d> Warnings<%d>\n", errs, warnings) // fallback to fmt
			return
		}

		buf.WriteTo(ansiWriter)
		if buf.Error != nil {
			fmt.Printf("Summary: Errors<%d> Warnings<%d>\n", errs, warnings) // fallback to fmt
			return
		}
	}()

	errorColor := ansiBoldBlue
	warningColor := ansiBoldBlue

	if errs > 0 {
		errorColor = ansiBoldRed
	}

	if warnings > 0 {
		warningColor = ansiBoldYellow
	}

	summaryData := []SummaryEntry{
		{"Total errors", "TotErrors", strconv.FormatUint(errs, 10), errorColor},
		{"Total warnings", "TotWarnings", strconv.FormatUint(warnings, 10), warningColor},
		{"Total actions", "TotActions", actions, ansiBoldBlue},
		{"Total requests", "TotRequests", requests, ansiBoldBlue},
		{"Duration", "Duration", testDuration.String(), ansiBoldBlue},
	}

	// Decide summary output
	switch summary {
	case SummaryTypeNone:
		//Don't log summary to stdout
		return
	case SummaryTypeFull, SummaryTypeExtended, SummaryTypeFile:
		summaryData = append(summaryData, []SummaryEntry{
			{"Total users", "TotUsers", strconv.FormatUint(counters.Users.Current(), 10), ansiBoldBlue},
			{"Total threads", "TotThreads", strconv.FormatUint(counters.Threads.Current(), 10), ansiBoldBlue},
			{"Total sessions", "TotSessions", strconv.FormatUint(counters.Sessions.Current(), 10), ansiBoldBlue},
			{"Total apps opened", "OpenedApps", strconv.FormatUint(counters.StatisticsCollector.OpenedApps(), 10), ansiBoldBlue},
			{"Total apps created", "CreatedApps", strconv.FormatUint(counters.StatisticsCollector.CreatedApps(), 10), ansiBoldBlue},
		}...)
	default:
		// default to simple summary
		summary = SummaryTypeSimple
	}

	if summary == SummaryTypeFile {
		jsn, err := json.Marshal(summaryData)
		if err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprint("failed to marshal summary file:", err))
			return
		}
		fileName := DefaultSummaryFilename
		if summaryFilename != "" {
			fileName = summaryFilename
		}
		if err := os.WriteFile(fileName, jsn, 0644); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprint("failed write summary file:", err))
		}
		return
	}

	buf.WriteString(ansiReset)
	for _, v := range summaryData {
		buf.WriteString(v.Color)
		buf.WriteString(v.Title(summary))
		buf.WriteString(v.ValueString(summary))
		buf.WriteString(ansiReset)
		buf.WriteString(summary.EntryEnd())
	}
	if summary == SummaryTypeSimple {
		buf.WriteString("\n")
	}

	// if not extended or full return now
	if summary < SummaryTypeExtended {
		return
	}

	// Separate sections
	buf.WriteString("\n")

	summaryHeaders := make(SummaryHeader)
	actionTblData := make([]SummaryActionDataEntry, 0, counters.StatisticsCollector.ActionsLen())

	// Create headers and default column sizes
	summaryHeaders["actn"] = &SummaryHeaderEntry{"Action", 6}
	summaryHeaders["lbl"] = &SummaryHeaderEntry{"Label", 5}
	summaryHeaders["app"] = &SummaryHeaderEntry{"AppGUID", 7}
	summaryHeaders["success"] = &SummaryHeaderEntry{"SuccessRate", 11}
	summaryHeaders["resp"] = &SummaryHeaderEntry{"AvgResp", 7}
	summaryHeaders["req"] = &SummaryHeaderEntry{"Requests", 8}
	summaryHeaders["errs"] = &SummaryHeaderEntry{"Errors", 6}
	summaryHeaders["warns"] = &SummaryHeaderEntry{"Warnings", 8}
	summaryHeaders["sent"] = &SummaryHeaderEntry{"Sent (Bytes)", 11}
	summaryHeaders["recvd"] = &SummaryHeaderEntry{"Received (Bytes)", 16}

	// todo max column size and truncate?
	// Calculate column lengths and fill data struct
	counters.StatisticsCollector.ForEachAction(func(stats *statistics.ActionStats) {
		// add data entry
		resp, successful := stats.RespAvg.Average()
		failed := stats.Failed.Current()

		// calculate success rate
		successRate := 0.0
		if successful > 0 {
			if failed < 1 {
				successRate = 100.0
			} else {
				successRate = successful / (successful + float64(failed)) * 100
			}
		}

		entry := SummaryActionDataEntry{
			Action:      stats.Name(),
			Label:       stats.Label(),
			AppGUID:     stats.AppGUID(),
			SuccessRate: fmt.Sprintf("%.2f%%", successRate),
			AvgResp:     time.Duration(resp).Round(time.Millisecond).String(),
			Requests:    stats.Requests.String(),
			Errs:        stats.ErrCount.String(),
			Warns:       stats.WarnCount.String(),
			Sent:        stats.Sent.String(),
			Received:    stats.Received.String(),
		}
		actionTblData = append(actionTblData, entry)

		summaryHeaders["actn"].UpdateColSize(len(stats.Name()))
		summaryHeaders["lbl"].UpdateColSize(len(stats.Label()))
		summaryHeaders["app"].UpdateColSize(len(stats.AppGUID()))
		summaryHeaders["success"].UpdateColSize(len(entry.SuccessRate))
		summaryHeaders["resp"].UpdateColSize(len(entry.AvgResp))
		summaryHeaders["req"].UpdateColSize(len(entry.Requests))
		summaryHeaders["errs"].UpdateColSize(len(entry.Errs))
		summaryHeaders["warns"].UpdateColSize(len(entry.Warns))
		summaryHeaders["sent"].UpdateColSize(len(entry.Sent))
		summaryHeaders["recvd"].UpdateColSize(len(entry.Received))
	})

	// Actions table
	tabbedOutput := tabular.New()

	for _, v := range []string{"actn", "lbl", "app"} {
		summaryHeaders.Col(v, &tabbedOutput)
	}

	for _, v := range []string{"success", "resp", "req", "errs", "warns", "sent", "recvd"} {
		summaryHeaders.ColRJ(v, &tabbedOutput)
	}

	// Action table headers
	table := tabbedOutput.Parse("*")
	writeTableHeaders(buf, &table)

	for _, v := range actionTblData {
		buf.WriteString(ansiBoldBlue)
		buf.WriteString(fmt.Sprintf(table.Format, v.Action, v.Label, v.AppGUID, v.SuccessRate, v.AvgResp, v.Requests, v.Errs, v.Warns, v.Sent, v.Received))
		buf.WriteString(ansiReset)
	}

	// Separate sections
	buf.WriteString("\n")

	summaryHeaders = make(SummaryHeader)
	requestsTblData := make([]SummaryRequestDataEntry, 0, counters.StatisticsCollector.RESTRequestLen())

	// Create headers and default column sizes
	summaryHeaders["path"] = &SummaryHeaderEntry{"Endpoint", 8}
	summaryHeaders["method"] = &SummaryHeaderEntry{"Method", 6}
	summaryHeaders["resp"] = &SummaryHeaderEntry{"AvgResp", 7}
	summaryHeaders["req"] = &SummaryHeaderEntry{"Requests", 8}
	summaryHeaders["sent"] = &SummaryHeaderEntry{"Sent (Bytes)", 11}
	summaryHeaders["recvd"] = &SummaryHeaderEntry{"Received (Bytes)", 16}

	counters.StatisticsCollector.ForEachRequest(func(stats *statistics.RequestStats) {
		resp, requests := stats.RespAvg.Average()
		entry := SummaryRequestDataEntry{
			Method:   stats.Method(),
			Path:     stats.Path(),
			AvgResp:  time.Duration(resp).Round(time.Millisecond).String(),
			Requests: fmt.Sprintf("%d", uint64(math.Round(requests))),
			Sent:     stats.Sent.String(),
			Received: stats.Received.String(),
		}
		requestsTblData = append(requestsTblData, entry)
		summaryHeaders["path"].UpdateColSize(len(stats.Path()))
		summaryHeaders["method"].UpdateColSize(len(stats.Method()))
		summaryHeaders["resp"].UpdateColSize(len(entry.AvgResp))
		summaryHeaders["req"].UpdateColSize(len(entry.Requests))
		summaryHeaders["sent"].UpdateColSize(len(entry.Sent))
		summaryHeaders["recvd"].UpdateColSize(len(entry.Received))
	})

	// if not full summary return here
	if summary < SummaryTypeFull {
		return
	}

	// REST Requests table
	tabbedOutput = tabular.New()

	for _, v := range []string{"path", "method"} {
		summaryHeaders.Col(v, &tabbedOutput)
	}

	for _, v := range []string{"resp", "req", "sent", "recvd"} {
		summaryHeaders.ColRJ(v, &tabbedOutput)
	}

	// Action table headers
	table = tabbedOutput.Parse("*")
	writeTableHeaders(buf, &table)

	for _, v := range requestsTblData {
		buf.WriteString(ansiBoldBlue)
		buf.WriteString(fmt.Sprintf(table.Format, v.Path, v.Method, v.AvgResp, v.Requests, v.Sent, v.Received))
		buf.WriteString(ansiReset)
	}

}

func writeTableHeaders(buf *helpers.Buffer, table *tabular.Output) {
	// Action table headers
	buf.WriteString(ansiBoldBlue)
	buf.WriteString(table.Header)
	buf.WriteString(ansiReset)
	buf.WriteString("\n")
	buf.WriteString(ansiBoldBlue)
	buf.WriteString(table.SubHeader)
	buf.WriteString(ansiReset)
	buf.WriteString("\n")
}

// UpdateColSize for summary header entry
func (entry *SummaryHeaderEntry) UpdateColSize(new int) {
	if new <= entry.ColSize {
		return
	}
	entry.ColSize = new
}

// Col sets column in table from header entry
func (header SummaryHeader) Col(key string, tbl *tabular.Table) {
	tbl.Col(key, header[key].FullName, header[key].ColSize)
}

// ColRJ sets column (Right Justified) in table from header entry
func (header SummaryHeader) ColRJ(key string, tbl *tabular.Table) {
	tbl.ColRJ(key, header[key].FullName, header[key].ColSize)
}

// Validate hooks settings
func (hooks Hooks) Validate() ([]string, error) {
	warnings := make([]string, 0)
	appendWarnings := func(w []string) {
		if len(w) > 0 {
			warnings = append(warnings, w...)
		}
	}
	if hooks.Pre != nil {
		w, err := hooks.Pre.Validate()
		appendWarnings(w)
		if err != nil {
			return warnings, errors.WithStack(err)
		}
	}
	if hooks.Post != nil {
		w, err := hooks.Post.Validate()
		appendWarnings(w)
		if err != nil {
			return warnings, errors.WithStack(err)
		}
	}
	return warnings, nil
}

func setupOutputs(settings OutputsSettings) (string, error) {
	if settings.Dir == "" {
		return "", nil
	}

	// Get absolute dir path
	absPath, err := filepath.Abs(settings.Dir)
	if err != nil {
		return "", err
	}
	// Create directory if not exists
	err = os.MkdirAll(absPath, os.ModePerm)
	return absPath, err
}

func addTSVFileLogger(log *logger.Log, filename string) error {
	filewriter, closeFile, fileWriterErr := createFileWriter(filename)
	if fileWriterErr != nil {
		return errors.WithStack(fileWriterErr)
	}

	tsvLogger, tsvErr := logger.CreateTSVLogger(logger.AllFields, filewriter, closeFile)
	if tsvErr != nil {
		return errors.WithStack(tsvErr)
	}
	log.AddLoggers(tsvLogger)
	return nil
}

func setupLogging(ctx context.Context, settings LogSettings, customLoggers []*logger.Logger, templateData interface{}, counters *statistics.ExecutionCounters) (*logger.Log, error) {
	log := logger.NewLog(logger.LogSettings{
		Traffic:    settings.Traffic,
		Metrics:    settings.TrafficMetrics,
		Debug:      settings.Debug,
		Regression: settings.Regression,
	})

	filename, err := settings.FileName.ExecuteString(templateData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to expand session variables in filename")
	}

	if log.Settings.Regression {
		err := log.SetRegressionLoggerFile(filename)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set up regression logging")
		}
	}

	switch settings.Format {
	case LogFormatTSVFile: // TSV log file
		if err := addTSVFileLogger(log, filename); err != nil {
			return log, errors.WithStack(err)
		}
	case LogFormatTSVConsole: // TSV console
		tsvLogger, tsvErr := logger.CreateTSVLogger(logger.AllFields, os.Stdout, nil)
		if tsvErr != nil {
			return nil, errors.WithStack(tsvErr)
		}
		log.AddLoggers(tsvLogger)
	case LogFormatJSONFile: // JSON log file
		filewriter, closeFile, fileWriterErr := createFileWriter(filename)
		if fileWriterErr != nil {
			return log, errors.WithStack(fileWriterErr)
		}

		jsonLogger := logger.CreateJSONLogger(filewriter, closeFile)
		log.AddLoggers(jsonLogger)
	case LogFormatJSONConsole: // JSON Console
		stdoutJSON := logger.CreateStdoutJSONLogger()
		log.AddLoggers(stdoutJSON)
	case LogFormatColorConsole: // Color console
		stdout := logger.CreateStdoutLogger()
		log.AddLoggers(stdout)
	case LogFormatTSVFileJSONConsole: // TSV file, JSON console
		if err := addTSVFileLogger(log, filename); err != nil {
			return log, errors.WithStack(err)
		}
		stdoutJSON := logger.CreateStdoutJSONLogger()
		log.AddLoggers(stdoutJSON)

	case LogFormatNoLogs: // No default logging
	case LogFormatOnlyStatus:
		// add dummy logger to make sure logs are not congesting log channel
		log.AddLoggers(logger.CreateDummyLogger())
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Unsupported log format requested: %v\n", settings.Format)
	}

	if len(customLoggers) > 0 {
		log.AddLoggers(customLoggers...)
	}

	if settings.shouldLogStatus() {
		// status output
		go statusPrinter(ctx, 10*time.Second, log.Closed, counters)
	}

	if settings.TrafficMetrics {
		log.SetMetrics()
	}

	// Traffic logging
	if settings.Traffic {
		log.SetTraffic()
	}

	// Debug logging
	if settings.Debug {
		log.SetDebug()
	}

	// Start logging
	log.StartLogger(ctx)

	return log, nil
}

// statusPrinter should be started as goroutine
func statusPrinter(ctx context.Context, statusDelay time.Duration, closeChan chan interface{}, counters *statistics.ExecutionCounters) {
	errorColor := ansiStatus
	warningColor := ansiStatus

	buf := helpers.NewBuffer()

	for {
		myErrors := counters.Errors.Current()
		warnings := counters.Warnings.Current()

		if myErrors > 0 {
			errorColor = ansiBoldRed
		}

		if warnings > 0 {
			warningColor = ansiBoldYellow
		}

		timestamp := time.Now().Format(time.RFC3339)

		// Example:
		// "Err<0> Warn<0> ActvSess<1> TotSess<2> Actns<14> Reqs<234>"
		strs := []string{
			ansiStatus,
			timestamp, " ",
			errorColor, "Err<", strconv.FormatUint(myErrors, 10), ">", ansiReset,
			ansiStatus, " ",
			warningColor, "Warn<", strconv.FormatUint(warnings, 10), ">", ansiReset,
			ansiStatus, " ActvSess<", strconv.FormatUint(counters.ActiveUsers.Current(), 10), ">",
			" TotSess<", strconv.FormatUint(counters.Sessions.Current(), 10), ">",
			" Actns<", strconv.FormatUint(counters.ActionID.Current(), 10), ">",
			" Reqs<", strconv.FormatUint(counters.Requests.Current(), 10), ">",
			ansiReset, "\n",
		}

		buf.Reset()

		// Status ticker is not that important, in case of error just ignore this tick
		for _, s := range strs {
			buf.WriteString(s)
			if buf.Error != nil {
				goto StatusWait
			}
		}

		buf.WriteTo(ansiWriter)
		if buf.Error != nil {
			goto StatusWait
		}

	StatusWait:
		select {
		case <-ctx.Done():
			return
		case <-closeChan:
			return
		case <-time.After(statusDelay):
		}
	}
}

func createFileWriter(filename string) (io.Writer, func() error, error) {
	writer, err := logger.NewWriter(filename)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return writer, writer.Close, nil
}
