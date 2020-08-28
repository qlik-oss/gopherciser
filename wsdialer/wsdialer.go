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
	"time"

	gobwas "github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pkg/errors"
)

type (
	WsDialer struct {
		gobwas.Dialer
		net.Conn
		Type string

		url *neturl.URL
	}

	// DisconnectError is sent on websocket disconnect
	DisconnectError struct {
		Type string
	}
)

const (
	DefaultTimeout = 30 * time.Second
)

// Error implements error interface
func (err DisconnectError) Error() string {
	return fmt.Sprintf("Websocket<%s> disconnected", err.Type)
}

// New Create new websocket dialer, use type to define a specific type which would be reported when getting a DisconnectError
func New(url *neturl.URL, httpHeader http.Header, cookieJar http.CookieJar, timeout time.Duration, allowUntrusted bool, wstype string) (*WsDialer, error) {
	if timeout.Nanoseconds() < 1 {
		timeout = DefaultTimeout
	}

	wsHeader := httpHeader.Clone()
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
		url:  url,
		Type: wstype,
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
		err = DisconnectError{Type: dialer.Type}
		fmt.Println("DisconnectError, type:", dialer.Type, "err:", err)
	}

	return len(data), data, err
}

// Close connection
func (dialer *WsDialer) Close() error {
	return dialer.Conn.Close()
}
