package statistics

import (
	"fmt"

	"github.com/qlik-oss/gopherciser/atomichandlers"
)

type (
	Errors struct {
		counter     atomichandlers.AtomicCounter
		maxErrors   uint64
		triggerFunc func(msg string)
	}

	// ExecutionCounters counts values during a execution
	ExecutionCounters struct {
		// Threads - Total started threads
		Threads atomichandlers.AtomicCounter
		// Sessions - Total started sessions
		Sessions atomichandlers.AtomicCounter
		// Users - Total unique users
		Users atomichandlers.AtomicCounter
		// Warnings - Total warnings
		Warnings atomichandlers.AtomicCounter
		// ActionID - Unique global action id
		ActionID atomichandlers.AtomicCounter
		// Requests - Total requests sent
		Requests atomichandlers.AtomicCounter
		// ActiveUsers - Currently active users
		ActiveUsers atomichandlers.AtomicCounter
		// AppCounter -  App counter for round robin access
		AppCounter atomichandlers.AtomicCounter
		// RestRequestID - Added to REST traffic log to connect Request and Response
		RestRequestID atomichandlers.AtomicCounter
		// StatisticsCollector optional collection of statistics
		StatisticsCollector *Collector

		// Errors - Total errors
		Errors Errors
	}
)

// SetMaxErrors set a function (e.g. execution cancel) to be triggered each time Inc is called with a value >= maxerrors
func (counters *ExecutionCounters) SetMaxErrors(maxErrors uint64, triggerFunc func(msg string)) {
	counters.Errors.maxErrors = maxErrors
	counters.Errors.triggerFunc = triggerFunc
}

// Current value of counter
func (errors Errors) Current() uint64 {
	return errors.counter.Current()
}

// // Inc increase counter by 1
func (errors Errors) Inc() uint64 {
	errCount := errors.counter.Inc()
	errors.checkMaxError(errCount)
	return errCount
}

// Add value to counter
func (errors Errors) Add(u uint64) uint64 {
	errCount := errors.counter.Add(u)
	errors.checkMaxError(errCount)
	return errCount
}

// Reset counter to 0
func (errors Errors) Reset() {
	errors.counter.Reset()
}

func (errors Errors) checkMaxError(errCount uint64) {
	if errors.triggerFunc != nil && errors.maxErrors > 0 && errCount >= errors.maxErrors {
		errors.triggerFunc(fmt.Sprintf("Max error count of %d surpassed, aborting execution!", errors.maxErrors))
	}
}
