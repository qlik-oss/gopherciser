package enigmahandlers

import (
	"fmt"

	"github.com/qlik-oss/gopherciser/logger"
)

type (
	Direction        string
	Protocol         string
	regressionLogger struct {
		ITrafficLogger
		LogEntry *logger.LogEntry
	}
)

const (
	Sent     Direction = "Sent"
	Received Direction = "Received"
)

const (
	Unknown Protocol = "UNKNOWN"
	WS      Protocol = "WS"
	HTTP    Protocol = "HTTP"
)

func WithRegressionLog(trafficLogger ITrafficLogger, logEntry *logger.LogEntry) ITrafficLogger {
	return &regressionLogger{
		ITrafficLogger: trafficLogger,
		LogEntry:       logEntry,
	}
}

func (logger *regressionLogger) Sent(message []byte) {
	if logger.ITrafficLogger != nil {
		logger.ITrafficLogger.Sent(message)
	}
	LogRegression(logger.LogEntry, Sent, message)
}

func (logger *regressionLogger) Received(message []byte) {
	if logger.ITrafficLogger != nil {
		logger.ITrafficLogger.Received(message)
	}
	LogRegression(logger.LogEntry, Received, message)
}

func LogRegression(logEntry *logger.LogEntry, direction Direction, message []byte) {
	protocol := Unknown
	switch {
	case len(message) > 0 && message[0] == byte('{'):
		protocol = WS
	default:
		protocol = HTTP
	}
	detail := fmt.Sprintf("%s %s", direction, protocol)
	message = reduceMessage(protocol, message)
	logEntry.LogDetail(logger.RegressionLevel, string(message), detail)
}

func reduceMessage(protocol Protocol, message []byte) []byte {
	// TODO(atluq): do some processing of the message
	return message
}
