package statistics

import (
	"github.com/qlik-oss/gopherciser/atomichandlers"
	"sync"

	"github.com/pkg/errors"
)

type (
	StatsLevel int

	// Collector of statistics
	Collector struct {
		Actions      ActionStatsMap
		RestRequests RequestStatsMap
		Level        StatsLevel

		// totOpenedApps increases each time an app is opened
		totOpenedApps atomichandlers.AtomicCounter
		// totCreatedApps increases each time ann app is created
		totCreatedApps atomichandlers.AtomicCounter

		actionsLock  sync.RWMutex
		requestsLock sync.RWMutex
	}
)

// StatsLevel enum
const (
	StatsLevelNone StatsLevel = iota
	StatsLevelOn
	StatsLevelFull
)

// NewCollector of statistics
func NewCollector() *Collector {
	return &Collector{
		Actions:      make(ActionStatsMap),
		RestRequests: make(RequestStatsMap),
		actionsLock:  sync.RWMutex{},
		requestsLock: sync.RWMutex{},
	}
}

// GetOrAddActionStats from action map, returns nil if statistics is turned off
func (collector *Collector) GetOrAddActionStats(name, label, appGUID string) *ActionStats {
	if collector == nil || !collector.IsOn() {
		return nil
	}

	key := name + label + appGUID

	// Read with Read lock as multiple reader can acquire read lock simultaneously
	if stats := collector.readActionWithKey(key); stats != nil {
		return stats
	}

	// action not yet registered, acquire write lock and add
	defer collector.actionsLock.Unlock()
	collector.actionsLock.Lock()

	// check if other thread has registered action before we acquired write lock
	if stats, ok := collector.Actions[key]; ok {
		return stats
	}

	stats := NewActionStats(name, label, appGUID)
	collector.Actions[key] = stats
	return stats
}

func (collector *Collector) readActionWithKey(key string) *ActionStats {
	defer collector.actionsLock.RUnlock()
	collector.actionsLock.RLock()
	return collector.Actions[key]
}

// GetOrAddRequestStats from REST request map, returns nil if StatsLevel is lower than "full"
func (collector *Collector) GetOrAddRequestStats(method, path string) *RequestStats {
	if collector == nil || !collector.IsFull() {
		return nil
	}

	key := method + path

	// Read with Read lock as multiple reader can acquire read lock simultaneously
	if stats := collector.readRequestWithKey(key); stats != nil {
		return stats
	}

	// request not yet registered, acquire write lock and add
	defer collector.requestsLock.Unlock()
	collector.requestsLock.Lock()

	// check if other thread has registered request before we acquired write lock
	if stats, ok := collector.RestRequests[key]; ok {
		return stats
	}

	stats := NewRequestStats(method, path)
	collector.RestRequests[key] = stats
	return stats
}

func (collector *Collector) readRequestWithKey(key string) *RequestStats {
	defer collector.requestsLock.RUnlock()
	collector.requestsLock.RLock()
	return collector.RestRequests[key]
}

// SetLevel of statistics collected
func (collector *Collector) SetLevel(level StatsLevel) error {
	if collector == nil {
		return errors.New("can't set collector level, collector is nil")
	}
	collector.Level = level
	return nil
}

// IsOn statistics collection is turned on
func (collector *Collector) IsOn() bool {
	return collector != nil && collector.Level > StatsLevelNone
}

// IsFull full statistics collection is turned on
func (collector *Collector) IsFull() bool {
	return collector != nil && collector.Level == StatsLevelFull
}

// ForEachAction read lock map and execute function for each ActionStats entry
func (collector *Collector) ForEachAction(f func(stats *ActionStats)) {
	if collector == nil {
		return
	}
	defer collector.actionsLock.RUnlock()
	collector.actionsLock.RLock()

	for _, stats := range collector.Actions {
		f(stats)
	}
}

// ForEachRequest read lock map and execute function for each RequestStats entry
func (collector *Collector) ForEachRequest(f func(stats *RequestStats)) {
	if collector == nil {
		return
	}
	defer collector.requestsLock.RUnlock()
	collector.requestsLock.RLock()

	for _, stats := range collector.RestRequests {
		f(stats)
	}
}

// ActionsLen length of action stats map of collector
func (collector *Collector) ActionsLen() int {
	if collector == nil {
		return 0
	}
	return len(collector.Actions)
}

// RESTRequestLen length of REST requests stats map of collector
func (collector *Collector) RESTRequestLen() int {
	if collector == nil {
		return 0
	}
	return len(collector.RestRequests)
}

// OpenedApps total opened apps counted
func (collector *Collector) OpenedApps() uint64 {
	if collector == nil {
		return 0
	}
	return collector.totOpenedApps.Current()
}

// IncOpenedApps increase total opened apps counted by one
func (collector *Collector) IncOpenedApps() {
	if collector == nil {
		return
	}
	collector.totOpenedApps.Inc()
}

// CreatedApps total apps created count
func (collector *Collector) CreatedApps() uint64 {
	if collector == nil {
		return 0
	}
	return collector.totCreatedApps.Current()
}

// IncCreatedApps increase total created apps counted by one
func (collector *Collector) IncCreatedApps() {
	if collector == nil {
		return
	}
	collector.totCreatedApps.Inc()
}
