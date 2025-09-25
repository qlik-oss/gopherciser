package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/rs/zerolog"
)

type (
	// MsgWriter implement to write log entry
	MsgWriter interface {
		// WriteMessage to log
		WriteMessage(msg *LogChanMsg) error
		// Set log level
		Level(lvl LogLevel)
	}

	// Logger container for writer and close functions
	Logger struct {
		Writer     MsgWriter
		closeFuncs []func() error
	}

	// LogLevel of logging
	LogLevel int

	//Message container
	message struct {
		Tick    uint64
		Time    time.Time
		Level   LogLevel
		Message string
	}

	// LogChanMsg container for row to be logged
	LogChanMsg struct {
		message
		SessionEntry
		ActionEntry
		*ephemeralEntry
	}

	// LogSettings settings
	LogSettings struct {
		Traffic    bool
		Metrics    bool
		Debug      bool
		Regression bool
	}

	// Log main struct to keep track of and propagate log entries to loggers. Close finished will be signaled on Closed channel.
	Log struct {
		loggers []*Logger
		logChan chan *LogChanMsg

		Closed   chan struct{}
		Settings LogSettings

		regressionLogger RegressionLoggerCloser
		stopListen       func()
	}
)

// When adding a new level also:
// * Add it to the String function
// * Add it in the StartLogger switch case if not to be logged on info level
const (
	UnknownLevel LogLevel = iota
	ResultLevel
	ErrorLevel
	WarningLevel
	InfoLevel
	MetricsLevel
	TrafficLevel
	DebugLevel
)

func (l LogLevel) String() string {
	switch l {
	case ResultLevel:
		return "result"
	case ErrorLevel:
		return "error"
	case WarningLevel:
		return "warning"
	case InfoLevel:
		return "info"
	case DebugLevel:
		return "debug"
	case TrafficLevel:
		return "traffic"
	case MetricsLevel:
		return "metric"
	default:
		return "unknown"
	}
}

// NewLog instance
func NewLog(settings LogSettings) *Log {
	return &Log{
		logChan:  make(chan *LogChanMsg, 2000),
		Settings: settings,
		Closed:   make(chan struct{}),
	}
}

// NewLogger instance
func NewLogger(w MsgWriter) *Logger {
	return &Logger{
		Writer: w,
	}
}

// NewLogChanMsg create new LogChanMsg, to be used for testing purposes
func NewEmptyLogChanMsg() *LogChanMsg {
	return &LogChanMsg{message{},
		SessionEntry{},
		ActionEntry{},
		&ephemeralEntry{}}
}

// NewLogEntry create new LogEntry using current logger
func (log *Log) NewLogEntry() *LogEntry {
	return NewLogEntry(log)
}

// AddLoggers to be used for logging
func (log *Log) AddLoggers(loggers ...*Logger) {
	if log.loggers == nil {
		log.loggers = []*Logger{}
		log.loggers = append(log.loggers, loggers...)
		return
	}
	log.loggers = append(log.loggers, loggers...)
}

// SetRegressionLoggerFile to be used for logging regression data to file. The
// file name is chosen, using `backupName`, to match the name of the standard
// log file.
func (log *Log) SetRegressionLoggerFile(fileName string) error {
	fileName = strings.TrimSuffix(backupName(fileName), filepath.Ext(fileName)) + ".regression"
	f, err := NewWriter(fileName)
	if err != nil {
		return errors.WithStack(err)
	}
	log.regressionLogger = NewRegressionLogger(f, HeaderEntry{"ID_FORMAT", "sessionID.actionID.objectID"})
	return nil
}

// CloseWithTimeout waits for the context used in "Startlogger" to cancel or max duration of timeout
func (log *Log) CloseWithTimeout(timeout time.Duration) error {
	//wait for all logs to be written or max duration
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if log.stopListen != nil {
		log.stopListen() // signal to start logging shutdown procedure
	}
	WaitForChanClose(ctx, log.Closed) // wait for listener to be finished or timeout

	// close down log writers
	var mErr *multierror.Error
	if log.loggers != nil {
		for _, v := range log.loggers {
			if err := v.Close(); err != nil {
				mErr = multierror.Append(mErr, err)
			}
		}
		log.loggers = nil
	}
	if log.regressionLogger != nil {
		if err := log.regressionLogger.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing regressionLogger: %v\n", err)
		}
	}

	return errors.WithStack(helpers.FlattenMultiError(mErr))
}

// Close functions with default timeout of 5 minutes
func (log *Log) Close() error {
	return errors.WithStack(log.CloseWithTimeout(5 * time.Minute))
}

// StartLogger start async reading on log channel
func (log *Log) StartLogger(ctx context.Context) {
	go log.logListen(ctx)
}

func (log *Log) logListen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	log.stopListen = cancel
	defer cancel()
	for {
		select {
		case msg, ok := <-log.logChan:
			if log.onLogChanMsg(msg, ok) {
				return
			}
		case <-ctx.Done():
			for { // empty channel from messages, best effort
				select {
				case msg, ok := <-log.logChan:
					if log.onLogChanMsg(msg, ok) {
						return
					}
				case <-time.After(time.Millisecond * 200):
					close(log.Closed) // no straggling logs, stop reading and mark channel closed
					return
				}
			}
		}
	}
}

func (log *Log) onLogChanMsg(msg *LogChanMsg, ok bool) bool {
	if !ok {
		close(log.Closed) //Notify logger closed
		return true
	}

	for _, l := range log.loggers {
		if l == nil || l.Writer == nil {
			continue
		}
		if err := l.Writer.WriteMessage(msg); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error writing log: %v\n", err)
		}
	}
	return false
}

// Write log message, should be done in go routine to not block
func (log *Log) Write(msg *LogChanMsg) {
	if msg == nil {
		return
	}

	for _, l := range log.loggers {
		if l == nil || l.Writer == nil {
			continue
		}
		if err := l.Writer.WriteMessage(msg); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error writing log: %v\n", err)
		}
	}
}

// SetMetrics level on logging for all loggers
func (log *Log) SetMetrics() {
	if log == nil {
		return
	}
	for _, l := range log.loggers {
		l.Writer.Level(MetricsLevel)
	}
}

// SetTraffic level on logging for all loggers
func (log *Log) SetTraffic() {
	if log == nil {
		return
	}
	for _, l := range log.loggers {
		l.Writer.Level(TrafficLevel)
	}
}

// SetDebug level on logging for all loggers
func (log *Log) SetDebug() {
	if log == nil {
		return
	}

	for _, l := range log.loggers {
		l.Writer.Level(DebugLevel)
	}
}

// Close logger
func (logger *Logger) Close() error {
	if logger == nil {
		return nil
	}
	var mErr *multierror.Error
	if logger.closeFuncs != nil {
		for _, v := range logger.closeFuncs {
			if err := v(); err != nil {
				mErr = multierror.Append(mErr, err)
			}
		}
	}

	return errors.WithStack(helpers.FlattenMultiError(mErr))
}

// AddCloseFunc add sub logger close function to be called upon logger close
func (logger *Logger) AddCloseFunc(f func() error) {
	if logger == nil {
		return
	}
	if logger.closeFuncs == nil {
		logger.closeFuncs = []func() error{f}
		return
	}
	logger.closeFuncs = append(logger.closeFuncs, f)
}

// CreateStdoutJSONLogger create logger for JSON on terminal for later adding to loggers list
func CreateStdoutJSONLogger() *Logger {
	zerolog.LevelFieldName = "zerologlevel"
	zlgr := zerolog.New(os.Stdout)
	zlgr = zlgr.Level(zerolog.InfoLevel)
	jsonWriter := NewJSONWriter(&zlgr)

	return NewLogger(jsonWriter)
}

// CreateJSONLogger with io.Writer
func CreateJSONLogger(writer io.Writer, closeFunc func() error) *Logger {
	zerolog.LevelFieldName = "zerologlevel"
	zlgr := zerolog.New(writer)
	zlgr = zlgr.Level(zerolog.InfoLevel)
	jsonLogger := NewLogger(NewJSONWriter(&zlgr))
	if closeFunc != nil {
		jsonLogger.AddCloseFunc(closeFunc)
	}
	return jsonLogger
}

// CreateTSVLogger with io.Writer
func CreateTSVLogger(header []string, writer io.Writer, closeFunc func() error) (*Logger, error) {
	tsvWriter := NewTSVWriter(header, writer)
	tsvLogger := NewLogger(tsvWriter)
	if closeFunc != nil {
		tsvLogger.AddCloseFunc(closeFunc)
	}
	if err := tsvWriter.WriteHeader(); err != nil {
		return nil, errors.Wrap(err, "Failed writing TSV header")
	}
	return tsvLogger, nil
}

// CreateStdoutLogger create logger for JSON on terminal for later adding to loggers list
func CreateStdoutLogger() *Logger {
	zlgr := zerolog.New(zerolog.ConsoleWriter{
		Out:     os.Stdout,
		NoColor: false,
	})
	zlgr = zlgr.Level(zerolog.InfoLevel)
	jsonWriter := NewJSONWriter(&zlgr)

	return NewLogger(jsonWriter)
}

// CreateDummyLogger auto discarding all entries
func CreateDummyLogger() *Logger {
	dummyWriter := NewTSVWriter(nil, io.Discard)
	dummyLogger := NewLogger(dummyWriter)
	return dummyLogger
}

// WaitForChanClose which ever comes first context cancel or c closed. Returns instantly if channel is nil.
func WaitForChanClose(ctx context.Context, c chan struct{}) {
	if c == nil {
		return
	}
	select {
	case <-ctx.Done():
	case <-c:
	}
}
