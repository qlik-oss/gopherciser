package logger

import (
	"time"

	"github.com/rs/zerolog"
)

type (
	// JSONWriter wrapper implementing MessageWriter for zerolog
	JSONWriter struct {
		writer *zerolog.Logger
	}
)

// NewJSONWriter wrapper implementing MessageWriter for zerolog
func NewJSONWriter(zlgr *zerolog.Logger) *JSONWriter {
	return &JSONWriter{
		writer: zlgr,
	}
}

// WriteMessage implement MsgWriter interface
func (jw *JSONWriter) WriteMessage(msg *LogChanMsg) error {
	var event *zerolog.Event
	switch msg.Level {
	case ErrorLevel:
		event = jw.writer.Error()
	case WarningLevel:
		event = jw.writer.Warn()
	case DebugLevel:
		event = jw.writer.Debug()
	case UnknownLevel:
		event = jw.writer.Info()
	default:
		event = jw.writer.Info()
	}

	setEntryFields(event, msg)

	return nil
}

// Level implement MsgWriter interface
func (jw *JSONWriter) Level(lvl LogLevel) {
	var zlgr zerolog.Logger
	switch lvl {
	case DebugLevel:
		zlgr = jw.writer.Level(zerolog.DebugLevel)
	case ErrorLevel:
		zlgr = jw.writer.Level(zerolog.ErrorLevel)
	default:
		zlgr = jw.writer.Level(zerolog.InfoLevel)
	}
	jw.writer = &zlgr
}

func setEntryFields(ev *zerolog.Event, msg *LogChanMsg) {
	if ev == nil || msg == nil {
		return
	}

	setSessionFields(ev, &msg.SessionEntry)
	setActionFields(ev, &msg.ActionEntry)
	setEphemeralFields(ev, msg.ephemeralEntry, msg.Level)
	setMessageFields(ev, &msg.message)
}

func setSessionFields(ev *zerolog.Event, s *SessionEntry) {
	ev.Str(FieldAppName, s.AppName)
	ev.Str(FieldAppGUID, s.AppGUID)
	ev.Uint64(FieldSession, s.Session)
	ev.Str(FieldSessionName, s.SessionName)
	ev.Uint64(FieldThread, s.Thread)
	ev.Str(FieldAuthUser, s.User)
}

func setActionFields(ev *zerolog.Event, a *ActionEntry) {
	ev.Str(FieldAction, a.Action)
	ev.Uint64(FieldActionID, a.ActionID)
	ev.Str(FieldLabel, a.Label)
	ev.Str(FieldObjectType, a.ObjectType)
}

func setEphemeralFields(ev *zerolog.Event, e *ephemeralEntry, level LogLevel) {
	ev.Str(FieldDetails, e.Details)
	ev.Uint64(FieldErrors, e.Errors)
	ev.Str(FieldInfoType, e.InfoType)
	ev.Uint64(FieldReceived, e.Received)
	ev.Uint64(FieldRequestsSent, e.RequestsSent)
	ev.Uint64(FieldSent, e.Sent)
	ev.Str(FieldStack, e.Stack)
	if level == ResultLevel {
		ev.Bool(FieldSuccess, e.Success)
	}
	ev.Uint64(FieldWarnings, e.Warnings)
	ev.Int64(FieldResponseTime, e.ResponseTime)
}

func setMessageFields(ev *zerolog.Event, m *message) {
	ev.Str(FieldLevel, m.Level.String())
	ev.Uint64(FieldTick, m.Tick)
	ev.Str(FieldTime, m.Time.Format(time.RFC3339Nano))
	ev.Str(FieldTimestamp, m.Time.UTC().Format(time.RFC3339Nano))
	ev.Msg(m.Message)
}
