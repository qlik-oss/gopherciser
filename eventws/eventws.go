package eventws

import (
	"context"
	"net/http"
	neturl "net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/requestmetrics"
	"github.com/qlik-oss/gopherciser/wsdialer"
)

type (
	EventWebsocket struct {
		*EventHandler
		*wsdialer.WsDialer
	}

	TrafficLogger interface {
		Received([]byte)
		Opened()
	}

	TrafficMetricsLogger interface {
		SocketOpenMetric(url *neturl.URL, duration time.Duration)
	}
)

const WsType = "EventWebsocket"

// SetupEventSocket to listen for events, event listening will stop at listenContext done.
func SetupEventSocket(dialContext context.Context, listenContext context.Context, timeout time.Duration, cookieJar http.CookieJar, trafficLogger TrafficLogger, metricsLogger TrafficMetricsLogger,
	url *neturl.URL, httpHeader http.Header, allowUntrusted bool, requestMetrics *requestmetrics.RequestMetrics, currentActionState func() *action.State) (*EventWebsocket, error) {
	dialer, err := wsdialer.New(url, httpHeader, cookieJar, timeout, allowUntrusted, WsType, 0)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	dialStart := time.Now()
	if err := dialer.Dial(dialContext); err != nil {
		return nil, errors.WithStack(err)
	}
	if trafficLogger != nil {
		trafficLogger.Opened()
	}

	if metricsLogger != nil {
		metricsLogger.SocketOpenMetric(url, time.Since(dialStart))
	}

	eventWs := &EventWebsocket{
		WsDialer:     dialer,
		EventHandler: NewEventHandler(),
	}

	go func() {
		for {
			select {
			case <-listenContext.Done():
				return
			case <-eventWs.Closed():
				return
			default:
				_, message, err := dialer.ReadMessage()
				actionState := currentActionState()
				if err != nil && actionState != nil {
					if !helpers.IsContextTriggered(listenContext) && !dialer.IsClosed() {
						actionState.AddErrors(errors.WithStack(err))
					}
					return
				}
				receivedTS := time.Now()
				var receivedSize = int64(len(message))
				if receivedSize == 0 {
					continue
				}

				// Log traffic
				if trafficLogger != nil {
					trafficLogger.Received(message)
				}

				// update received data counter
				if requestMetrics != nil {
					if err := requestMetrics.UpdateReceived(receivedTS, receivedSize); err != nil {
						if actionState != nil {
							actionState.AddErrors(errors.WithStack(err))
						}
						return
					}
				}

				// traffic metrics and request statistics not applicable since we are not sending any requests
				eventWs.EventHandler.event(actionState, message)
			}
		}
	}()

	return eventWs, nil
}
