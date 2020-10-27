package enigmahandlers

import (
	"github.com/qlik-oss/gopherciser/atomichandlers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/statistics"
)

type (
	// trafficLogger implementation of enigma.trafficLogger interface
	TrafficLogger struct {
		LogEntry *logger.LogEntry
		Requests *atomichandlers.AtomicCounter
		Counters *statistics.ExecutionCounters
	}
)

// NewTrafficLogger create new instance of traffic logger with default values
func NewTrafficLogger(logEntry *logger.LogEntry, counters *statistics.ExecutionCounters) *TrafficLogger {
	var req atomichandlers.AtomicCounter
	return &TrafficLogger{
		LogEntry: logEntry,
		Requests: &req,
		Counters: counters,
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
		tl.Requests.Inc() // Increase local request counter
	}
	tl.Counters.Requests.Inc() // Increase execution wide request counter
	LogRegression(tl.LogEntry, Sent, message)
}

// Received message received on socket
func (tl *TrafficLogger) Received(message []byte) {
	tl.LogEntry.LogDetail(logger.TrafficLevel, string(message), "Received")
	LogRegression(tl.LogEntry, Recieved, message)
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
