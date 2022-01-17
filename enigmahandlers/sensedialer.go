package enigmahandlers

import (
	"context"
	"net/http"
	neturl "net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/wsdialer"
)

type (
	// SenseDialer glue between net.Conn and enigma.Socket implementing required methods
	SenseDialer struct {
		*wsdialer.WsDialer
	}
)

const (
	// SenseWsType defines websocket type, used for logging purposes
	SenseWsType = "SenseWebsocket"
)

func setupDialer(dialer *enigma.Dialer, timeout time.Duration, logEntry *logger.LogEntry, onUnexpectedDisconnect func()) {
	dialer.CreateSocket = func(ctx context.Context, url string, httpHeader http.Header) (enigma.Socket, error) {
		nURL, err := neturl.Parse(url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		senseDialer := SenseDialer{}
		senseDialer.WsDialer, err = wsdialer.New(nURL, httpHeader, dialer.Jar, timeout, dialer.TLSClientConfig.InsecureSkipVerify, SenseWsType)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		senseDialer.OnUnexpectedDisconnect = onUnexpectedDisconnect

		if err := senseDialer.WsDialer.Dial(ctx); err != nil {
			return nil, errors.WithStack(err)
		}

		return senseDialer, nil
	}
}
