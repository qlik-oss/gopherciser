package enigmahandlers

import (
	"context"
	"net"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
)

// SenseDialer glue between net.Conn and enigma.Socket implementing required methods
type SenseDialer struct {
	net.Conn
}

// WriteMessage Write message to a frame on the websocket
func (dialer *SenseDialer) WriteMessage(messageType int, data []byte) error {
	return wsutil.WriteClientMessage(dialer, ws.OpCode(messageType), data)
}

// ReadMessage Read one entire message from websocket
func (dialer *SenseDialer) ReadMessage() (int, []byte, error) {
	var msg []wsutil.Message
	var err error
	msg, err = wsutil.ReadServerMessage(dialer, msg)
	var data []byte

	for _, m := range msg {
		data = append(data, m.Payload...)
	}

	return len(msg), data, err
}

// Close connection
func (dialer *SenseDialer) Close() error {
	return dialer.Conn.Close()
}

func setupDialer(dialer *enigma.Dialer, timeout time.Duration) {
	if timeout.Nanoseconds() < 1 {
		timeout = 30 * time.Second
	}
	dialer.CreateSocket = func(ctx context.Context, url string, httpHeader http.Header) (enigma.Socket, error) {
		u, err := neturl.Parse(url)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if dialer.Jar != nil {
			// cookie needs to be set using http not ws scheme
			switch u.Scheme {
			case "wss":
				u.Scheme = "https"
			case "ws":
				u.Scheme = "http"
			}
			cookies := dialer.Jar.Cookies(u)
			cookieStrings := make([]string, 0, len(cookies))
			for _, cookie := range cookies {
				if cookie.String() != "" {
					cookieStrings = append(cookieStrings, cookie.String())
				}
			}
			if len(cookieStrings) > 0 {
				httpHeader.Add("Cookie", strings.Join(cookieStrings, "; "))
			}
		}

		wsDialer := ws.Dialer{
			Timeout: timeout,
			Header:  ws.HandshakeHeaderHTTP(httpHeader),
			OnHeader: func(key, value []byte) error {
				if strings.ToLower(string(key)) == "set-cookie" {
					// http doesn't expose cookie parser so we need to fake a http response to have it parsed
					header := http.Header{}
					header.Add("Set-Cookie", string(value))
					response := http.Response{Header: header}
					cookies := response.Cookies()
					if dialer.Jar != nil {
						dialer.Jar.SetCookies(u, cookies)
					}
				}
				return nil
			},
			NetDial: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, addr)
			},
			TLSConfig: dialer.TLSClientConfig,
		}

		conn, _ /* br*/, _ /*hs*/, err := wsDialer.Dial(ctx, url)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return &SenseDialer{conn}, nil
	}
}
