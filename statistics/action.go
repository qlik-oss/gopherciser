package statistics

import (
	"github.com/qlik-oss/gopherciser/atomichandlers"
)

type (
	ActionStatsMap map[string]*ActionStats

	// ActionStats statistics collector for an Action
	ActionStats struct {
		name    string
		label   string
		appGuid string
		// RespAvg average response time for successful actions
		RespAvg *SampleCollector
		// Requests total count of requests sent within action
		Requests atomichandlers.AtomicCounter
		// ErrCount total amount of errors within action
		ErrCount atomichandlers.AtomicCounter
		// WarnCount total amount of warnings within action
		WarnCount atomichandlers.AtomicCounter
		// Sent total amount of sent bytes
		Sent atomichandlers.AtomicCounter
		// Received total amount of received bytes
		Received atomichandlers.AtomicCounter
		// Failed total amount of failed actions, can be compared with RespAvg.count for success rate
		Failed atomichandlers.AtomicCounter
	}
)

// NewActionStats creates a new action statistics collector
func NewActionStats(name, label, appGUID string) *ActionStats {
	return &ActionStats{
		name:    name,
		label:   label,
		appGuid: appGUID,
		RespAvg: NewSampleCollector(),
	}
}

// Name of action
func (action *ActionStats) Name() string {
	if action == nil {
		return ""
	}
	return action.name
}

// Label action label
func (action *ActionStats) Label() string {
	if action == nil {
		return ""
	}
	return action.label
}

// AppGUID in which action was performed
func (action *ActionStats) AppGUID() string {
	if action == nil {
		return ""
	}
	return action.appGuid
}
