package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestTSVWriter(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBufferString("")

	header := []string{
		FieldLevel,
		FieldAuthUser,
		FieldAction,
		FieldDetails,
		FieldErrors,
		FieldSuccess,
	}

	isClosed := false
	tsvLogger, errTSVLogger := CreateTSVLogger(header, buf, func() error {
		isClosed = true
		return nil
	})
	if errTSVLogger != nil {
		t.Fatal("Failed to create TSV Logger")
	}

	// Test header
	expectedString := strings.Join(header, "\t") + "\n"
	verifyString(t, buf.String(), expectedString)

	// Set info level
	tsvLogger.Writer.Level(InfoLevel)

	// Test rows
	buf.Reset()
	msg := &LogChanMsg{
		message{
			Level: InfoLevel,
		},
		SessionEntry{
			User: "myuser1",
		},
		ActionEntry{
			Action: "myaction",
		},
		&ephemeralEntry{
			Details: "mydetails",
			Errors:  12,
			Success: true,
		},
	}
	msgs := []string{
		InfoLevel.String(),
		"myuser1",
		"myaction",
		"mydetails",
		"12",
		"",
	}
	expectedString = strings.Join(msgs, "\t") + "\n"
	if err := tsvLogger.Writer.WriteMessage(msg); err != nil {
		t.Error()
	} else {
		verifyString(t, buf.String(), expectedString)
	}

	// Test result row
	buf.Reset()
	msg.Level = ResultLevel
	msgs[0] = ResultLevel.String()
	msgs[5] = "true"
	expectedString = strings.Join(msgs, "\t") + "\n"
	if err := tsvLogger.Writer.WriteMessage(msg); err != nil {
		t.Error()
	} else {
		verifyString(t, buf.String(), expectedString)
	}

	// Test ErrorLevel
	buf.Reset()
	tsvLogger.Writer.Level(ErrorLevel)
	msg.Level = InfoLevel
	expectedString = ""
	if err := tsvLogger.Writer.WriteMessage(msg); err != nil {
		t.Error()
	} else {
		verifyString(t, buf.String(), expectedString)
	}

	buf.Reset()
	msg.Level = ErrorLevel
	msgs[0] = ErrorLevel.String()
	msgs[5] = ""
	expectedString = strings.Join(msgs, "\t") + "\n"
	if err := tsvLogger.Writer.WriteMessage(msg); err != nil {
		t.Error()
	} else {
		verifyString(t, buf.String(), expectedString)
	}

	// Test DebugLevel
	buf.Reset()
	tsvLogger.Writer.Level(InfoLevel)
	msg.Level = DebugLevel
	expectedString = ""
	if err := tsvLogger.Writer.WriteMessage(msg); err != nil {
		t.Error()
	} else {
		verifyString(t, buf.String(), expectedString)
	}

	tsvLogger.Writer.Level(DebugLevel)
	msgs[0] = DebugLevel.String()
	expectedString = strings.Join(msgs, "\t") + "\n"
	if err := tsvLogger.Writer.WriteMessage(msg); err != nil {
		t.Error()
	} else {
		verifyString(t, buf.String(), expectedString)
	}

	if err := tsvLogger.Close(); err != nil {
		t.Error("Close failed:", err)
	}
	if !isClosed {
		t.Error("Test of close function failed")
	}
}

func verifyString(t *testing.T, value, expected string) {
	t.Helper()

	if value != expected {
		t.Errorf("VerifyString failed, expected<%s> got<%s>", expected, value)
	}
}
