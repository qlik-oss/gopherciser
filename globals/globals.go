package globals

import (
	"fmt"

	"github.com/qlik-oss/gopherciser/globals/constant"
	"github.com/qlik-oss/gopherciser/version"
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

var (
	userAgentString = fmt.Sprintf("gopherciser %s", version.Version)
)

func SetUserAgent(userAgent string) {
	userAgentString = userAgent
}

// UserAgent returns user-agent string to be set
func UserAgent() string {
	return userAgentString
}
