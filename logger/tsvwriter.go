package logger

import (
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	// TSVWriter write log rows in tab separated format
	TSVWriter struct {
		Out    io.Writer
		Header []string
		lvl    LogLevel
		mu     sync.Mutex
	}
)

const (
	separator = "\t"
	rowEnd    = "\n" //TODO configurable or adaptable file ending?
)

var (
	replacer = strings.NewReplacer("\t", "    ", "\n", "\\n", "\r", "")
)

// NewTSVWriter instance
func NewTSVWriter(header []string, w io.Writer) *TSVWriter {
	return &TSVWriter{
		Header: header,
		Out:    w,
		lvl:    InfoLevel,
	}
}

// WriteHeader to Out
func (writer *TSVWriter) WriteHeader() error {
	if writer == nil {
		return errors.New("TSV writer is nil")
	}

	row := ""
	for i, v := range writer.Header {
		if i != 0 {
			row += separator
		}
		row += v
	}
	row += rowEnd

	if _, err := writer.Out.Write([]byte(row)); err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}

// WriteMessage implement MsgWriter interface
func (writer *TSVWriter) WriteMessage(msg *LogChanMsg) error {
	if writer == nil {
		return nil
	}
	if writer.Out == nil {
		return nil
	}

	if writer.noLog(msg.Level) {
		return nil
	}

	buf := helpers.NewBuffer()
	for i, v := range writer.Header {
		if i != 0 {
			buf.WriteString(separator)
		}

		switch v {
		case FieldAuthUser:
			buf.WriteString(replacer.Replace(msg.User))
		case FieldThread:
			buf.WriteString(strconv.FormatUint(msg.Thread, 10))
		case FieldSession:
			buf.WriteString(strconv.FormatUint(msg.Session, 10))
		case FieldSessionName:
			buf.WriteString(replacer.Replace(msg.SessionName))
		case FieldAppName:
			buf.WriteString(replacer.Replace(msg.AppName))
		case FieldAppGUID:
			buf.WriteString(replacer.Replace(msg.AppGUID))
		case FieldTick:
			buf.WriteString(strconv.FormatUint(msg.Tick, 10))
		case FieldAction:
			buf.WriteString(replacer.Replace(msg.Action))
		case FieldResponseTime:
			buf.WriteString(strconv.FormatInt(msg.ResponseTime, 10))
		case FieldSuccess:
			if msg.Level == ResultLevel {
				buf.WriteString(strconv.FormatBool(msg.Success))
			}
		case FieldWarnings:
			buf.WriteString(strconv.FormatUint(msg.Warnings, 10))
		case FieldErrors:
			buf.WriteString(strconv.FormatUint(msg.Errors, 10))
		case FieldStack:
			buf.WriteString(replacer.Replace(msg.Stack))
		case FieldSent:
			buf.WriteString(strconv.FormatUint(msg.Sent, 10))
		case FieldReceived:
			buf.WriteString(strconv.FormatUint(msg.Received, 10))
		case FieldLabel:
			buf.WriteString(replacer.Replace(msg.Label))
		case FieldActionID:
			buf.WriteString(strconv.FormatUint(msg.ActionID, 10))
		case FieldObjectType:
			buf.WriteString(replacer.Replace(msg.ObjectType))
		case FieldDetails:
			buf.WriteString(replacer.Replace(msg.Details))
		case FieldInfoType:
			buf.WriteString(replacer.Replace(msg.InfoType))
		case FieldRequestsSent:
			buf.WriteString(strconv.FormatUint(msg.RequestsSent, 10))
		case FieldTime:
			buf.WriteString(msg.Time.Format(time.RFC3339Nano))
		case FieldTimestamp:
			buf.WriteString(msg.Time.UTC().Format(time.RFC3339Nano))
		case FieldLevel:
			buf.WriteString(msg.Level.String())
		case FieldMessage:
			buf.WriteString(replacer.Replace(msg.Message))
		default:
			return errors.Errorf("TSVWriter: Unsupported field<%s>", v)
		}
	}
	buf.WriteString(rowEnd)
	buf.WriteTo(writer.Out)

	if buf.Error != nil {
		return errors.Wrap(buf.Error, "Failed to write TSV row")
	}

	return nil
}

// Level implement MsgWriter interface
func (writer *TSVWriter) Level(lvl LogLevel) {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	writer.lvl = lvl
}

func (writer *TSVWriter) noLog(lvl LogLevel) bool {
	if writer == nil {
		return true
	}
	writer.mu.Lock()
	defer writer.mu.Unlock()
	return writer.lvl < lvl
}
