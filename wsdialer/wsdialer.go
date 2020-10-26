package wsdialer

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	neturl "net/url"
	"strings"
	"sync"
	"time"

	gobwas "github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	ReConnectSettings struct {
		// AutoReconnect, set to automatically try to reconnect when disconnected, will be set to false on "Close"
		AutoReconnect bool
		// Backoff scheme in seconds for reconnecting websocket if "AutoReconnect" set to true. Defaults to
		Backoff []float64
		// OnReconnectStart triggers when reconnect start (only with AutoReconnect set)
		OnReconnectStart func()
		// OnReconnectDone triggers when reconnect is considered successful or failed
		// err : error during last attempt (nil of successful)
		// attempts: amount of attempts including successful attempt tried
		// timeSpent: total duration spent trying to re-connect
		OnReconnectDone func(err error, attempts int, timeSpent time.Duration)
		// GetContext context used to abort waiting during backoff and as a mother context to dial, defaults to background context
		GetContext func() context.Context
	}

	WsDialer struct {
		gobwas.Dialer
		net.Conn
		// Type of websocket, will be used by DisconnectError
		Type      string
		Reconnect ReConnectSettings

		url       *neturl.URL
		closed    chan struct{}
		closeLock sync.Mutex
	}

	// DisconnectError is sent on websocket disconnect
	DisconnectError struct {
		Type string
	}
)

const (
	DefaultTimeout = 30 * time.Second
)

var (
	DefaultBackoff = []float64{0.0, 2.0, 2.0, 2.0, 2.0}

	closedChan = make(chan struct{}) // reusable closed channel
)

func init() {
	close(closedChan)
}

// Error implements error interface
func (err DisconnectError) Error() string {
	return fmt.Sprintf("Websocket<%s> disconnected", err.Type)
}

// New Create new websocket dialer, use type to define a specific type which would be reported when getting a DisconnectError
func New(url *neturl.URL, httpHeader http.Header, cookieJar http.CookieJar, timeout time.Duration, allowUntrusted bool, wstype string) (*WsDialer, error) {
	if timeout.Nanoseconds() < 1 {
		timeout = DefaultTimeout
	}
	var wsHeader http.Header
	if httpHeader == nil {
		wsHeader = make(http.Header)
	} else {
		wsHeader = httpHeader.Clone()
	}

	// cookie needs to be set using http not ws scheme
	cookieUrl := *url
	if cookieJar != nil {
		switch cookieUrl.Scheme {
		case "wss":
			cookieUrl.Scheme = "https"
		case "ws":
			cookieUrl.Scheme = "http"
		}
		cookieUrl.Path = ""
		cookies := cookieJar.Cookies(&cookieUrl)
		cookieStrings := make([]string, 0, len(cookies))
		for _, cookie := range cookies {
			if cookie.String() != "" {
				cookieStrings = append(cookieStrings, cookie.String())
			}
		}
		if len(cookieStrings) > 0 {
			wsHeader.Add("Cookie", strings.Join(cookieStrings, "; "))
		}
	}

	dialer := WsDialer{
		Dialer: gobwas.Dialer{
			Timeout: timeout,
			Header:  gobwas.HandshakeHeaderHTTP(wsHeader),
			OnHeader: func(key, value []byte) error {
				if strings.ToLower(string(key)) == "set-cookie" {
					// http doesn't expose cookie parser so we need to fake a http response to have it parsed
					header := http.Header{}
					header.Add("Set-Cookie", string(value))
					response := http.Response{Header: header}
					cookies := response.Cookies()
					if cookieJar != nil {
						cookieJar.SetCookies(&cookieUrl, cookies)
					}
				}
				return nil
			},
			NetDial: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, addr)
			},
			TLSConfig: &tls.Config{
				InsecureSkipVerify: allowUntrusted,
			},
		},
		url:    url,
		Type:   wstype,
		closed: make(chan struct{}),
	}

	return &dialer, nil
}

func (dialer *WsDialer) Dial(ctx context.Context) error {
	var err error
	dialer.Conn, _ /*br*/, _ /*hs*/, err = dialer.Dialer.Dial(ctx, dialer.url.String())
	return errors.WithStack(err)
}

// WriteMessage Write message to a frame on the websocket
func (dialer *WsDialer) WriteMessage(messageType int, data []byte) error {
	return wsutil.WriteClientMessage(dialer, gobwas.OpCode(messageType), data)
}

// ReadMessage Read one entire message from websocket
func (dialer *WsDialer) ReadMessage() (int, []byte, error) {
	var msg []wsutil.Message
	var err error
	msg, err = wsutil.ReadServerMessage(dialer, msg)
	var data []byte

	for _, m := range msg {
		data = append(data, m.Payload...)
	}

	if err == io.EOF {
		if !dialer.Reconnect.AutoReconnect {
			err = DisconnectError{Type: dialer.Type}
		} else {
			var motherContext context.Context
			if dialer.Reconnect.GetContext != nil {
				motherContext = dialer.Reconnect.GetContext()
			} else {
				motherContext = context.Background()
			}

			isClosed := func() bool {
				if helpers.IsContextTriggered(motherContext) {
					return true
				}
				return dialer.IsClosed()
			}

			if isClosed() {
				return len(data), data, DisconnectError{Type: dialer.Type}
			}

			attempts := 0
			started := time.Now()
			if dialer.Reconnect.OnReconnectStart != nil {
				dialer.Reconnect.OnReconnectStart()
			}
			reConnectDone := dialer.Reconnect.OnReconnectDone
			if reConnectDone != nil {
				defer func() {
					reConnectDone(err, attempts, time.Since(started))
				}()
			}

			backoff := dialer.Reconnect.Backoff
			if len(backoff) < 1 {
				backoff = DefaultBackoff
			}
			for i, w := range backoff {
				// wait for defined time before attempting re-connect
				helpers.WaitFor(motherContext, time.Duration(w*float64(time.Second)))
				if isClosed() {
					return len(data), data, DisconnectError{Type: dialer.Type}
				}

				if helpers.IsContextTriggered(motherContext) {
					return len(data), data, DisconnectError{Type: dialer.Type}
				}
				if dialer.IsClosed() {
					return len(data), data, DisconnectError{Type: dialer.Type}
				}

				// Attempt re-connect
				attempts = i + 1
				func() {
					ctx, cancel := context.WithTimeout(motherContext, dialer.Timeout)
					defer cancel()
					err = dialer.Dial(ctx)
				}()
			}
		}
	}

	return len(data), data, err
}

// Close connection
func (dialer *WsDialer) Close() error {
	if dialer == nil {
		return nil
	}
	dialer.closeLock.Lock()
	defer dialer.closeLock.Unlock()
	if dialer.closed != nil {
		defer close(dialer.closed)
		dialer.closed = nil
	}

	dialer.Reconnect.AutoReconnect = false
	if dialer.Conn == nil {
		return nil
	}
	return dialer.Conn.Close()
}

// Closed returns chan which will be closed when Close() is triggered
func (dialer *WsDialer) Closed() <-chan struct{} {
	if dialer == nil || dialer.closed == nil {
		return closedChan
	}
	return dialer.closed
}

// IsClosed check if Close() has been triggered
func (dialer *WsDialer) IsClosed() bool {
	dialer.closeLock.Lock()
	defer dialer.closeLock.Unlock()
	return dialer.closed == nil
}
