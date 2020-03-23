package statistics

import (
	"github.com/qlik-oss/gopherciser/atomichandlers"
)

type (
	// Counters collecting statistics
	Counters struct {
		// totOpenedApps increases each time an app is opened
		totOpenedApps atomichandlers.AtomicCounter
		// totCreatedApps increases each time ann app is created
		totCreatedApps atomichandlers.AtomicCounter
	}
)

// OpenedApps total opened apps counted
func (counter *Counters) OpenedApps() uint64 {
	return counter.totOpenedApps.Current()
}

// OpenedApps total opened apps counted in global stats collector
func OpenedApps() uint64 {
	if !globalCollector.IsOn() {
		return 0
	}
	return globalCollector.OpenedApps()
}

// IncOpenedApps increase total opened apps counted by one
func (counter *Counters) IncOpenedApps() {
	counter.totOpenedApps.Inc()
}

// IncOpenedApps increase total opened apps counted by one in global stats collector
func IncOpenedApps() {
	if globalCollector.IsOn() {
		globalCollector.totOpenedApps.Inc()
	}
}

// CreatedApps total apps created count
func (counter *Counters) CreatedApps() uint64 {
	return counter.totCreatedApps.Current()
}

// CreatedApps total apps created count in global stats collector
func CreatedApps() uint64 {
	if !globalCollector.IsOn() {
		return 0
	}
	return globalCollector.CreatedApps()
}

// IncCreatedApps increase total created apps counted by one
func (counter *Counters) IncCreatedApps() {
	counter.totCreatedApps.Inc()
}

// IncCreatedApps increase total created apps counted by one in global stats collector
func IncCreatedApps() {
	if globalCollector.IsOn() {
		globalCollector.totCreatedApps.Inc()
	}
}
