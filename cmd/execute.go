package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/buildmetrics"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/profile"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/scheduler"
	"github.com/qlik-oss/gopherciser/senseobjdef"
	"github.com/spf13/cobra"
)

type (
	// JSONParseError error during unmarshal of JSON file
	JSONParseError string
	// JSONValidateError error during validation of JSON file
	JSONValidateError string
	// LogFormatError error resolving log format
	LogFormatError string
	// ObjectDefError error reading object definitions
	ObjectDefError string
	// ProfilingError error starting profiling
	ProfilingError string
	// MetricError error starting profiling
	MetricError string
	// OsError error when interacting with host OS
	OsError string
	// SummaryTypeError incorrect summary type
	SummaryTypeError string
	// MaxErrorsReachedError max errors limit was reached during execution
	MaxErrorsReachedError struct {
		SubError error
		Msg      string
	}
)

var (
	metricsPort      int
	metricsAddress   string
	metricsLabel     string
	metricsGroupings []string
	profTyp          string
	objDefFile       string
	regression       bool
)

// *** Custom errors ***

// Error implementation of Error interface
func (err JSONParseError) Error() string {
	return string(err)
}

// Error implementation of Error interface
func (err JSONValidateError) Error() string {
	return string(err)
}

// Error implementation of Error interface
func (err LogFormatError) Error() string {
	return string(err)
}

// Error implementation of Error interface
func (err ObjectDefError) Error() string {
	return string(err)
}

// Error implementation of Error interface
func (err ProfilingError) Error() string {
	return string(err)
}

// Error implementation of Error interface
func (err MetricError) Error() string {
	return string(err)
}

// Error implementation of Error interface
func (err OsError) Error() string {
	return string(err)
}

// Error incorrect summary type
func (err SummaryTypeError) Error() string {
	return string(err)
}

// Error max error limit reached
func (err MaxErrorsReachedError) Error() string {
	msg := ""
	switch subErr := errors.Cause(err.SubError).(type) {
	case *multierror.Error:
		msg = TruncatedMultiErrorMessage(subErr)
	default:
		msg = fmt.Sprint("1 error occurred:\n", subErr)
	}
	return fmt.Sprintf("%s\n%s\n", msg, err.Msg)
}

// *********************

// executeCmd represents the execute command
var executeCmd = &cobra.Command{
	Use:     "execute",
	Aliases: []string{"x"},
	Short:   "Execute gopherciser scenario towards a sense environment",
	Long:    `Execute gopherciser scenario towards a sense environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			var panicErr error = nil
			helpers.RecoverWithError(&panicErr)
			if panicErr != nil {
				_, _ = os.Stderr.WriteString(fmt.Sprintf("%+v", panicErr))
			}
		}()

		if execErr := execute(); execErr != nil {
			var errMsg string
			var exitCode int

			cause := errors.Cause(execErr)
			switch cErr := cause.(type) {
			case JSONParseError:
				errMsg = fmt.Sprint("JSONParseError: ", execErr)
				exitCode = ExitCodeJSONParseError
			case JSONValidateError:
				errMsg = fmt.Sprint("JSONValidateError: ", execErr)
				exitCode = ExitCodeJSONValidateError
			case LogFormatError:
				errMsg = fmt.Sprint("LogFormatError: ", execErr)
				exitCode = ExitCodeLogFormatError
			case ObjectDefError:
				errMsg = fmt.Sprint("ObjectDefError: ", execErr)
				exitCode = ExitCodeObjectDefError
			case ProfilingError:
				errMsg = fmt.Sprint("ProfilingError: ", execErr)
				exitCode = ExitCodeProfilingError
			case MetricError:
				errMsg = fmt.Sprint("MetricError: ", execErr)
				exitCode = ExitCodeMetricError
			case OsError:
				errMsg = fmt.Sprint("OsError: ", execErr)
				exitCode = ExitCodeOsError
			case SummaryTypeError:
				errMsg = fmt.Sprint("SummaryError: ", execErr)
				exitCode = ExitCodeSummaryTypeError
			case *multierror.Error:
				errMsg = TruncatedMultiErrorMessage(cErr)
				exitCode = MultiErrorCode(cErr)
			case MaxErrorsReachedError:
				errMsg = cErr.Error()
				switch subErr := errors.Cause(cErr.SubError).(type) {
				case *multierror.Error:
					exitCode = MultiErrorCode(subErr)
				default:
					exitCode = 1
				}
			default:
				// only one error
				errMsg = fmt.Sprint("1 error occurred:\n", execErr)
				exitCode = 1
			}

			_, _ = fmt.Fprintf(os.Stderr, "%s\n", errMsg)
			os.Exit(exitCode)
		}
	},
}

// TruncatedMultiErrorMessage to first message + error count
func TruncatedMultiErrorMessage(err *multierror.Error) string {
	errMsg := ""
	errCount := 0
	if err != nil {
		errCount = len(err.Errors)
		if errCount > 0 {
			errMsg = fmt.Sprintf("%d errors occurred:\nFirst error: %s", errCount, err.Errors[0].Error())
		}
	}
	return errMsg
}

// MultiErrorCode error count truncated to 0x7F
func MultiErrorCode(err *multierror.Error) int {
	if err == nil {
		return 0
	}
	errCount := len(err.Errors)
	if errCount > 0x7F {
		errCount = 0x7F
	}
	return errCount
}

func init() {
	RootCmd.AddCommand(executeCmd)
	AddAllSharedParameters(executeCmd)

	// Custom object definitions
	executeCmd.Flags().StringVarP(&objDefFile, "definitions", "d", "", `Custom object definitions and overrides.`)

	// Logging
	AddLoggingParameters(executeCmd)

	// Prometheus
	executeCmd.Flags().IntVar(&metricsPort, "metrics", 0, "Export via http prometheus metrics.")
	executeCmd.Flags().StringVar(&metricsAddress, "metricsaddress", "", "If set other than empty string then Push otherwise pull, will be appended by port.")
	executeCmd.Flags().StringVar(&metricsLabel, "metricslabel", "gopherciser", "The job label to use for push metrics")
	executeCmd.Flags().StringSliceVarP(&metricsGroupings, "metricsgroupingkey", "g", nil, "The grouping keys (in key=value form) to use for push metrics. Specify multiple times for more grouping keys.")
	executeCmd.Flags().BoolVar(&regression, "regression", false, "Log data needed to run regression analysis.")

	// profiling
	executeCmd.Flags().StringVar(&profTyp, "profile", "", profile.Help())
}

func execute() error {

	// === config section ===
	cfg, errUnmarshal := unmarshalConfigFile()
	if errUnmarshal != nil {
		return JSONParseError(errUnmarshal.Error())
	}

	if err := validateConfigAndPrintWarnings(cfg); err != nil {
		return JSONValidateError(err.Error())
	}

	// === logging section ===
	if err := ConfigOverrideLogSettings(cfg); err != nil {
		return errors.WithStack(err)
	}

	if cfg.Settings.LogSettings.Regression {
		cfg.Scheduler = scheduler.Regression()
	}

	// === object definition section ===
	if err := ReadObjectDefinitions(); err != nil {
		return err
	}

	// === profiling section ===
	if profTyp != "" {
		defer func() {
			if err := profile.Close(); err != nil { //safe to use even if profiler was not started
				_, _ = fmt.Fprintf(os.Stderr, "error closiung profiler: %v", err)
			}
		}()

		typ, profErr := profile.ResolveParameter(profTyp)
		if profErr != nil {
			return ProfilingError(profErr.Error())
		}
		profErr = profile.StartProfiler(typ)
		if profErr != nil {
			return ProfilingError(fmt.Sprint("failed to start profiler. ", profErr))
		}
	}

	// === Handle SIGINT ===
	// this could be replaced by
	// 	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	// when moving above go 1.15
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
	}()

	// If process is not killed 5 minutes after context cancelled, create hang.stack file and force quit.
	go func() {
		<-ctx.Done()
		killcontext, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		<-killcontext.Done()

		stackFile := fmt.Sprintf("%s_%d_hang.stack", path.Base(os.Args[0]), os.Getpid())

		_, _ = os.Stderr.WriteString("5 minutes passed since process was cancelled, creating stack file for debugging and force quitting!")

		buf := make([]byte, 1<<16)
		runtime.Stack(buf, true)

		if err := helpers.WriteToFile(stackFile, buf); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("failed to write %s: %v\n", stackFile, err))
		} else {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("stack file written to: %s\n", stackFile))
		}

		os.Exit(ExitCodeForceQuit)
	}()

	// === Prometheus section ===
	if metricsPort > 0 {
		// Check if push or pull by looking at whether address is set or not
		if metricsAddress == "" {
			//Prometheus metrics will be pulled from the endpoint /metrics
			err := buildmetrics.PullMetrics(ctx, metricsPort, scenario.RegisteredActions())
			if err != nil {
				return MetricError(fmt.Sprintf("failed to start prometheus : %s ", err))
			}
		} else {
			//Prometheus metrics will be pushed to a pushgateway
			err := buildmetrics.PushMetrics(ctx, metricsPort, metricsAddress, metricsLabel, metricsGroupings, scenario.RegisteredActions())
			if err != nil {
				return MetricError(fmt.Sprintf("failed to start prometheus : %s ", err))
			}
		}
	}

	// Data for variable templates
	templateData := struct {
		ConfigFile string
	}{strings.Split(filepath.Base(cfgFile), ".")[0]}

	// === start execution ===
	var msgErrorReachedMsg *string

	cfg.Cancel = func(msg string) {
		msgErrorReachedMsg = &msg
		cancel()
	}

	err := cfg.Execute(ctx, templateData)
	if msgErrorReachedMsg != nil {
		return MaxErrorsReachedError{
			Msg:      *msgErrorReachedMsg,
			SubError: err,
		}
	}
	return err
}

func ReadObjectDefinitions() error {
	if objDefFile != "" {
		if _, err := senseobjdef.OverrideFromFile(objDefFile); err != nil {
			return ObjectDefError(fmt.Sprintf("failed to read config from file<%s>): %v", objDefFile, err))
		}
	}

	return nil
}

func addEnumToBuf(buf *helpers.Buffer, k int, v string) {
	buf.WriteString("[")
	buf.WriteString(strconv.Itoa(k))
	buf.WriteString("]: ")
	buf.WriteString(v)
	buf.WriteString("\n")
}

func resolveLogFormat(param string) (config.LogFormatType, error) {
	var i int
	var err error

	// Do we have an int?
	if i, err = strconv.Atoi(param); err != nil {
		// it's a string
		i, err = config.LogFormatType(0).GetEnumMap().Int(param)
		if err != nil {
			return -1, errors.Wrapf(err, "failed to parse <%s> to log format", param)
		}
	}
	// it's an int

	// make sure it's a valid type
	_, err = config.LogFormatType(0).GetEnumMap().String(i)
	if err != nil {
		return -1, errors.Wrapf(err, "failed to parse <%s> to log format", param)
	}

	return config.LogFormatType(i), nil
}

func resolveSummaryType() (config.SummaryType, error) {
	if i, err := strconv.Atoi(summaryType); err != nil {
		// it's a string
		i, err = config.SummaryType(0).GetEnumMap().Int(summaryType)
		if err != nil {
			return config.SummaryTypeDefault, errors.Errorf("Summary type<%s> does not exist", summaryType)
		}
		return config.SummaryType(i), nil
	} else {
		// it's an int
		_, err = config.SummaryType(0).GetEnumMap().String(i)
		if err != nil {
			return config.SummaryTypeDefault, errors.Errorf("Summary type<%s> does not exist", summaryType)
		}
		return config.SummaryType(i), nil
	}
}
