package session

import (
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/logger"
)

// WarnOrError add error to action or log warning
func WarnOrError(actionState *action.State, logEntry *logger.LogEntry, failOnError bool, err error) {
	if failOnError {
		actionState.AddErrors(err)
		return
	}
	logEntry.Log(logger.WarningLevel, err.Error())
}
