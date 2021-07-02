package enigmahandlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v2"
	"github.com/qlik-oss/gopherciser/enigmainterceptors"
	"github.com/qlik-oss/gopherciser/globals"
	"github.com/qlik-oss/gopherciser/globals/constant"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/requestmetrics"
	"github.com/qlik-oss/gopherciser/senseobjects"
)

type (
	// SenseUplink handle sense connection for a user
	SenseUplink struct {
		Global     *enigma.Global
		CurrentApp *senseobjects.App
		Objects    ObjectsMap
		FieldCache FieldCache
		VarCache   VarCache
		Traffic    ITrafficLogger

		ctx                context.Context
		cancel             context.CancelFunc
		logEntry           *logger.LogEntry
		trafficMetrics     *requestmetrics.RequestMetrics
		failedConnectFuncs []func()

		MockMode bool
	}

	Cache struct {
		Field FieldCache
		Var   VarCache
	}

	// SenseConnection direct sense connection implementing IConnection interface
	SenseConnection struct {
		*SenseUplink
	}

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
)

const (
	//MaxRetries when engine aborts request
	MaxRetries = 3
)

var jsonit = jsoniter.ConfigCompatibleWithStandardLibrary

// NewSenseUplink SenseUplink constructor
func NewSenseUplink(ctx context.Context, logentry *logger.LogEntry, metrics *requestmetrics.RequestMetrics, trafficLogger ITrafficLogger) *SenseUplink {
	cCtx, cancel := context.WithCancel(ctx)

	return &SenseUplink{
		ctx:            cCtx,
		cancel:         cancel,
		trafficMetrics: metrics,
		logEntry:       logentry,
		Traffic:        trafficLogger,
		FieldCache:     NewFieldCache(),
		VarCache:       NewVarCache(),
	}
}

// Disconnect uplink
func (connection *SenseConnection) Disconnect() error {
	if connection == nil || connection.SenseUplink == nil {
		return nil
	}

	connection.SenseUplink.Disconnect()
	return nil
}

// SetSense implements IConnection interface
func (connection *SenseConnection) SetSense(uplink *SenseUplink) {

	if connection.SenseUplink != nil {
		connection.SenseUplink.Disconnect()
	}

	connection.SenseUplink = uplink
}

// Sense implements IConnection interface
func (connection *SenseConnection) Sense() *SenseUplink {
	return connection.SenseUplink
}

// Connect connect to sense environment
func (uplink *SenseUplink) Connect(ctx context.Context, url string, headers http.Header, cookieJar http.CookieJar, allowUntrusted bool, timeout time.Duration) error {
	if uplink.Global != nil {
		uplink.Global.DisconnectFromServer()
		uplink.Global = nil
	}

	dialer := enigma.Dialer{
		MockMode: uplink.MockMode,
		Interceptors: []enigma.Interceptor{
			(&enigmainterceptors.MetricsHandler{
				Log: uplink.LogMetric,
			}).MetricsInterceptor,
			uplink.retryInterceptor,
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: allowUntrusted,
		},
	}
	if cookieJar != nil {
		dialer.Jar = cookieJar
	}
	dialer.TrafficLogger = uplink.Traffic

	onUnexpectedDisconnect := func() {
		if helpers.IsContextTriggered(uplink.ctx) {
			return
		}
		uplink.executeFailedConnectFuncs()
	}

	setupDialer(&dialer, timeout, uplink.logEntry, onUnexpectedDisconnect)

	// TODO somehow get better values for connect time
	startTimestamp := time.Now()
	global, err := dialer.Dial(ctx, url, headers)
	if err != nil {
		return errors.Wrap(err, "Error connecting to Sense")
	}

	// Log X-Qlik-Session if available
	u, err := neturl.Parse(url)
	if err != nil {
		uplink.logEntry.Logf(logger.WarningLevel, "error<%v> resolving url<%s>", err, url)
	} else if dialer.Jar != nil {
		// fake http/s else cookie jar won't return cookie
		switch u.Scheme {
		case "wss":
			u.Scheme = "https"
		case "ws":
			u.Scheme = "http"
		}
	}

	if err := uplink.trafficMetrics.Update(startTimestamp, time.Now(), 0, 0); err != nil {
		return errors.WithStack(err)
	}

	// Setup
	connectMsgChan := global.SessionMessageChannel(globals.EventTopics...)
	defer global.CloseSessionMessageChannel(connectMsgChan)

	// setup logging of traffic metrics for pushed events
	go func() {
		sessionChan := global.SessionMessageChannel()
		for {
			select {
			case event, ok := <-sessionChan:
				if !ok {
					return
				}
				// Metrics not triggered for pushed, update metrics here
				if err := uplink.trafficMetrics.UpdateReceived(time.Now(), int64(len(event.Content))); err != nil {
					uplink.logEntry.LogError(err)
				}
			case <-uplink.ctx.Done():
				return
			}
		}
	}()

	uplink.Global = global

	if !uplink.MockMode {
		// send a quick request, after this OnConnected and EventTopicOnAuthenticationInformation has been done and websocket possibly force closed
		_, connectErr := global.EngineVersion(uplink.ctx)

		mustAuthenticate, onConnectedSessionState, otherTopics := emptyMsgChan(connectMsgChan, uplink.logEntry)

		if connectErr != nil { // couldn't get version check authentication information
			if mustAuthenticate != nil && *mustAuthenticate {
				return errors.Errorf("websocket connected, but authentication failed")
			}

			if onConnectedSessionState == nil && len(otherTopics) < 1 {
				return errors.Errorf("websocket connected, but no state created")
			}

			if onConnectedSessionState != nil {
				switch *onConnectedSessionState {
				case constant.OnConnectedSessionCreated, constant.OnConnectedSessionAttached:
					return nil // connected ok
				case constant.OnConnectedSessionErrorNoLicense, constant.OnConnectedSessionErrorLicenseReNew, constant.OnConnectedSessionErrorLimitExceeded,
					constant.OnConnectedSessionErrorSecurityHeaderChanged, constant.OnConnectedSessionAccessControlSetupFailure, constant.OnConnectedSessionErrorAppAccessDenied,
					constant.OnConnectedSessionErrorAppFailure: // known error states
					return errors.Errorf("error connecting to engine: %s", *onConnectedSessionState)
				default:
					uplink.logEntry.Logf(logger.WarningLevel, "unknown engine session state: %s", *onConnectedSessionState)
				}
			}

			// No OnConnected received, return list of "other topics
			if len(otherTopics) > 0 {
				return errors.Errorf("websocket connected, but received error topic/-s: %s", strings.Join(otherTopics, ","))
			}

			// no mustAuthenticate, no onConnectedSessionState, and no other topics, post the connectErr (although it's most likely just EOF...)
			return errors.Errorf("websocket connected, but got error on requesting version: %v", connectErr)
		}
	}

	return nil
}

// emptyMsgChan returns *mustAuthenticate, *onConnectedSessionState
func emptyMsgChan(msgChan chan enigma.SessionMessage, logEntry *logger.LogEntry) (*bool, *string, []string) {
	var mustAuthenticate *bool = nil
	var onConnectedSessionState *string = nil
	otherTopics := make([]string, 0, 1)

	for {
		select {
		case event, ok := <-msgChan:
			if !ok {
				return mustAuthenticate, onConnectedSessionState, otherTopics
			}
			switch event.Topic {
			case constant.EventTopicOnConnected:
				var onConnected OnConnected
				if err := jsonit.Unmarshal(event.Content, &onConnected); err != nil {
					logEntry.Log(logger.WarningLevel, "failed to unmarshal pushed onConnected message")
					return mustAuthenticate, onConnectedSessionState, otherTopics
				}
				logEntry.LogInfo("OnConnected", string(event.Content))
				onConnectedSessionState = &onConnected.SessionState
			case constant.EventTopicOnAuthenticationInformation:
				mustAuthenticate = handleOnAuthenticationInformation(event.Content, logEntry)
			default:
				otherTopics = append(otherTopics, event.Topic)
			}
		default: // nothing more in channel
			return mustAuthenticate, onConnectedSessionState, otherTopics
		}
	}
}

func handleOnAuthenticationInformation(content json.RawMessage, logEntry *logger.LogEntry) *bool {
	var onAuthInfo OnAuthenticationInformation
	if err := jsonit.Unmarshal(content, &onAuthInfo); err != nil {
		logEntry.Log(logger.WarningLevel, "failed to unmarshal pushed OnAuthenticationInformation message")
		return nil
	}
	return &onAuthInfo.MustAuthenticate
}

// Disconnect Sense connection
func (uplink *SenseUplink) Disconnect() {
	uplink.cancel()
	if uplink.Global != nil {
		uplink.Global.DisconnectFromServer()
	}
}

// AddNewObject to object list
func (uplink *SenseUplink) AddNewObject(handle int, t ObjectType, id string, enigmaobject interface{}) (*Object, error) {
	obj := NewObject(handle, t, id, enigmaobject)
	if err := uplink.Objects.AddObject(obj); err != nil {
		return nil, errors.WithStack(err)
	}
	return obj, nil
}

// OnUnexpectedDisconnect registers an function to be executed on unexpected disconnect
func (uplink *SenseUplink) OnUnexpectedDisconnect(f func()) {
	if uplink == nil || f == nil {
		return
	}

	if uplink.failedConnectFuncs == nil {
		uplink.failedConnectFuncs = make([]func(), 0, 1)
	}

	uplink.failedConnectFuncs = append(uplink.failedConnectFuncs, f)
}

func (uplink *SenseUplink) executeFailedConnectFuncs() {
	if uplink == nil || len(uplink.failedConnectFuncs) < 1 {
		return
	}
	for _, f := range uplink.failedConnectFuncs {
		f()
	}
}

// LogMetric async log metric, this is injected into the enigma dialer and is responsible for recording
// message sent and received times, these times are used to record response times both for individual
// requests and entire actions
func (uplink *SenseUplink) LogMetric(invocation *enigma.Invocation, metrics *enigma.InvocationMetrics, result *enigma.InvocationResponse) {
	requestID := -1
	if result != nil {
		requestID = result.RequestID
	}

	var method string
	var params string
	if invocation != nil {
		method = invocation.Method
		if invocation.RemoteObject != nil && strings.TrimSpace(invocation.RemoteObject.GenericId) != "" {
			buf := helpers.NewBuffer()
			buf.WriteString(method)
			buf.WriteString(" [")
			buf.WriteString(invocation.RemoteObject.GenericId)
			buf.WriteString("]")
			if buf.Error == nil {
				method = buf.String()
			}
		}
		if invocation.Params != nil {
			if jB, err := jsonit.Marshal(invocation.Params); err == nil && jB != nil {
				params = string(jB)
			}
		}
	}

	if err := uplink.trafficMetrics.Update(metrics.SocketWriteTimestamp, metrics.SocketReadTimestamp,
		int64(metrics.RequestMessageSize), int64(metrics.ResponseMessageSize)); err != nil {
		uplink.logEntry.LogError(err)
	}

	if uplink.Traffic != nil {
		uplink.logEntry.LogTrafficMetric(metrics.SocketReadTimestamp.Sub(metrics.SocketWriteTimestamp).Nanoseconds(),
			uint64(metrics.RequestMessageSize), uint64(metrics.ResponseMessageSize), requestID, method, params, "WS")
	}

	reqStall := metrics.SocketWriteTimestamp.Sub(metrics.InvocationRequestTimestamp)
	if reqStall > constant.MaxStallTime {
		uplink.logEntry.LogDetail(logger.WarningLevel, "WS request stall", strconv.FormatInt(reqStall.Nanoseconds(), 10))
	}

	respStall := metrics.InvocationResponseTimestamp.Sub(metrics.SocketReadTimestamp)
	if !metrics.InvocationRequestTimestamp.IsZero() && !metrics.SocketReadTimestamp.IsZero() && respStall > constant.MaxStallTime {
		uplink.logEntry.LogDetail(logger.WarningLevel, "WS response stall", strconv.FormatInt(reqStall.Nanoseconds(), 10))
	}

}

func (uplink *SenseUplink) retryInterceptor(ctx context.Context, invocation *enigma.Invocation,
	next enigma.InterceptorContinuation) *enigma.InvocationResponse {

	//Temporary check for code 15 in interceptor, we don't want to resend requests which have been
	//aborted e.g. during a wrapped selection or other cases where aborted is expected. Hence this
	//should be done somewhere more controlled
	var response *enigma.InvocationResponse
	var retries int
	for {
		response = next(ctx, invocation)
		if qixErr, ok := response.Error.(enigma.Error); ok && qixErr.Code() == constant.LocerrGenericAborted && retries < MaxRetries {
			if uplink != nil && uplink.logEntry != nil {
				uplink.logEntry.LogInfo("LocerrGenericAborted", fmt.Sprintf("%s %v", invocation.Method, invocation.Params))
			}
			retries++
			continue
		}
		if retries >= MaxRetries {
			if uplink != nil && uplink.logEntry != nil {
				uplink.logEntry.Logf(logger.WarningLevel, "max retries<%d> exceeded, request not re-sent, method: %s params: %v", MaxRetries, invocation.Method, invocation.Params)
			}
		}
		break
	}
	return response
}

func (uplink *SenseUplink) SetCurrentApp(appGUID string, doc *enigma.Doc) error {
	err := uplink.Objects.AddObject(&Object{
		Handle:       doc.ObjectInterface.Handle,
		Type:         ObjTypeApp,
		EnigmaObject: doc,
	})
	if err != nil {
		return err
	}
	uplink.CurrentApp = &senseobjects.App{
		GUID: appGUID,
		Doc:  doc,
	}
	return nil
}
