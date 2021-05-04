package config

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/synced"
)

type (
	HookCore struct {
		Url         string                      `json:"url"`
		Method      string                      `json:"method"`
		Payload     synced.Template             `json:"payload"`
		RespCodes   []int                       `json:"respcodes"`
		ContentType string                      `json:"contenttype"`
		Extractors  map[string]helpers.DataPath `json:"extractors"`
		Headers     synced.TemplateMap          `json:"headers"`
		// TODO StopOnError bool                   `json:"stoponerror"`
		// TODO response data extract and validation rules on response
		initOnce sync.Once
	}

	Hook struct {
		HookCore
	}

	miniTrafficLogger struct {
		LogEntry *logger.LogEntry
	}
)

func (hook *Hook) init() {
	hook.initOnce.Do(func() {
		if len(hook.RespCodes) < 1 {
			hook.RespCodes = []int{200, 201}
		}

		if hook.ContentType == "" {
			hook.ContentType = "application/json"
		}
	})
}

// UnmarshalJSON hook
func (hook *Hook) UnmarshalJSON(arg []byte) error {
	if err := jsonit.Unmarshal(arg, &hook.HookCore); err != nil {
		return err
	}
	hook.init()

	return nil
}

// Execute hook
func (hook *Hook) Execute(ctx context.Context, logEntry *logger.LogEntry, data *hookData, allowUntrusted bool) error {
	hook.init()

	payload, err := hook.Payload.ExecuteString(data)
	if err != nil {
		return errors.WithStack(err)
	}

	headers, err := hook.Headers.Execute(data)
	if err != nil {
		return errors.WithStack(err)
	}

	buf := helpers.GlobalBufferPool.Get()
	defer helpers.GlobalBufferPool.Put(buf)
	_, err = buf.WriteString(payload)
	if err != nil {
		return errors.WithStack(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.Url, buf)
	if err != nil {
		return errors.WithStack(err)
	}
	req.Header.Set("Content-Type", hook.ContentType)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// TODO try avoid creating client here
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return nil },
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: allowUntrusted,
			},
		},
	}

	trafficLogger := &miniTrafficLogger{
		LogEntry: logEntry,
	}

	session.LogTrafficOut(req, true, trafficLogger, logEntry, 0)
	resp, err := client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	session.LogTrafficIn(resp, trafficLogger, logEntry, 0)

	buf = helpers.GlobalBufferPool.Get()
	defer helpers.GlobalBufferPool.Put(buf)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return errors.WithStack(err)
	}

	if !hook.OkResponse(resp.StatusCode) {
		return errors.Errorf("unexpected hook response code<%d>, response body: %s", resp.StatusCode, buf.String())
	}

	if err := hook.ExtractData(buf.Bytes(), data); err != nil {
		return errors.Wrapf(err, "hook<%s> failed", hook.Url)
	}

	// TODO add response validators

	return nil
}

func (hook *Hook) ExtractData(source []byte, data *hookData) error {
	for k, v := range hook.Extractors {
		value, err := v.LookupNoQuotes(source)
		if err != nil {
			return errors.WithStack(err)
		}
		data.Vars[k] = string(value)
	}
	return nil
}

func (hook *Hook) OkResponse(respCode int) bool {
	for _, code := range hook.RespCodes {
		if code == respCode {
			return true
		}
	}
	return false
}

func (tl *miniTrafficLogger) Sent(message []byte) {
	tl.LogEntry.LogDetail(logger.TrafficLevel, string(message), "Sent")
}

func (tl *miniTrafficLogger) Received(message []byte) {
	tl.LogEntry.LogDetail(logger.TrafficLevel, string(message), "Received")
}
