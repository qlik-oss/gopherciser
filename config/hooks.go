package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/synced"
)

type (
	FailLevel      int
	ValidationType int
	HttpMethod     int

	ValidatorCore struct {
		Type  ValidationType `json:"type" doc-key:"hook.extractor.validator.type" displayname:"Type"`
		Value interface{}    `json:"value" doc-key:"hook.extractor.validator.value" displayname:"Value"`
	}

	Validator struct {
		ValidatorCore
	}

	Extractor struct {
		Name      string           `Json:"name" doc-key:"hook.extractor.name" displayname:"Name"`
		Path      helpers.DataPath `json:"path" doc-key:"hook.extractor.path" displayname:"Path"`
		Level     FailLevel        `json:"faillevel" doc-key:"hook.extractor.faillevel" displayname:"Fail level"`
		Validator *Validator       `json:"validator" doc-key:"hook.extractor.validator" displayname:"Validator"`
	}

	HookCore struct {
		Url         string             `json:"url" doc-key:"hook.url" displayname:"Url"`
		Method      HttpMethod         `json:"method" doc-key:"hook.method" displayname:"Method"`
		Content     synced.Template    `json:"payload" doc-key:"hook.content" displayname:"Content"`
		RespCodes   []int              `json:"respcodes" doc-key:"hook.respcodes" displayname:"Response codes"`
		ContentType string             `json:"contenttype" doc-key:"hook.contenttype" displayname:"Content-Type"`
		Extractors  []Extractor        `json:"extractors" doc-key:"hook.extractors" displayname:"Extractors"`
		Headers     synced.TemplateMap `json:"headers" doc-key:"hook.headers" displayname:"Headers"`

		initOnce sync.Once
	}

	Hook struct {
		HookCore
	}

	miniTrafficLogger struct {
		LogEntry *logger.LogEntry
	}
)

var (
	// FailLevel enumeration
	failLevelEnum = enummap.NewEnumMapOrPanic(map[string]int{
		"none":    int(FailLevelNone),
		"info":    int(FailLevelInfo),
		"warning": int(FailLevelWarning),
		"error":   int(FailLevelError),
	})

	// ValidationType enumeration
	validationTypeEnum = enummap.NewEnumMapOrPanic(map[string]int{
		"none":   int(ValidationTypeNone),
		"bool":   int(ValidationTypeBool),
		"number": int(ValidationTypeNumber),
		"string": int(ValidationTypeString),
	})

	httpMethodEnum = enummap.NewEnumMapOrPanic(map[string]int{
		strings.ToLower(http.MethodGet):     int(MethodGet),
		strings.ToLower(http.MethodHead):    int(MethodHead),
		strings.ToLower(http.MethodPost):    int(MethodPost),
		strings.ToLower(http.MethodPut):     int(MethodPut),
		strings.ToLower(http.MethodPatch):   int(MethodPatch),
		strings.ToLower(http.MethodDelete):  int(MethodDelete),
		strings.ToLower(http.MethodConnect): int(MethodConnect),
		strings.ToLower(http.MethodOptions): int(MethodOptions),
		strings.ToLower(http.MethodTrace):   int(MethodTrace),
	})
)

// FailLevel
const (
	FailLevelError FailLevel = iota
	FailLevelWarning
	FailLevelInfo
	FailLevelNone
)

// ValidationType
const (
	ValidationTypeNone ValidationType = iota
	ValidationTypeBool
	ValidationTypeNumber
	ValidationTypeString
)

const (
	MethodGet HttpMethod = iota
	MethodHead
	MethodPost
	MethodPut
	MethodPatch
	MethodDelete
	MethodConnect
	MethodOptions
	MethodTrace
)

// GetEnumMap of FailLevel for GUI
func (fl FailLevel) GetEnumMap() *enummap.EnumMap {
	return failLevelEnum
}

// GetEnumMap of ValidationType for GUI
func (val ValidationType) GetEnumMap() *enummap.EnumMap {
	return validationTypeEnum
}

// GetEnumMap of HttpMethod for GUI
func (method HttpMethod) GetEnumMap() *enummap.EnumMap {
	return httpMethodEnum
}

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

// UnmarshalJSON(arg []byte) error {
func (hook *Hook) UnmarshalJSON(arg []byte) error {
	if err := jsonit.Unmarshal(arg, &hook.HookCore); err != nil {
		return err
	}
	hook.init()

	return nil
}

// UnmarshalJSON Validator
func (validator *Validator) UnmarshalJSON(arg []byte) error {
	err := jsonit.Unmarshal(arg, &validator.ValidatorCore)
	if err != nil {
		return err
	}

	switch value := validator.Value.(type) {
	case string:
		switch validator.Type {
		case ValidationTypeBool:
			validator.Value, err = strconv.ParseBool(value)
			if err != nil {
				return errors.Errorf("value<%s> is not boolean", value)
			}
		case ValidationTypeNumber:
			validator.Value, err = strconv.ParseFloat(value, 64)
			if err != nil {
				return errors.Errorf("value<%s> is not a number", value)
			}
		}
	}

	return nil
}

// UnmarshalJSON HttpMethod
func (method *HttpMethod) UnmarshalJSON(arg []byte) error {
	i, err := httpMethodEnum.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal httpMethodEnum")
	}
	*method = HttpMethod(i)
	return nil
}

// MarshalJSON marshal HttpMethod
func (method HttpMethod) MarshalJSON() ([]byte, error) {
	str, err := httpMethodEnum.String(int(method))
	if err != nil {
		return nil, errors.Errorf("Unknown HttpMethod<%d>", method)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// UnmarshalJSON ValidationType
func (typ *ValidationType) UnmarshalJSON(arg []byte) error {
	i, err := validationTypeEnum.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal validationTypeEnum")
	}
	*typ = ValidationType(i)
	return nil
}

// MarshalJSON marshal ValidationType
func (typ ValidationType) MarshalJSON() ([]byte, error) {
	str, err := httpMethodEnum.String(int(typ))
	if err != nil {
		return nil, errors.Errorf("Unknown ValidationType<%d>", typ)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// UnmarshalJSON FailLevel
func (lvl *FailLevel) UnmarshalJSON(arg []byte) error {
	i, err := failLevelEnum.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal failLevelEnum")
	}
	*lvl = FailLevel(i)
	return nil
}

// MarshalJSON marshal ValidationType
func (lvl FailLevel) MarshalJSON() ([]byte, error) {
	str, err := httpMethodEnum.String(int(lvl))
	if err != nil {
		return nil, errors.Errorf("Unknown FailLevel<%d>", lvl)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// Validate hook settings, returns list of warnings or error
func (hook *Hook) Validate() ([]string, error) {
	if hook.Url == "" {
		return nil, errors.Errorf("hook defined with empty URL")
	}

	// Validate extractors
	for _, extractor := range hook.Extractors {
		if err := extractor.Validate(); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return nil, nil
}

// Validate extractor
func (extractor *Extractor) Validate() error {
	if extractor == nil {
		return nil
	}
	if extractor.Name == "" {
		return errors.New("no key name defined for extractor")
	}
	if extractor.Path.String() == "" {
		return errors.New("no path defined for extractor")
	}

	return extractor.Validator.Validate()
}

//Validate validator settings
func (validator *Validator) Validate() error {
	if validator == nil {
		return nil
	}

	switch validator.Type {
	case ValidationTypeBool:
		if _, isBool := validator.Value.(bool); !isBool {
			return errors.Errorf("validator value<%v> not of type bool", validator.Value)
		}
	case ValidationTypeNumber:
		if _, isFloat := validator.Value.(float64); !isFloat {
			return errors.Errorf("validator value<%v> not of type number", validator.Value)
		}
	case ValidationTypeString:
		if _, isString := validator.Value.(string); !isString {
			return errors.Errorf("validator value<%v> not of type string", validator.Value)
		}
	case ValidationTypeNone:
	default:
		return errors.Errorf("ValidationType<%v> not supported", validator.Type)
	}
	return nil
}

// Execute hook
func (hook *Hook) Execute(ctx context.Context, logEntry *logger.LogEntry, data *hookData, allowUntrusted bool) error {
	hook.init()

	payload, err := hook.Content.ExecuteString(data)
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

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(httpMethodEnum.StringDefault(int(hook.Method), http.MethodPost)), hook.Url, buf)
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

	if err := hook.ExtractAndValidateData(buf.Bytes(), data, logEntry); err != nil {
		return errors.Wrapf(err, "hook<%s> failed", hook.Url)
	}

	return nil
}

// ExtractAndValidateData
func (hook *Hook) ExtractAndValidateData(source []byte, data *hookData, logEntry *logger.LogEntry) error {
	for _, extractor := range hook.Extractors {
		value, err := extractor.Path.LookupNoQuotes(source)
		if err != nil {
			return reportValidation(logEntry, extractor, err)
		}
		strValue := string(value)
		if err := extractor.Validator.ValidateValue(strValue); err != nil {
			return reportValidation(logEntry, extractor, err)
		}
		data.Vars[extractor.Name] = strValue
	}
	return nil
}

func reportValidation(logEntry *logger.LogEntry, extractor Extractor, err error) error {
	switch extractor.Level {
	case FailLevelError:
		return errors.WithStack(err)
	case FailLevelWarning:
		logEntry.Logf(logger.WarningLevel, "hook %s extractor validation failed: %v", extractor.Name, err)
		return nil
	case FailLevelInfo:
		logEntry.LogInfo("HookValidation", fmt.Sprintf("hook %s extractor validation failed: %v", extractor.Name, err))
		return nil
	case FailLevelNone:
		return nil
	}
	return errors.Errorf("unknown faillevel<%s>", failLevelEnum.StringDefault(int(extractor.Level), fmt.Sprintf("%v", extractor.Level)))
}

// OkResponse checks if response is listed response code
func (hook *Hook) OkResponse(respCode int) bool {
	for _, code := range hook.RespCodes {
		if code == respCode {
			return true
		}
	}
	return false
}

//  Sent implements minimal traffic logger
func (tl *miniTrafficLogger) Sent(message []byte) {
	tl.LogEntry.LogDetail(logger.TrafficLevel, string(message), "Sent")
}

// Received implements minimal traffic logger
func (tl *miniTrafficLogger) Received(message []byte) {
	tl.LogEntry.LogDetail(logger.TrafficLevel, string(message), "Received")
}

// Validate validation rule
func (val *Validator) ValidateValue(value string) error {
	if val == nil {
		return nil
	}
	switch val.Type {
	case ValidationTypeNumber:
		floatValA, ok := val.Value.(float64)
		if !ok {
			return errors.Errorf("value<%v> type<%T> not a float64", val.Value, val.Value)
		}
		floatValB, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return errors.Errorf("extracted value<%s> not a number", value)
		}
		if helpers.NearlyEqual(floatValA, floatValB) {
			return nil
		}
		return errors.Errorf("value<%v> and extracted value<%s> not equal", val.Value, value)
	case ValidationTypeString:
		str, ok := val.Value.(string)
		if !ok {
			return errors.Errorf("value<%v> type<%T> not a string", val.Value, val.Value)
		}
		if str == value {
			return nil
		}
		return errors.Errorf("value<%s> and extracted value<%s> not equal", str, value)
	case ValidationTypeBool:
		boolValA, ok := val.Value.(bool)
		if !ok {
			return errors.Errorf("value<%v> type<%T> not a bool", val.Value, val.Value)
		}
		boolValB, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Errorf("extracted value<%v> not a bool", value)
		}
		if boolValA == boolValB {
			return nil
		}
		return errors.Errorf("value<%v> and extracted value<%v> not equal", val.Value, value)
	}
	return nil
}
