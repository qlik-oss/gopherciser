package logger

import (
	"bytes"
	"testing"
)

func TestJsonWriter(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBufferString("")

	isClosed := false
	jsonLogger := CreateJSONLogger(buf, func() error {
		isClosed = true
		return nil
	})

	// Set info level
	jsonLogger.Writer.Level(InfoLevel)

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

	// Test rows
	expectedString := `{"zerologlevel":"info","AppName":"","AppGUID":"","Session":0,"SessionName":"","Thread":0,"User":"myuser1","Action":"myaction","ActionId":0,"Label":"","ObjectType":"","Details":"mydetails","Errors":12,"InfoType":"","Received":0,"RequestsSent":0,"Sent":0,"Stack":"","Warnings":0,"ResponseTime":0,"level":"info","Tick":0,"time":"0001-01-01T00:00:00Z","timestamp":"0001-01-01T00:00:00Z"}`
	expectedString += "\n"
	if err := jsonLogger.Writer.WriteMessage(msg); err != nil {
		t.Error()
	} else {
		verifyString(t, buf.String(), expectedString)
	}

	// Test ErrorLevel
	buf.Reset()
	jsonLogger.Writer.Level(ErrorLevel)
	expectedString = ""
	if err := jsonLogger.Writer.WriteMessage(msg); err != nil {
		t.Error()
	} else {
		verifyString(t, buf.String(), expectedString)
	}

	buf.Reset()
	msg.Level = ErrorLevel
	expectedString = `{"zerologlevel":"error","AppName":"","AppGUID":"","Session":0,"SessionName":"","Thread":0,"User":"myuser1","Action":"myaction","ActionId":0,"Label":"","ObjectType":"","Details":"mydetails","Errors":12,"InfoType":"","Received":0,"RequestsSent":0,"Sent":0,"Stack":"","Warnings":0,"ResponseTime":0,"level":"error","Tick":0,"time":"0001-01-01T00:00:00Z","timestamp":"0001-01-01T00:00:00Z"}`
	expectedString += "\n"
	if err := jsonLogger.Writer.WriteMessage(msg); err != nil {
		t.Error()
	} else {
		verifyString(t, buf.String(), expectedString)
	}

	// Test DebugLevel
	buf.Reset()
	jsonLogger.Writer.Level(InfoLevel)
	msg.Level = DebugLevel
	expectedString = ""
	if err := jsonLogger.Writer.WriteMessage(msg); err != nil {
		t.Error()
	} else {
		verifyString(t, buf.String(), expectedString)
	}

	buf.Reset()
	jsonLogger.Writer.Level(DebugLevel)
	expectedString = `{"zerologlevel":"debug","AppName":"","AppGUID":"","Session":0,"SessionName":"","Thread":0,"User":"myuser1","Action":"myaction","ActionId":0,"Label":"","ObjectType":"","Details":"mydetails","Errors":12,"InfoType":"","Received":0,"RequestsSent":0,"Sent":0,"Stack":"","Warnings":0,"ResponseTime":0,"level":"debug","Tick":0,"time":"0001-01-01T00:00:00Z","timestamp":"0001-01-01T00:00:00Z"}`
	expectedString += "\n"
	if err := jsonLogger.Writer.WriteMessage(msg); err != nil {
		t.Error()
	} else {
		verifyString(t, buf.String(), expectedString)
	}

	// Test close function
	if err := jsonLogger.Close(); err != nil {
		t.Error("Close failed:", err)
	}
	if !isClosed {
		t.Error("Test of close function failed")
	}
}
