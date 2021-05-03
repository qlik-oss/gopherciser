package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
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
	}

	Hook struct {
		HookCore
	}
)

// UnmarshalJSON hook
func (hook *Hook) UnmarshalJSON(arg []byte) error {
	if err := jsonit.Unmarshal(arg, &hook.HookCore); err != nil {
		return err
	}
	hook.SetDefaults()

	return nil
}

func (hook *Hook) SetDefaults() {
	// set default values
	if len(hook.RespCodes) < 1 {
		hook.RespCodes = []int{200, 201}
	}

	if hook.ContentType == "" {
		hook.ContentType = "application/json"
	}
}

// Execute hook
func (hook *Hook) Execute(data *hookData) error {
	payload, err := hook.Payload.ExecuteString(data)
	if err != nil {
		return errors.WithStack(err)
	}

	headers, err := hook.Headers.Execute(data)
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Println("SEND:", payload)

	buf := helpers.GlobalBufferPool.Get()
	defer helpers.GlobalBufferPool.Put(buf)
	_, err = buf.WriteString(payload)
	if err != nil {
		return errors.WithStack(err)
	}

	// TODO use real context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.Url, buf)
	if err != nil {
		return errors.WithStack(err)
	}
	req.Header.Set("Content-Type", hook.ContentType)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	allowUntrusted := true // TODO set from config

	// TODO add traffic logging
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

	resp, err := client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	buf = helpers.GlobalBufferPool.Get()
	defer helpers.GlobalBufferPool.Put(buf)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return errors.WithStack(err)
	}

	if !hook.OkResponse(resp.StatusCode) {
		return errors.Errorf("unexpected hook response code<%d>, response body: %s", resp.StatusCode, buf.String())
	}

	fmt.Println(resp.StatusCode, "REC:", buf.String())

	if err := hook.ExtractData(buf.Bytes(), data); err != nil {
		return errors.Wrapf(err, "hook<%s> failed", hook.Url)
	}

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
