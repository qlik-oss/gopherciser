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

	delay, err := settings.DistributionSettings.RandDuration(sessionState.Randomizer())
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}
	if delay < time.Nanosecond {
		actionState.AddErrors(errors.New("timer delay not set"))
		return
	}

	// "Think"
	select {
	case <-sessionState.BaseContext().Done():
		// returning without updating end time makes log result log info: aborted instead of result: true
		return
	case <-time.After(delay):
	}

	// Fake received message to not trigger error in onResult interceptor
	if err := sessionState.RequestMetrics.UpdateReceived(time.Now(), 0); err != nil {
		sessionState.LogEntry.Log(logger.WarningLevel, "Faking received message in timer delay failed")
	}

	sessionState.Wait(actionState) // wait for any requests triggered by pushed engine message
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
