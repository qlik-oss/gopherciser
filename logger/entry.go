package logger

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/qlik-oss/gopherciser/atomichandlers"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	// SessionEntry log fields living of entire session
	SessionEntry struct {
		User        string
		Thread      uint64
		Session     uint64
		SessionName string
		AppName     string
		AppGUID     string
	}

	// ActionEntry log fields living during action
	ActionEntry struct {
		Action     string
		Label      string
		ActionID   uint64
		ObjectType string
	}

	// EphemeralEntry log fields living during log entry only
	ephemeralEntry struct {
		ResponseTime int64
		Success      bool
		Warnings     uint64
		Errors       uint64
		Stack        string
		Sent         uint64
		Received     uint64
		Details      string
		InfoType     string
		RequestsSent uint64
	}

	//LogEntry entry used for logging
	LogEntry struct {
		logger  *Log
		Session *SessionEntry
		Action  *ActionEntry
		// logging interceptor return false to break.
		interceptors map[LogLevel]func(entry *LogEntry) bool
		mu           sync.Mutex
	}
)

var tickCounter = atomichandlers.AtomicCounter{}

// NewLogEntry new instance of LogEntry
func NewLogEntry(log *Log) *LogEntry {
	return &LogEntry{
		logger: log,
	}
}

// ShallowCopy log entry, creates new log entry with pointers to exact same data, but with a new mutex
func (entry *LogEntry) ShallowCopy() *LogEntry {
	newLogEntry := NewLogEntry(entry.logger)
	newLogEntry.interceptors = entry.interceptors
	return newLogEntry
}

// Logf write formatted log entry to log
func (entry *LogEntry) Logf(level LogLevel, format string, args ...interface{}) {
	if entry == nil {
		return
	}

	entry.log(level, fmt.Sprintf(format, args...), nil)
}

// Log write log entry to log
func (entry *LogEntry) Log(level LogLevel, args ...interface{}) {
	if entry == nil {
		return
	}

	entry.log(level, fmt.Sprint(args...), nil)
}

// LogTrafficMetric log traffic metric entry
func (entry *LogEntry) LogTrafficMetric(responseTime int64, sent, received uint64, requestID int, method, params, trafficType, msg string) {
	if entry == nil || entry.logger == nil || !entry.logger.Settings.Metrics {
		return
	}

	separator := rune('\u001e') // 30 is RS (Record Separator)
	var reqIDString string

	if requestID != -1 {
		reqIDString = strconv.Itoa(requestID)
	}

	buf := helpers.NewBuffer()
	buf.WriteString(trafficType)
	buf.WriteRune(separator)
	buf.WriteString(reqIDString)
	buf.WriteRune(separator)
	buf.WriteString(method)
	buf.WriteRune(separator)
	buf.WriteString(params)

	var details string
	if buf.Error == nil {
		details = buf.String()
	}

	entry.log(MetricsLevel, msg, &ephemeralEntry{
		ResponseTime: responseTime,
		Sent:         sent,
		Received:     received,
		Details:      details,
	})
}

// LogDetail log message with detail
func (entry *LogEntry) LogDetail(level LogLevel, msg, detail string) {
	if entry == nil {
		return
	}

	entry.log(level, msg, &ephemeralEntry{
		Details: detail,
	})
}

// LogResult log result entry
func (entry *LogEntry) LogResult(success bool, warnings, errors, sent, received, requests uint64, responsetime int64, details string) {
	if entry == nil {
		return
	}

	entry.log(ResultLevel, "", &ephemeralEntry{
		Success:      success,
		Warnings:     warnings,
		Errors:       errors,
		Sent:         sent,
		Received:     received,
		RequestsSent: requests,
		ResponseTime: responsetime,
		Details:      details,
	})
}

// LogInfo log info entry
func (entry *LogEntry) LogInfo(infoType, msg string) {
	if entry == nil {
		return
	}

	entry.log(InfoLevel, msg, &ephemeralEntry{
		InfoType: infoType,
	})
}

// LogErrorReport log warning and error count
func (entry *LogEntry) LogErrorReport(reportType string, errors, warnings uint64) {
	if entry == nil {
		return
	}

	entry.Action = nil
	entry.log(InfoLevel, fmt.Sprintf("%d", errors+warnings), &ephemeralEntry{
		InfoType: reportType,
		Errors:   errors,
		Warnings: warnings,
	})
}

// LogError log error entry
func (entry *LogEntry) LogError(err error) {
	if entry == nil {
		return
	}

	entry.log(ErrorLevel, fmt.Sprintf("%s", err), &ephemeralEntry{
		Stack: fmt.Sprintf("%+v", err),
	})
}

// LogErrorWithMsg log error entry with message
func (entry *LogEntry) LogErrorWithMsg(msg string, err error) {
	if entry == nil {
		return
	}
	entry.log(ErrorLevel, msg, &ephemeralEntry{
		Stack: fmt.Sprintf("%+v", err),
	})
}

// LogDebug log debug entry with message
func (entry *LogEntry) LogDebug(msg string) {
	if !entry.ShouldLogDebug() {
		return
	}
	entry.log(DebugLevel, msg, nil)
}

// LogDebugf log debug entry with message
func (entry *LogEntry) LogDebugf(format string, args ...interface{}) {
	if !entry.ShouldLogDebug() {
		return
	}
	entry.log(DebugLevel, fmt.Sprintf(format, args...), nil)
}

func (entry *LogEntry) LogRegression(id string, data interface{}, meta map[string]interface{}) error {
	return entry.logger.regressionLogger.Log(id, data, meta)
}

func (entry *LogEntry) log(level LogLevel, msg string, eph *ephemeralEntry) {
	if entry == nil {
		return
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	m := message{
		Level:   level,
		Time:    time.Now(),
		Message: msg,
		Tick:    tickCounter.Inc(),
	}

	var s SessionEntry
	if entry.Session != nil {
		s = *entry.Session
	} else {
		s = SessionEntry{}
	}

	var a ActionEntry
	if entry.Action != nil {
		a = *entry.Action
	} else {
		a = ActionEntry{}
	}

	if eph == nil {
		eph = &ephemeralEntry{}
	}

	chanMsg := &LogChanMsg{m, s, a, eph}

	if entry.interceptors[level] != nil {
		if !entry.interceptors[level](entry) {
			return
		}
	}

	go entry.queueWrite(chanMsg)
}

func (entry *LogEntry) queueWrite(msg *LogChanMsg) {
	if entry == nil || entry.logger == nil {
		return
	}

	entry.logger.logChan <- msg
}

// SetSessionEntry set new session entry
func (entry *LogEntry) SetSessionEntry(s *SessionEntry) {
	if entry == nil {
		return
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	entry.Session = s
}

// SetActionEntry set new session entry
func (entry *LogEntry) SetActionEntry(a *ActionEntry) {
	if entry == nil {
		return
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	entry.Action = a
}

// AddInterceptor to log entry, return false to avoid logging
func (entry *LogEntry) AddInterceptor(level LogLevel, f func(entry *LogEntry) bool) {
	if entry == nil {
		return
	}

	if entry.interceptors == nil {
		entry.interceptors = make(map[LogLevel]func(entry *LogEntry) bool)
	}

	entry.interceptors[level] = f
}

// ShouldLogTraffic should traffic be logged
func (entry *LogEntry) ShouldLogTraffic() bool {
	if entry == nil || entry.logger == nil {
		return false
	}
	return entry.logger.Settings.Traffic
}

// ShouldLogTrafficMetrics should traffic metrics be logged
func (entry *LogEntry) ShouldLogTrafficMetrics() bool {
	if entry == nil || entry.logger == nil {
		return false
	}
	return entry.logger.Settings.Metrics
}

// ShouldLogDebug should debug info be logged
func (entry *LogEntry) ShouldLogDebug() bool {
	if entry == nil || entry.logger == nil {
		return false
	}
	return entry.logger.Settings.Debug
}

// ShouldLogRegression should regression data be logged
func (entry *LogEntry) ShouldLogRegression() bool {
	return entry != nil && entry.logger != nil && entry.logger.Settings.Regression
}
