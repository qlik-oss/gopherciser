package logger

import (
	"context"
	"strings"
	"testing"
	"time"
)

type (
	stringWriter struct {
		b             []byte
		writeNotifier chan struct{}
	}
)

func (s *stringWriter) Write(p []byte) (int, error) {
	s.b = append(s.b, p...)
	go s.notify()
	return len(p), nil
}

func (s *stringWriter) Reset() {
	s.b = nil
}

func (s *stringWriter) String() string {
	return string(s.b)
}

func (s *stringWriter) notify() {
	s.writeNotifier <- struct{}{}
}

func (s *stringWriter) wait(t *testing.T, timeoutExpected bool) {
	t.Helper()
	select {
	case <-s.writeNotifier:
		if timeoutExpected {
			t.Fatal("unexpected log row written:", s.String())
		}
	case <-time.After(2 * time.Millisecond):
		if !timeoutExpected {
			t.Fatal("timeout while waiting for log to be written")
		}
	}
}

func TestLogging(t *testing.T) {
	header := []string{
		FieldLevel,
		FieldAction,
		FieldAuthUser,
		FieldMessage,
		FieldDetails,
	}
	buf := &stringWriter{
		writeNotifier: make(chan struct{}),
	}

	// Create logger
	lgr, errLogCreate := CreateTSVLogger(header, buf, nil)
	if errLogCreate != nil {
		t.Fatal("Failed to create TSV logger:", errLogCreate)
	}

	log := NewLog(LogSettings{
		Traffic: false,
	})

	log.AddLoggers(lgr)
	log.StartLogger(context.Background())
	buf.wait(t, false)

	// Verify written header
	expectedString := strings.Join(header, "\t") + "\n"
	verifyString(t, buf.String(), expectedString)

	// Test log row
	entry := NewLogEntry(log)
	if entry == nil {
		t.Fatal("Failed creating log entry")
	}

	buf.Reset()
	entry.SetActionEntry(&ActionEntry{
		Action: "myaction",
	})
	entry.SetSessionEntry(&SessionEntry{
		User: "myuser",
	})
	message := "mymessage"
	entry.LogDetail(InfoLevel, message, "mydetail")

	msgs := []string{
		InfoLevel.String(),
		"myaction",
		"myuser",
		message,
		"mydetail",
	}

	buf.wait(t, false)

	expectedString = strings.Join(msgs, "\t") + "\n"
	verifyString(t, buf.String(), expectedString)

	// test debug
	buf.Reset()
	entry.Log(DebugLevel, message)
	buf.wait(t, true)
	expectedString = ""
	verifyString(t, buf.String(), expectedString)

	buf.Reset()
	log.SetDebug()
	entry.Logf(DebugLevel, "%s", message)
	buf.wait(t, false)
	msgs[0] = DebugLevel.String()
	msgs[4] = ""
	expectedString = strings.Join(msgs, "\t") + "\n"
	verifyString(t, buf.String(), expectedString)

	//Test interceptor
	buf.Reset()
	interceptorTriggered := false
	entry.AddInterceptor(WarningLevel, func(msg *LogEntry) bool {
		interceptorTriggered = true
		return true
	})
	msgs[0] = WarningLevel.String()
	entry.log(WarningLevel, message, nil)
	buf.wait(t, false)
	if !interceptorTriggered {
		t.Error("Interceptor not triggered")
	}
	expectedString = strings.Join(msgs, "\t") + "\n"
	verifyString(t, buf.String(), expectedString)

	// Test avoiding logging through interceptor
	buf.Reset()
	entry.AddInterceptor(WarningLevel, func(msg *LogEntry) bool {
		return false
	})
	entry.Log(WarningLevel)
	buf.wait(t, true)
	expectedString = ""
	verifyString(t, buf.String(), expectedString)

	if err := log.Close(); err != nil {
		t.Error("Error closing logger:", err)
	}
}
