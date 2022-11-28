package enigmahandlers

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
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

		ctx               context.Context
		cancel            context.CancelFunc
		logEntry          *logger.LogEntry
		trafficMetrics    *requestmetrics.RequestMetrics
		failedConnectFunc func()

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

	doNotRetry struct{}

	NoSessionOnConnectError struct{}
)

const (
	//MaxRetries when engine aborts request
	MaxRetries = 3
)

// Error to be returned on reconnect without session attached
func (err NoSessionOnConnectError) Error() string {
	return "websocket connected, but no session attached"
}

// ContextWithoutRetries creates a new context which disables retires for
// aborted ws requests.
func ContextWithoutRetries(ctx context.Context) context.Context {
	return context.WithValue(ctx, doNotRetry{}, true)
}

func retriesDisabled(ctx context.Context) bool {
	disabled, isBool := ctx.Value(doNotRetry{}).(bool)
	return isBool && disabled
}

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
func (uplink *SenseUplink) Connect(ctx context.Context, url string, headers http.Header, cookieJar http.CookieJar, allowUntrusted bool, timeout time.Duration, reconnect bool) error {
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
	postDialTimestamp := time.Now()
	uplink.Global = global

	// Setup
	connectMsgChan := global.SessionMessageChannel(globals.EventTopics...)
	defer global.CloseSessionMessageChannel(connectMsgChan)

	if err := uplink.trafficMetrics.Update(startTimestamp, postDialTimestamp, 0, 0); err != nil {
		return errors.WithStack(err)
	}

	if uplink.MockMode {
		return nil
	}

	topicshandler := NewTopicsHandler(connectMsgChan)
	topicshandler.Start(uplink.logEntry)

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

	select {
	case <-topicshandler.OnConnectedReceived:
	case <-time.After(timeout):
		return errors.Errorf("websocket connected, but no state created or attach (timeout)")
	case <-uplink.ctx.Done():
		return errors.Errorf("websocket connected, but no state created or attach (context cancelled)")
	}

	// send a quick request, after this OnConnected and EventTopicOnAuthenticationInformation has been done and websocket possibly force closed
	_, connectErr := global.EngineVersion(uplink.ctx)

	// By now topics should be received, first check topic errors, before errors on version message
	if err := topicshandler.IsErrorState(reconnect, uplink.logEntry); err != nil {
		return errors.WithStack(err)
	}

	return errors.Wrap(connectErr, "websocket connected, but got error on requesting version")
}

// Disconnect Sense connection
func (uplink *SenseUplink) Disconnect() {
	uplink.cancel()
	if uplink.Global != nil {
		uplink.Global.DisconnectFromServer()
	}
}

// IsConnected returns true if uplink is live
func (uplink *SenseUplink) IsConnected() bool {
	if uplink == nil || uplink.Global == nil {
		return false
	}
	select {
	case <-uplink.Global.Disconnected():
		return false
	default:
		return true
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

	uplink.failedConnectFunc = f
}

func (uplink *SenseUplink) executeFailedConnectFuncs() {
	if uplink == nil || uplink.failedConnectFunc == nil {
		return
	}
	uplink.failedConnectFunc()
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
			if jB, err := json.Marshal(invocation.Params); err == nil && jB != nil {
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
			uint64(metrics.RequestMessageSize), uint64(metrics.ResponseMessageSize), requestID, method, params, "WS", "")
	}

	reqStall := metrics.SocketWriteTimestamp.Sub(metrics.InvocationRequestTimestamp)
	if reqStall > constant.MaxStallTime {
		uplink.logEntry.LogDetail(logger.WarningLevel, "WS request stall", strconv.FormatInt(reqStall.Nanoseconds(), 10))
	}

	respStall := metrics.InvocationResponseTimestamp.Sub(metrics.SocketReadTimestamp)
	if !metrics.InvocationRequestTimestamp.IsZero() && !metrics.SocketReadTimestamp.IsZero() && respStall > constant.MaxStallTime {
		uplink.logEntry.LogDetail(logger.WarningLevel, "WS response stall", strconv.FormatInt(respStall.Nanoseconds(), 10))
	}

}

func (uplink *SenseUplink) retryInterceptor(ctx context.Context, invocation *enigma.Invocation,
	next enigma.InterceptorContinuation) *enigma.InvocationResponse {

	doNotRetry := retriesDisabled(ctx)

	var response *enigma.InvocationResponse
	var retries int
	for {
		response = next(ctx, invocation)
		if doNotRetry {
			break
		}
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
