package globals

import (
	"github.com/qlik-oss/gopherciser/globals/constant"
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
