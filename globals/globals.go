package globals

import (
	"github.com/qlik-oss/gopherciser/atomichandlers"
	"github.com/qlik-oss/gopherciser/globals/constant"
)

var (
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
)

// All event topics
var EventTopics = []string{
	constant.EventTopicOnAuthenticationInformation,
	constant.EventTopicOnConnected,
	constant.EventTopicOnMaxParallelSessionsExceeded,
	constant.EventTopicOnLicenseAccessDenied,
	constant.EventTopicOnLicenseAccessDeniedPendingUserSync,
	constant.EventTopicOnSessionClosed,
	constant.EventTopicOnSessionLoggedOut,
	constant.EventTopicOnSessionTimedOut,
	constant.EventTopicOnEngineWebsocketFailed,
	constant.EventTopicOnNoEngineAvailable,
	constant.EventTopicOnRepositoryWebsocketFailed,
	constant.EventTopicOnNoRepositoryAvailable,
	constant.EventTopicOnDataPrepServiceWebsocketFailed,
	constant.EventTopicOnNoDataPrepServiceAvailable,
	constant.EventTopicOnNoPrintingServiceAvailable,
	constant.EventTopicOnExcessLicenseAssignment,
}
