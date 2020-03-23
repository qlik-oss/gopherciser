package statistics

import (
	"sync"

	"github.com/pkg/errors"
)

type (
	StatsLevel int

	// Collector of statistics
	Collector struct {
		Counters
		Actions      ActionStatsMap
		RestRequests RequestStatsMap
		Level        StatsLevel

		actionsLock  sync.RWMutex
		requestsLock sync.RWMutex
	}
)

var (
	// globalCollector global collector of statistics (allocated by SetGlobalLevel)
	globalCollector *Collector
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
		Counters:     Counters{},
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

// GetOrAddActionStats from action map of global collector, returns nil if statistics is turned off
func GetOrAddGlobalActionStats(name, label, appGUID string) *ActionStats {
	return globalCollector.GetOrAddActionStats(name, label, appGUID)
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

// GetOrAddGlobalRequestStats from REST request map of global collector, returns nil if StatsLevel is lower than "full"
func GetOrAddGlobalRequestStats(method, path string) *RequestStats {
	return globalCollector.GetOrAddRequestStats(method, path)
}

// SetLevel of statistics collected
func (collector *Collector) SetLevel(level StatsLevel) error {
	if collector == nil {
		return errors.New("can't set collector level, collector is nil")
	}
	collector.Level = level
	return nil
}

// SetGlobalLevel of statistics collected of global collector
func SetGlobalLevel(level StatsLevel) {
	if globalCollector == nil {
		globalCollector = NewCollector()
	}
	_ = globalCollector.SetLevel(level) // only error is when nil, and we allocate collector just before this
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

// ForEachAction locks map and execute function for each ActionStats entry of global collector
func ForEachAction(f func(stats *ActionStats)) {
	globalCollector.ForEachAction(f)
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

// ForEachRequest read lock map and execute function for each RequestStats entry of global collector
func ForEachRequest(f func(stats *RequestStats)) {
	globalCollector.ForEachRequest(f)
}

// DestroyGlobalCollector set global connector to nil, mostly to be used in tests.
func DestroyGlobalCollector() {
	globalCollector = nil
}

// GlobalActionsLen length of action stats map of global collector
func GlobalActionsLen() int {
	if globalCollector == nil {
		return 0
	}
	return len(globalCollector.Actions)
}

// GlobalRESTRequestLen length of REST requests stats map of global collector
func GlobalRESTRequestLen() int {
	if globalCollector == nil {
		return 0
	}
	return len(globalCollector.RestRequests)
}
