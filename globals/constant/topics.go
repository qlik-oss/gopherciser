package constant

// Topics for session pushed messages
const (
	EventTopicOnAuthenticationInformation          = "OnAuthenticationInformation"
	EventTopicOnConnected                          = "OnConnected"
	EventTopicOnMaxParallelSessionsExceeded        = "OnMaxParallelSessionsExceeded"
	EventTopicOnLicenseAccessDenied                = "OnLicenseAccessDenied"
	EventTopicOnLicenseAccessDeniedPendingUserSync = "OnLicenseAccessDeniedPendingUserSync"
	EventTopicOnSessionClosed                      = "OnSessionClosed"
	EventTopicOnSessionLoggedOut                   = "OnSessionLoggedOut"
	EventTopicOnSessionTimedOut                    = "OnSessionTimedOut"
	EventTopicOnEngineWebsocketFailed              = "OnEngineWebsocketFailed"
	EventTopicOnNoEngineAvailable                  = "OnNoEngineAvailable"
	EventTopicOnRepositoryWebsocketFailed          = "OnRepositoryWebsocketFailed"
	EventTopicOnNoRepositoryAvailable              = "OnNoRepositoryAvailable"
	EventTopicOnDataPrepServiceWebsocketFailed     = "OnDataPrepServiceWebsocketFailed"
	EventTopicOnNoDataPrepServiceAvailable         = "OnNoDataPrepServiceAvailable"
	EventTopicOnNoPrintingServiceAvailable         = "OnNoPrintingServiceAvailable"
	EventTopicOnExcessLicenseAssignment            = "OnExcessLicenseAssignment"
)

// EventTopicOnConnected possible states
const (
	OnConnectedSessionCreated                    = "SESSION_CREATED"
	OnConnectedSessionAttached                   = "SESSION_ATTACHED"
	OnConnectedSessionErrorNoLicense             = "SESSION_ERROR_NO_LICENSE"
	OnConnectedSessionErrorLicenseReNew          = "SESSION_ERROR_LICENSE_RENEW"
	OnConnectedSessionErrorLimitExceeded         = "SESSION_ERROR_LIMIT_EXCEEDED"
	OnConnectedSessionErrorSecurityHeaderChanged = "SESSION_ERROR_SECURITY_HEADER_CHANGED"
	OnConnectedSessionAccessControlSetupFailure  = "SESSION_ACCESS_CONTROL_SETUP_FAILURE"
	OnConnectedSessionErrorAppAccessDenied       = "SESSION_ERROR_APP_ACCESS_DENIED"
	OnConnectedSessionErrorAppFailure            = "SESSION_ERROR_APP_FAILURE"
)
