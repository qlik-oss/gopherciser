package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/go-ps"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/spf13/cobra"
)

var (
	// parameters for config file
	cfgFile string

	// parameters for overrides
	scriptOverrides    []string
	scriptOverrideFile string

	// parameters for logging
	traffic        bool
	trafficMetrics bool
	debug          bool
	logFormat      string
	summaryType    string

	jsonit = jsoniter.ConfigCompatibleWithStandardLibrary
)

// AddAllSharedParameters add shared parameters to command
func AddAllSharedParameters(cmd *cobra.Command) {
	AddConfigParameter(cmd)
	AddOverrideParameters(cmd)
}

// AddConfigParameter add config file parameter to command
func AddConfigParameter(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cfgFile, "config", "c", "", `Scenario config file.`)
}

// AddOverrideParameters to command
func AddOverrideParameters(cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&scriptOverrides, "set", "s", nil, "Override a value in script with 'path/to/key=value'.")
	cmd.Flags().StringVar(&scriptOverrideFile, "setfromfile", "", "Override values from file where each row is path/to/key=value.")
}

// AddLoggingParameters add logging parameters to command
func AddLoggingParameters(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&traffic, "traffic", "t", false, "Log traffic. Logging traffic is heavy and should only be done for debugging purposes.")
	cmd.Flags().BoolVarP(&trafficMetrics, "trafficmetrics", "m", false, "Log traffic metrics.")
	cmd.Flags().BoolVar(&debug, "debug", false, "Log debug info.")
	cmd.Flags().StringVar(&logFormat, "logformat", "", getLogFormatHelpString())
	cmd.Flags().StringVar(&summaryType, "summary", "", getSummaryTypeHelpString())
}

func unmarshalConfigFile() (*config.Config, error) {
	var err error
	var cfgJSON []byte
	var hasPipe bool

	if !IsLaunchedByDebugger() {
		hasPipe, err = HasPipe()
		if err != nil {
			return nil, errors.Wrap(err, "error discovering if piped data exist")
		}
	}

	if cfgFile == "" {
		if !hasPipe {
			return nil, errors.New("no config file and nothing on stdin")
		}

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			b := scanner.Bytes()
			if cfgJSON == nil {
				cfgJSON = make([]byte, 0, len(b))
			}
			cfgJSON = append(cfgJSON, b...)
		}
	} else {
		cfgJSON, err = cfgJsonFromFile()
		if err != nil {
			return nil, errors.Wrapf(err, "Error reading config from file<%s>", cfgFile)
		}
	}

	var overrides []string
	cfgJSON, overrides, err = overrideScriptValues(cfgJSON, hasPipe)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var cfg config.Config
	if regression {
		cfg.Options.AcceptNoScheduler = true
	}
	if err = jsonit.Unmarshal(cfgJSON, &cfg); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal config from json")
	}

	if cfg.Settings.LogSettings.Format != config.LogFormatNoLogs {
		PrintOverrides(overrides)
	}

	return &cfg, nil
}

func cfgJsonFromFile() ([]byte, error) {
	if cfgFile == "" {
		return nil, errors.Errorf("No config file defined")
	}

	return ioutil.ReadFile(cfgFile)
}

func overrideScriptValues(cfgJSON []byte, hasPipe bool) ([]byte, []string, error) {
	var overrides []string
	if scriptOverrideFile != "" {
		overrideFile, err := helpers.NewRowFile(scriptOverrideFile)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Error reading overrides from file<%s>", scriptOverrideFile)
		}
		if scriptOverrides == nil {
			scriptOverrides = make([]string, 0, len(overrideFile.Rows()))
		}
		scriptOverrides = append(overrideFile.Rows(), scriptOverrides...) // let command line overrides override file overrides
	} else if cfgFile != "" && hasPipe { // if cfg file has been pointed to, but has stdin piped, assume it's overrides
		if scriptOverrides == nil {
			scriptOverrides = make([]string, 0, 10)
		}

		// golang can't detect char devices properly in cygwin, handle this by closing stdin after a second
		readingCtx, done := context.WithCancel(context.Background())
		if runtime.GOOS == "windows" {
			go func() {
				select {
				case <-time.After(time.Second):
					os.Stdin.Close()
				case <-readingCtx.Done():
				}
			}()
		}

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			scriptOverrides = append(scriptOverrides, scanner.Text())
		}
		done()
	}

	overrides = make([]string, 0, len(scriptOverrides))
	for _, kvp := range scriptOverrides {
		kvSplit := strings.SplitN(kvp, "=", 2)
		if len(kvSplit) != 2 {
			return cfgJSON, overrides, errors.Errorf("malformed override: %s, should be in the form 'path/to/key=value'", kvp)
		}

		path := helpers.DataPath(kvSplit[0])
		rawOrig, err := path.Lookup(cfgJSON)
		if err != nil {
			return cfgJSON, overrides, errors.Wrap(err, "invalid script override")
		}
		cfgJSON, err = path.Set(cfgJSON, []byte(kvSplit[1]))
		if err != nil {
			return cfgJSON, overrides, errors.WithStack(err)
		}
		rawModified, err := path.Lookup(cfgJSON)
		if err != nil {
			return cfgJSON, overrides, errors.WithStack(err)
		}
		overrides = append(overrides, fmt.Sprintf("%s: %s -> %s\n", path, rawOrig, rawModified))
	}

	return cfgJSON, overrides, nil
}

func getLogFormatHelpString() string {
	buf := helpers.NewBuffer()
	buf.WriteString("Set a log format to be used. One of:\n")
	config.LogFormatType(0).GetEnumMap().ForEachSorted(func(k int, v string) {
		addEnumToBuf(buf, k, v)
	})
	buf.WriteString("Defaults to in-script definition and falls back on ")
	defaultFormat, _ := config.LogFormatType(0).GetEnumMap().String(0)
	buf.WriteString(defaultFormat)
	buf.WriteString("\n")
	return buf.String()
}

func getSummaryTypeHelpString() string {
	buf := helpers.NewBuffer()
	buf.WriteString("Set a summary type to be used. One of:\n")
	config.SummaryType(0).GetEnumMap().ForEachSorted(func(k int, v string) {
		addEnumToBuf(buf, k, v)
	})
	return buf.String()
}

// ConfigOverrideLogSettings override log settings with parameters
func ConfigOverrideLogSettings(cfg *config.Config) error {
	if trafficMetrics {
		cfg.SetTrafficMetricsLogging()
	}

	if traffic {
		cfg.SetTrafficLogging()
	}

	if debug {
		cfg.SetDebugLogging()
	}

	if regression {
		cfg.SetRegressionLogging()
	}

	if logFormat != "" {
		var errLogformat error
		cfg.Settings.LogSettings.Format, errLogformat = resolveLogFormat(logFormat)
		if errLogformat != nil {
			return LogFormatError(fmt.Sprintf("error resolving log format<%s>: %v", logFormat, errLogformat))
		}
	}

	if summaryType != "" {
		if summary, errSummaryType := resolveSummaryType(); errSummaryType != nil {
			return SummaryTypeError(fmt.Sprintf("error resolving summary type<%s>: %v", summaryType, errSummaryType))
		} else {
			cfg.Settings.LogSettings.Summary = summary
		}
	}

	return nil
}

// PrintOverrides to script
func PrintOverrides(overrides []string) {
	if len(overrides) < 1 {
		return
	}
	os.Stderr.WriteString("=== Script overrides ===\n")
	for _, override := range overrides {
		os.Stderr.WriteString(override)
	}
	os.Stderr.WriteString("========================\n")
}

// HasPipe discovers if process has a chardevice attached
func HasPipe() (bool, error) {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false, errors.WithStack(err)
	}
	return fileInfo.Mode()&os.ModeCharDevice == 0, nil
}

// IsLaunchedByDebugger discovers if pararent process is deleve
func IsLaunchedByDebugger() bool {
	parent, err := ps.FindProcess(os.Getppid())
	if err != nil {
		return false
	}
	name := parent.Executable()
	switch name {
	case "dlv", "dlv.exe", "debugserver":
		return true
	}
	return false
}
