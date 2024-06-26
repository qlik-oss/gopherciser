package enigmahandlers

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/globals/constant"
	"github.com/qlik-oss/gopherciser/logger"
)

type (
	// OnAuthenticationInformation content structure of OnAuthenticationInformation event
	OnAuthenticationInformation struct {
		// MustAuthenticate tells us if we are authenticated
		MustAuthenticate bool `json:"mustAuthenticate"`
	}

	// OnConnected content structure of EventTopicOnConnected event
	OnConnected struct {
		// SessionState received session state, possible states listed as constants in constant.EventTopicOnConnected*
		SessionState string `json:"qSessionState"`
	}

	topicsHandler struct {
		msgChannel              chan enigma.SessionMessage
		OnConnectedReceived     chan struct{}
		mustAuthenticate        *bool
		onConnectedSessionState *string
		otherTopics             map[string]string
	}
)

// NewTopicsHandler handles pusched topics on message channel
func NewTopicsHandler(channel chan enigma.SessionMessage) *topicsHandler {
	return &topicsHandler{
		msgChannel:          channel,
		OnConnectedReceived: make(chan struct{}),
		otherTopics:         make(map[string]string),
	}
}

// Start topics handler and return channel messaging when OnConnected topic is received
func (handler *topicsHandler) Start(logEntry *logger.LogEntry) {
	go func() {
		for event := range handler.msgChannel {
			switch event.Topic {
			case constant.EventTopicOnConnected:
				logEntry.LogInfo("OnConnected", string(event.Content))

				var onConnected OnConnected
				var sessionState string
				if err := json.Unmarshal(event.Content, &onConnected); err != nil {
					sessionState = constant.OnConnectedSessionSessionParseFailed
				} else {
					sessionState = onConnected.SessionState
				}
				handler.onConnectedSessionState = &sessionState
				close(handler.OnConnectedReceived)
			case constant.EventTopicOnAuthenticationInformation:
				var onAuthInfo OnAuthenticationInformation
				if err := json.Unmarshal(event.Content, &onAuthInfo); err != nil {
					logEntry.Log(logger.WarningLevel, "failed to unmarshal pushed OnAuthenticationInformation message")
					continue
				}
				handler.mustAuthenticate = &onAuthInfo.MustAuthenticate
			default:
				handler.otherTopics[event.Topic] = string(event.Content)
			}
		}
	}()
}

func (handler *topicsHandler) IsErrorState(reconnect bool, logEntry *logger.LogEntry) error {
	if handler.mustAuthenticate != nil && *handler.mustAuthenticate {
		return errors.Errorf("websocket connected, but authentication failed")
	}

	if handler.onConnectedSessionState != nil {
		switch *handler.onConnectedSessionState {
		case constant.OnConnectedSessionSessionParseFailed:
			return errors.New("failed to parse OnConnected pushed message")
		case constant.OnConnectedSessionCreated, constant.OnConnectedSessionAttached:
			if reconnect && *handler.onConnectedSessionState != constant.OnConnectedSessionAttached {
				return NoSessionOnConnectError{}
			}
			return nil // connected ok
		case constant.OnConnectedSessionErrorNoLicense, constant.OnConnectedSessionErrorLicenseReNew, constant.OnConnectedSessionErrorLimitExceeded,
			constant.OnConnectedSessionErrorSecurityHeaderChanged, constant.OnConnectedSessionAccessControlSetupFailure, constant.OnConnectedSessionErrorAppAccessDenied,
			constant.OnConnectedSessionErrorAppFailure: // known error states
			return errors.Errorf("error connecting to engine: %s", *handler.onConnectedSessionState)
		default:
			logEntry.Logf(logger.WarningLevel, "unknown engine session state: %s", *handler.onConnectedSessionState)
		}
	}

	// No OnConnected received, return list of "other" topics
	if len(handler.otherTopics) > 0 {
		var msg string
		for key, value := range handler.otherTopics {
			msg += fmt.Sprintf("<%s>(%s)", key, value)
		}
		return errors.Errorf("websocket connected, but received error topic/-s: %s", msg)
	}

	return nil
}
