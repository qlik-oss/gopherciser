package statistics

import "github.com/qlik-oss/gopherciser/atomichandlers"

type (
	// ExecutionCounters counts values during a execution
	ExecutionCounters struct {
		// Threads - Total started threads
		Threads atomichandlers.AtomicCounter
		// Sessions - Total started sessions
		Sessions atomichandlers.AtomicCounter
		// Users - Total unique users
		Users atomichandlers.AtomicCounter
		// Errors - Total errors
		Errors atomichandlers.AtomicCounter
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
	}
)
