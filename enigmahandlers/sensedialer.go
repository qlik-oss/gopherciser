package enigmahandlers

import (
	"context"
	"net/http"
	neturl "net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/wsdialer"
)

type (
	// SenseDialer glue between net.Conn and enigma.Socket implementing required methods
	SenseDialer struct {
		*wsdialer.WsDialer
	}
)

const (
	SenseWsType = "SenseWebsocket"
)

func setupDialer(dialer *enigma.Dialer, timeout time.Duration) {
	dialer.CreateSocket = func(ctx context.Context, url string, httpHeader http.Header) (enigma.Socket, error) {
		nUrl, err := neturl.Parse(url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		senseDialer := SenseDialer{}
		senseDialer.WsDialer, err = wsdialer.New(nUrl, httpHeader, dialer.Jar, timeout, dialer.TLSClientConfig.InsecureSkipVerify, SenseWsType)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if err := senseDialer.WsDialer.Dial(ctx); err != nil {
			return nil, errors.WithStack(err)
		}

		return senseDialer, nil
	}
}
