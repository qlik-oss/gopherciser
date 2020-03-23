package enigmahandlers

import (
	"github.com/qlik-oss/gopherciser/atomichandlers"
	"github.com/qlik-oss/gopherciser/globals"
	"github.com/qlik-oss/gopherciser/logger"
)

type (
	// trafficLogger implementation of enigma.trafficLogger interface
	TrafficLogger struct {
		LogEntry *logger.LogEntry
		Requests *atomichandlers.AtomicCounter
	}
)

// NewTrafficLogger create new instance of traffic logger with default values
func NewTrafficLogger(logEntry *logger.LogEntry) *TrafficLogger {
	var req atomichandlers.AtomicCounter
	return &TrafficLogger{
		LogEntry: logEntry,
		Requests: &req,
	}
}

// Opened socket was opened
func (tl *TrafficLogger) Opened() {
	tl.LogEntry.Log(logger.TrafficLevel, "Socket opened")
}

// Sent message sent on socket
func (tl *TrafficLogger) Sent(message []byte) {
	tl.LogEntry.LogDetail(logger.TrafficLevel, string(message), "Sent")
	if tl.Requests != nil {
		tl.Requests.Inc()
	}
	globals.Requests.Inc()
}

// Received message received on socket
func (tl *TrafficLogger) Received(message []byte) {
	tl.LogEntry.LogDetail(logger.TrafficLevel, string(message), "Received")
}

// Closed log socket closed
func (tl *TrafficLogger) Closed() {
	tl.LogEntry.Log(logger.TrafficLevel, "Socket Closed")
}

// RequestCount get current request count
func (tl *TrafficLogger) RequestCount() uint64 {
	return tl.Requests.Current()
}

// ResetRequestCount reset current request count
func (tl *TrafficLogger) ResetRequestCount() {
	tl.Requests.Reset()
}
