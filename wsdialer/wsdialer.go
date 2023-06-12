package wsdialer

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
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
	"github.com/qlik-oss/gopherciser/globals"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	// ReConnectSettings settings for automatically reconnecting websocket
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
		// SetPending reconnect
		SetPending func()
		// UnsetPending reconnect
		UnsetPending func()
	}

	// WsDialer wraps gobwas websocket dialer
	WsDialer struct {
		gobwas.Dialer
		net.Conn
		// Type of websocket, will be used by DisconnectError
		Type string
		// Reconnect settings
		Reconnect ReConnectSettings
		// OnUnexpectedDisconnect triggers on disconnect of websocket
		OnUnexpectedDisconnect func()
		MaxFrameSize           int64

		url       *neturl.URL
		closed    chan struct{}
		closeLock sync.Mutex
	}

	// DisconnectError is sent on websocket disconnect
	DisconnectError struct {
		Type string
		Err  error
	}
)

const (
	// DefaultTimeout of websocker dialer
	DefaultTimeout = 30 * time.Second
)

var (
	// DefaultBackoff of reconnection
	DefaultBackoff = []float64{0.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0}

	closedChan = make(chan struct{}) // reusable closed channel
)

func init() {
	close(closedChan)
}

// Error implements error interface
func (err DisconnectError) Error() string {
	return fmt.Sprintf("Websocket<%s> disconnected (%s)", err.Type, err.Err)
}

// New Create new websocket dialer, use type to define a specific type which would be reported when getting a DisconnectError
func New(url *neturl.URL, httpHeader http.Header, cookieJar http.CookieJar, timeout time.Duration, allowUntrusted bool, wstype string, maxFrameSize int64) (*WsDialer, error) {
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

	wsHeader.Add("User-Agent", globals.UserAgent())

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
		url:          url,
		Type:         wstype,
		closed:       make(chan struct{}),
		MaxFrameSize: maxFrameSize,
	}

	return &dialer, nil
}

// Dial new websocket connection
func (dialer *WsDialer) Dial(ctx context.Context) error {
	_, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		// make sure we have a timeout on connect
		dialer.Dialer.Timeout = time.Minute
	}

	var err error
	var br *bufio.Reader
	dialer.Conn, br, _ /*hs*/, err = dialer.Dialer.Dial(ctx, dialer.url.String())
	if br != nil {
		gobwas.PutReader(br)
	}
	return errors.WithStack(err)
}

// WriteMessage Write message to a frame on the websocket
func (dialer *WsDialer) WriteMessage(messageType int, data []byte) error {
	return wsutil.WriteClientMessage(dialer, gobwas.OpCode(messageType), data)
}

// readMessage is copied from github.com/gobwas/wsutil package and modified with maxframesize parameter
func readMessage(r io.Reader, m []wsutil.Message, maxFrameSize int64) ([]wsutil.Message, error) {
	rd := wsutil.Reader{
		Source:    r,
		State:     gobwas.StateClientSide,
		CheckUTF8: true,
		OnIntermediate: func(hdr gobwas.Header, src io.Reader) error {
			bts, err := io.ReadAll(src)
			if err != nil {
				return err
			}
			m = append(m, wsutil.Message{OpCode: hdr.OpCode, Payload: bts})
			return nil
		},
		MaxFrameSize: maxFrameSize,
	}
	h, err := rd.NextFrame()
	if err != nil {
		return m, err
	}
	var p []byte
	if h.Fin {
		// No more frames will be read. Use fixed sized buffer to read payload.
		p = make([]byte, h.Length)
		// It is not possible to receive io.EOF here because Reader does not
		// return EOF if frame payload was successfully fetched.
		// Thus we consistent here with io.Reader behavior.
		_, err = io.ReadFull(&rd, p)
	} else {
		// Frame is fragmented, thus use ioutil.ReadAll behavior.
		var buf bytes.Buffer
		var n int64
		n, err = buf.ReadFrom(&rd)
		fmt.Println("BYTES READ:", n)
		p = buf.Bytes()
	}
	if err != nil {
		return m, err
	}
	return append(m, wsutil.Message{OpCode: h.OpCode, Payload: p}), nil
}

// ReadMessage Read one entire message from websocket
func (dialer *WsDialer) ReadMessage() (int, []byte, error) {
	var msg []wsutil.Message
	var err error
	msg, err = readMessage(dialer, msg, dialer.MaxFrameSize)
	var data []byte

	var closeMsg []byte
	for _, m := range msg {
		if m.OpCode == gobwas.OpClose {
			closeMsg = append(closeMsg, m.Payload...)
		} else {
			data = append(data, m.Payload...)
		}
	}

	disconnected := false
	switch err := err.(type) {
	// We might want to check sub errors, but have seen both net.ErrClosed and os.SyscallError on disconnect to handling all OpErrors as disconnects for now
	case *net.OpError:
		disconnected = true
	default:
		if err == io.EOF {
			disconnected = true
		}
	}

	if len(closeMsg) > 3 {
		closeStatusCode := binary.BigEndian.Uint16(closeMsg[:2])
		closeReason := closeMsg[2:]
		err = fmt.Errorf("websocket closed with code<%d> and reason<%s>", closeStatusCode, closeReason)
		disconnected = true
	}

	if disconnected {
		if dialer.OnUnexpectedDisconnect != nil {
			dialer.OnUnexpectedDisconnect()
		}
		if !dialer.Reconnect.AutoReconnect {
			err = DisconnectError{Type: dialer.Type, Err: err}
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
				return len(data), data, DisconnectError{Type: dialer.Type, Err: err}
			}

			err = dialer.reconnect(motherContext, isClosed)
		}
	}

	return len(data), data, err
}

func (dialer *WsDialer) reconnect(ctx context.Context, isClosed func() bool) error {
	attempts := 0
	started := time.Now()

	var err error
	if dialer.Reconnect.SetPending != nil && dialer.Reconnect.UnsetPending != nil {
		dialer.Reconnect.SetPending()
		defer dialer.Reconnect.UnsetPending()
	}

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
		helpers.WaitFor(ctx, time.Duration(w*float64(time.Second)))
		if isClosed() {
			return DisconnectError{Type: dialer.Type}
		}

		if helpers.IsContextTriggered(ctx) {
			return DisconnectError{Type: dialer.Type}
		}
		if dialer.IsClosed() {
			return DisconnectError{Type: dialer.Type}
		}

		// Attempt re-connect
		attempts = i + 1
		func() {
			dialCtx, cancel := context.WithTimeout(ctx, dialer.Timeout)
			defer cancel()
			err = dialer.Dial(dialCtx)
		}()
	}
	return err
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
