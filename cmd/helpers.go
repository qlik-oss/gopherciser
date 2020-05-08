package cmd

import (
	"fmt"
	"io/ioutil"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/spf13/cobra"
)

var (
	cfgFile string

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
}

// AddConfigParameter add config file parameter to command
func AddConfigParameter(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cfgFile, "config", "c", "", `Scenario config file.`)
}

func AddLoggingParameters(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&traffic, "traffic", "t", false, "Log traffic. Logging traffic is heavy and should only be done for debugging purposes.")
	cmd.Flags().BoolVarP(&trafficMetrics, "trafficmetrics", "m", false, "Log traffic metrics.")
	cmd.Flags().BoolVar(&debug, "debug", false, "Log debug info.")
	cmd.Flags().StringVar(&logFormat, "logformat", "", getLogFormatHelpString())
	cmd.Flags().StringVar(&summaryType, "summary", "", getSummaryTypeHelpString())
}

func unmarshalConfigFile() (*config.Config, error) {
	if cfgFile == "" {
		return nil, errors.Errorf("No config file defined")
	}

	cfgJSON, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading config from file<%s>", cfgFile)
	}

	var cfg config.Config
	if err = jsonit.Unmarshal(cfgJSON, &cfg); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal config from json")
	}

	return &cfg, nil
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
