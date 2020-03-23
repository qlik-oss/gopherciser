package scenario

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

const nanosecond float64 = 0.000000001

type (
	// ThinkTimeSettings think time settings
	ThinkTimeSettings struct {
		helpers.DistributionSettings
	}
)

// Execute simulated think time
func (settings ThinkTimeSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	actionState.Details = settings.LogDetails()
	// Fake sent message to not trigger error in onResult interceptor
	if err := sessionState.RequestMetrics.UpdateSent(time.Now(), 0); err != nil {
		sessionState.LogEntry.Log(logger.WarningLevel, "Faking sent message in timer delay failed")
	}

	seconds, err := settings.DistributionSettings.GetSample(sessionState.Randomizer())
	delay := time.Duration(int(seconds*1000000000)) * time.Nanosecond
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}
	if seconds < nanosecond {
		actionState.AddErrors(errors.New("timer delay not set"))
		return
	}

	// "Think"
	select {
	case <-sessionState.BaseContext().Done():
		// returning withouh updating end time makes log result log info: aborted instead of result: true
		return
	case <-time.After(delay):
	}

	// Fake received message to not trigger error in onResult interceptor
	if err := sessionState.RequestMetrics.UpdateReceived(time.Now(), 0); err != nil {
		sessionState.LogEntry.Log(logger.WarningLevel, "Faking received message in timer delay failed")
	}
}

// Validate think time settings
func (settings ThinkTimeSettings) Validate() error {
	return settings.DistributionSettings.Validate()
}

// LogDetails log think time settings
func (settings ThinkTimeSettings) LogDetails() string {
	switch settings.DistributionSettings.Type {
	case helpers.StaticDistribution:
		return fmt.Sprintf("delay:%s", strconv.FormatFloat(settings.DistributionSettings.Delay, 'f', -1, 64))
	case helpers.UniformDistribution:
		return fmt.Sprintf("mean:%f;deviation:%f", settings.DistributionSettings.Mean, settings.DistributionSettings.Deviation)
	default:
		return ""
	}
}
