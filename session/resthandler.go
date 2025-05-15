package session

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/buildmetrics"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/globals"
	"github.com/qlik-oss/gopherciser/globals/constant"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/pending"
	"github.com/qlik-oss/gopherciser/runid"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/version"
	"github.com/rs/dnscache"
)

type (
	// RestMethod method with which to execute request
	RestMethod int

	// RestHandler handles waiting for pending requests and responses
	RestHandler struct {
		timeout         time.Duration
		Client          *http.Client
		trafficLogger   enigma.TrafficLogger
		headers         *HeaderJar
		virtualProxy    string
		ctx             context.Context
		pending         *pending.Handler
		defaultHost     string
		defaultProtocol string
	}

	// RestRequest represents a REST request and its response
	RestRequest struct {
		Method             RestMethod
		ContentType        string
		Content            []byte
		ContentReader      io.Reader
		Destination        string
		response           *http.Response
		ResponseBody       []byte
		ResponseStatus     string
		ResponseStatusCode int
		ResponseHeaders    http.Header
		ExtraHeaders       map[string]string
		NoVirtualProxy     bool
	}

	// ConnectionSettings interface
	ConnectionSettings interface {
		AllowUntrusted() bool
	}

	// Transport http transport interceptor
	Transport struct {
		*http.Transport
		*State
	}

	// ReqOptions options controlling handling of requests
	ReqOptions struct {
		// ExpectedStatusCode of response, empty list accepts everything (used e.g. for separate checking status)
		ExpectedStatusCode []int
		// FailOnError set to true for request to add an error to actionState, otherwise a warning is logged.
		FailOnError bool
		// ContentType defaults to application/json
		ContentType string
		// NoVirtualProxy disables the automatic adding of virtualproxy to request when a virtualproxy is defined.
		// This is useful e.g. when sending requests towards non sense environments as part of custom actions.
		NoVirtualProxy bool
	}

	MinimalTrafficLogger interface {
		Sent(message []byte)
		Received(message []byte)
	}

	CustomHeadersInput struct {
		// RunID is a string unique to an execution of gopherciser
		RunID              string
		GopherciserVersion string
		Session            logger.SessionEntry
		Action             logger.ActionEntry
	}

	CustomHeadersFunc func(input CustomHeadersInput, setHeader func(key, value string))
)

// RestMethod values
const (
	// GET RestMethod
	GET RestMethod = iota
	// POST RestMethod
	POST
	// DELETE RestMethod
	DELETE
	// PUT RestMethod
	PUT
	// PATCH RestMethod
	PATCH
	// HEAD RestMethod
	HEAD
	// OPTIONS RestMethod
	OPTIONS
)

var (
	registeredAddCustomHeaders    CustomHeadersFunc = func(CustomHeadersInput, func(key, value string)) {}
	registerCustomHeadersFuncOnce sync.Once
	streamContentTypes            = map[string]struct{}{
		"text/event-stream":       {},
		"application/stream+json": {},
		"application/x-ndjson":    {},
	}
)

var (
	restMethodEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
		"get":     int(GET),
		"post":    int(POST),
		"delete":  int(DELETE),
		"put":     int(PUT),
		"patch":   int(PATCH),
		"head":    int(HEAD),
		"options": int(OPTIONS),
	})

	defaultReqOptions = ReqOptions{
		ExpectedStatusCode: []int{http.StatusOK},
		FailOnError:        true,
		ContentType:        "application/json",
	}

	dnsResolver = &dnscache.Resolver{}
)

// RegisterCustomHeadersFunc adds extra headers to all http requests.
// RegisterCustomHeadersFunc shall be called only once and this call shall be before
// the gopherciser scenario runs.
func RegisterCustomHeadersFunc(f CustomHeadersFunc) error {
	if f == nil {
		return errors.New("can not register nil middleware")
	}
	registered := false
	registerCustomHeadersFuncOnce.Do(func() {
		registeredAddCustomHeaders = f
		registered = true
	})
	if !registered {
		return errors.New("RegisterCustomHeadersFunc shall not be called more than once")
	}
	return nil
}

func addCustomHeaders(req *http.Request, logEntry *logger.LogEntry) {
	if logEntry == nil {
		return
	}
	registeredAddCustomHeaders(
		CustomHeadersInput{
			RunID:              runid.Get(),
			GopherciserVersion: version.Version,
			Session: func() logger.SessionEntry {
				if logEntry == nil || logEntry.Session == nil {
					return logger.SessionEntry{}
				}
				return *logEntry.Session
			}(),
			Action: func() logger.ActionEntry {
				if logEntry == nil || logEntry.Action == nil {
					return logger.ActionEntry{}
				}
				return *logEntry.Action
			}(),
		},
		req.Header.Set,
	)
}

// NewRestHandler new instance of RestHandler
func NewRestHandler(ctx context.Context, trafficLogger enigma.TrafficLogger, headerjar *HeaderJar, virtualProxy string, timeout time.Duration, pendingHandler *pending.Handler) *RestHandler {
	return &RestHandler{
		trafficLogger: trafficLogger,
		headers:       headerjar,
		virtualProxy:  virtualProxy,
		timeout:       timeout,
		ctx:           ctx,
		pending:       pendingHandler,
	}
}

// UnmarshalJSON unmarshal RestMethod
func (method *RestMethod) UnmarshalJSON(arg []byte) error {
	i, err := restMethodEnumMap.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal RestMethod")
	}

	*method = RestMethod(i)
	return nil
}

// MarshalJSON marshal RestMethod
func (method RestMethod) MarshalJSON() ([]byte, error) {
	str, err := restMethodEnumMap.String(int(method))
	if err != nil {
		return nil, errors.Errorf("Unknown RestMethod<%d>", method)
	}
	return []byte(fmt.Sprintf(`"%s"`, strings.ToUpper(str))), nil
}

// String implements fmt.Stringer interface
func (method RestMethod) String() string {
	str, err := restMethodEnumMap.String(int(method))
	if err != nil {
		return "unknown"
	}
	return str
}

// DefaultClient creates client instance with default client settings
func DefaultClient(allowUntrusted bool, state *State) (*http.Client, error) {
	// todo client values are currently from http.DefaultTransport, should choose better values depending on
	// configured timeout etc

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return nil },
		Transport: &Transport{
			&http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
					host, port, err := net.SplitHostPort(addr)
					if err != nil {
						return nil, err
					}
					ips, err := dnsResolver.LookupHost(ctx, host)
					if err != nil {
						return nil, err
					}
					for _, ip := range ips {
						dialer := &net.Dialer{
							Timeout:   state.Timeout,
							KeepAlive: 30 * time.Second,
						}
						var conn net.Conn
						conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
						if err == nil {
							return conn, nil
						}
					}
					return nil, err
				},
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: allowUntrusted,
				},
			},
			state,
		},
	}

	if state.Cookies != nil {
		client.Jar = state.Cookies
	} else {
		var err error
		client.Jar, err = cookiejar.New(nil)
		if err != nil {
			return client, errors.Wrap(err, "failed creating cookie jar")
		}
		state.Cookies = client.Jar
	}

	return client, nil
}

// DefaultReqOptions sets expected status code to 200 and fails on error
func DefaultReqOptions() *ReqOptions {
	options := ReqOptions{
		ExpectedStatusCode: make([]int, len(defaultReqOptions.ExpectedStatusCode)),
		FailOnError:        defaultReqOptions.FailOnError,
		ContentType:        defaultReqOptions.ContentType,
	}
	copy(options.ExpectedStatusCode, defaultReqOptions.ExpectedStatusCode)
	return &options
}

// SetClient set HTTP client for this RestHandler
func (handler *RestHandler) SetClient(client *http.Client, defaultHost, defaultProtocol string) {
	handler.Client = client
	handler.defaultHost = defaultHost
	handler.defaultProtocol = defaultProtocol
}

func (handler *RestHandler) Host() string {
	return handler.defaultHost
}

func (handler *RestHandler) Protocol() string {
	return handler.defaultProtocol
}

// GetSync sends synchronous GET request with options, using options=nil default options are used
func (handler *RestHandler) GetSync(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions) (*RestRequest, error) {
	return handler.GetSyncWithCallback(url, actionState, logEntry, options, nil)
}

// GetSyncOnce same as GetSync but only called once in the same session
func (handler *RestHandler) GetSyncOnce(url string, actionState *action.State, sessionState *State, logEntry *logger.LogEntry, options *ReqOptions, uniqueString string) (*RestRequest, error) {
	var req *RestRequest
	var err error
	sessionState.Once(fmt.Sprintf("%sGET%s", uniqueString, url), func() {
		req, err = handler.GetSyncWithCallback(url, actionState, logEntry, options, nil)
	})
	return req, err
}

// GetSyncWithCallback sends synchronous GET request with options and callback, using options=nil default options are used
func (handler *RestHandler) GetSyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions, callback func(err error, req *RestRequest)) (*RestRequest, error) {
	var reqErr error
	var request *RestRequest
	var wg sync.WaitGroup // todo rewrite send request to be able to not run in goroutine.
	wg.Add(1)
	syncCallback := func(err error, req *RestRequest) {
		defer wg.Done()
		reqErr = err
		request = req
		if callback != nil {
			callback(err, req)
		}
	}
	handler.getAsyncWithCallback(url, actionState, logEntry, nil, options, syncCallback)
	wg.Wait()
	return request, reqErr
}

// GetAsync send async GET request with options, using options=nil default options are used
func (handler *RestHandler) GetAsync(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions) *RestRequest {
	return handler.getAsyncWithCallback(url, actionState, logEntry, nil, options, nil)
}

// GetAsyncOnce same as GetAsync but only called once in the same session
func (handler *RestHandler) GetAsyncOnce(url string, actionState *action.State, sessionState *State, logEntry *logger.LogEntry, options *ReqOptions, uniqueString string) *RestRequest {
	var req *RestRequest
	sessionState.Once(fmt.Sprintf("%sGET%s", uniqueString, url), func() {
		req = handler.getAsyncWithCallback(url, actionState, logEntry, nil, options, nil)
	})
	return req
}

// GetWithHeadersAsync send async GET request with headers and options, using options=nil default options are used
func (handler *RestHandler) GetWithHeadersAsync(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.getAsyncWithCallback(url, actionState, logEntry, headers, options, callback)
}

// GetAsyncWithCallback send async GET request with options and callback, with options=nil default options are used
func (handler *RestHandler) GetAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.getAsyncWithCallback(url, actionState, logEntry, nil, options, callback)
}

func (handler *RestHandler) getAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.sendAsyncWithCallback(GET, url, actionState, logEntry, nil, headers, options, callback)
}

// HeadSync sends synchronous HEAD request with options, using options=nil default options are used
func (handler *RestHandler) HeadSync(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions) (*RestRequest, error) {
	return handler.HeadSyncWithCallback(url, actionState, logEntry, options, nil)
}

// HeadSyncWithCallback sends synchronous HEAD request with options and callback, using options=nil default options are used
func (handler *RestHandler) HeadSyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions, callback func(err error, req *RestRequest)) (*RestRequest, error) {
	var reqErr error
	var request *RestRequest
	var wg sync.WaitGroup
	wg.Add(1)
	syncCallback := func(err error, req *RestRequest) {
		defer wg.Done()
		reqErr = err
		request = req
		if callback != nil {
			callback(err, req)
		}
	}
	handler.headAsyncWithCallback(url, actionState, logEntry, nil, options, syncCallback)
	wg.Wait()
	return request, reqErr
}

// HeadAsync send async HEAD request with options, using options=nil default options are used
func (handler *RestHandler) HeadAsync(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions) *RestRequest {
	return handler.headAsyncWithCallback(url, actionState, logEntry, nil, options, nil)
}

// HeadWithHeadersAsync send async HEAD request with headers and options, using options=nil default options are used
func (handler *RestHandler) HeadWithHeadersAsync(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.headAsyncWithCallback(url, actionState, logEntry, headers, options, callback)
}

// HeadAsyncWithCallback send async HEAD request with options and callback, with options=nil default options are used
func (handler *RestHandler) HeadAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.headAsyncWithCallback(url, actionState, logEntry, nil, options, callback)
}

func (handler *RestHandler) headAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.sendAsyncWithCallback(HEAD, url, actionState, logEntry, nil, headers, options, callback)
}

// PutAsync send async PUT request with options, using options=nil default options are used
func (handler *RestHandler) PutAsync(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, options *ReqOptions) *RestRequest {
	return handler.PutAsyncWithCallback(url, actionState, logEntry, content, nil, options, nil)
}

// PutWithHeadersAsync send async PUT request with options and headers, using options=nil default options are used
func (handler *RestHandler) PutWithHeadersAsync(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions) *RestRequest {
	return handler.PutAsyncWithCallback(url, actionState, logEntry, content, headers, options, nil)
}

// PutAsyncWithCallback send async PUT request with options and callback, using options=nil default options are used
func (handler *RestHandler) PutAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.sendAsyncWithCallback(PUT, url, actionState, logEntry, content, headers, options, callback)
}

// PatchAsync send async PATCH request with options, using options=nil default options are used
func (handler *RestHandler) PatchAsync(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, options *ReqOptions) *RestRequest {
	return handler.PatchAsyncWithCallback(url, actionState, logEntry, content, nil, options, nil)
}

// PatchWithHeadersAsync send async PATCH request with options and headers, using options=nil default options are used
func (handler *RestHandler) PatchWithHeadersAsync(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions) *RestRequest {
	return handler.PatchAsyncWithCallback(url, actionState, logEntry, content, headers, options, nil)
}

// PatchAsyncWithCallback send async PATCH request with options and callback, using options=nil default options are used
func (handler *RestHandler) PatchAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.sendAsyncWithCallback(PATCH, url, actionState, logEntry, content, headers, options, callback)
}

// PostAsync send async POST request with options, using options=nil default options are used
func (handler *RestHandler) PostAsync(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, options *ReqOptions) *RestRequest {
	return handler.PostAsyncWithCallback(url, actionState, logEntry, content, nil, options, nil)
}

// PostSync send sync POST request with options, using options=nil default options are used
func (handler *RestHandler) PostSync(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, options *ReqOptions) (*RestRequest, error) {
	return handler.PostSyncWithCallback(url, actionState, logEntry, content, nil, options, nil)
}

// PostWithHeadersAsync send async POST request with options and headers, using options=nil default options are used
func (handler *RestHandler) PostWithHeadersAsync(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions) *RestRequest {
	return handler.PostAsyncWithCallback(url, actionState, logEntry, content, headers, options, nil)
}

// PostAsyncWithCallback send async POST request with options and callback, using options=nil default options are used
func (handler *RestHandler) PostAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.sendAsyncWithCallback(POST, url, actionState, logEntry, content, headers, options, callback)
}

// PostSyncWithCallback send sync POST request with options and callback, using options=nil default options are used
func (handler *RestHandler) PostSyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) (*RestRequest, error) {
	return handler.sendSyncWithCallback(POST, url, actionState, logEntry, content, headers, options, callback)
}

// DeleteAsyncWithCallback send async DELETE request with options and callback, using options=nil default options are used
func (handler *RestHandler) DeleteAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.sendAsyncWithCallback(DELETE, url, actionState, logEntry, nil, headers, options, callback)
}

// DeleteSyncWithCallback send sync DELETE request with options and callback, using options=nil default options are used
func (handler *RestHandler) DeleteSyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) (*RestRequest, error) {
	return handler.sendSyncWithCallback(DELETE, url, actionState, logEntry, nil, headers, options, callback)
}

// DeleteAsyncWithHeaders send async DELETE request with options and headers, using options=nil default options are used
func (handler *RestHandler) DeleteAsyncWithHeaders(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions) *RestRequest {
	return handler.DeleteAsyncWithCallback(url, actionState, logEntry, headers, options, nil)
}

// DeleteAsync send async DELETE request with options, using options=nil default options are used
func (handler *RestHandler) DeleteAsync(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions) *RestRequest {
	return handler.DeleteAsyncWithCallback(url, actionState, logEntry, nil, options, nil)
}

// DeleteSync send sync DELETE request with options, using options=nil default options are used
func (handler *RestHandler) DeleteSync(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions) (*RestRequest, error) {
	return handler.DeleteSyncWithCallback(url, actionState, logEntry, nil, options, nil)
}

// OptionsAsync send async request with options, using options=nil default options are used
func (handler *RestHandler) OptionsAsync(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions) *RestRequest {
	return handler.sendAsyncWithCallback(OPTIONS, url, actionState, logEntry, nil, headers, options, nil)
}

// OptionsAsyncWithCallback send async request with options and callback, using options=nil default options are used
func (handler *RestHandler) OptionsAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.sendAsyncWithCallback(OPTIONS, url, actionState, logEntry, nil, headers, options, callback)
}

// OptionsSync send sync request with options, using options=nil default options are used
func (handler *RestHandler) OptionsSync(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions) (*RestRequest, error) {
	return handler.sendSyncWithCallback(OPTIONS, url, actionState, logEntry, nil, headers, options, nil)
}

// OptionsSyncWithCallback send sync request with options and callback, using options=nil default options are used
func (handler *RestHandler) OptionsSyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) (*RestRequest, error) {
	return handler.sendSyncWithCallback(OPTIONS, url, actionState, logEntry, nil, headers, options, callback)
}

// PatchSync send sync PATCH request with options, using options=nil default options are used
func (handler *RestHandler) PatchSync(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, options *ReqOptions) (*RestRequest, error) {
	return handler.PatchSyncWithCallback(url, actionState, logEntry, content, nil, options, nil)
}

// PatchSyncWithCallback send sync PATCH request with options and callback, using options=nil default options are used
func (handler *RestHandler) PatchSyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) (*RestRequest, error) {
	return handler.sendSyncWithCallback(PATCH, url, actionState, logEntry, content, headers, options, callback)
}

func (handler *RestHandler) sendSyncWithCallback(method RestMethod, url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) (*RestRequest, error) {
	var returnErr error
	var wg sync.WaitGroup
	wg.Add(1)
	returnReq := handler.sendAsyncWithCallback(method, url, actionState, logEntry, content, headers, options, func(err error, req *RestRequest) {
		defer wg.Done()
		returnErr = err
		if callback != nil {
			callback(err, req)
		}
	})
	wg.Wait()
	return returnReq, returnErr
}

func (handler *RestHandler) sendAsyncWithCallback(method RestMethod, url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	if options == nil {
		options = &defaultReqOptions
	}

	sendRequest := RestRequest{
		Method:         method,
		ContentType:    options.ContentType,
		Content:        content,
		Destination:    url,
		NoVirtualProxy: options.NoVirtualProxy,
		ExtraHeaders:   headers,
	}

	handler.QueueRequestWithCallback(actionState, options.FailOnError, &sendRequest, logEntry, createStatusCallback(actionState, logEntry, &sendRequest, options, callback))

	return &sendRequest
}

func createStatusCallback(actionState *action.State, logEntry *logger.LogEntry, request *RestRequest, options *ReqOptions, callback func(err error, req *RestRequest)) func(err error, req *RestRequest) {
	return func(err error, req *RestRequest) {
		// check response status code
		if err == nil && len(options.ExpectedStatusCode) > 0 {
			if err = CheckResponseStatus(request, options.ExpectedStatusCode); err != nil {
				WarnOrError(actionState, logEntry, options.FailOnError, errors.Wrapf(err, "Unexpected status code: %s (%s %s)", req.ResponseStatus, req.Method, request.Destination))
			}
		}
		if callback != nil {
			callback(err, req)
		}
	}
}

// CheckResponseStatus validates that a response has acceptable
func CheckResponseStatus(request *RestRequest, statusCodes []int) error {
	if request.ResponseBody == nil {
		return errors.New("got empty response")
	}
	if request.ResponseStatus == "" {
		return errors.New("did not get a response status")
	}

	for _, code := range statusCodes {
		if code == request.ResponseStatusCode {
			return nil
		}
	}
	return errors.Errorf("unexpected response status code<%d> expected<%v>", request.ResponseStatusCode, statusCodes)
}

func getHost(fullURL string) (string, error) {
	urlObj, err := url.Parse(fullURL)
	if err != nil {
		return "", err
	}
	host := strings.Split(urlObj.Host+urlObj.Path, "/")[0]
	if host == "" {
		return "", errors.Errorf("Failed to extract hostname from <%v>", fullURL)
	}

	return strings.Split(host, ":")[0], nil
}

// QueueRequest Async request
//
// Deprecated: Use method specific function instead
func (handler *RestHandler) QueueRequest(actionState *action.State, failOnError bool,
	request *RestRequest, logEntry *logger.LogEntry) {
	handler.QueueRequestWithCallback(actionState, failOnError, request, logEntry, nil)
}

// QueueRequestWithCallback Async request with callback, set warnOnError to log warning instead of registering error for request
func (handler *RestHandler) QueueRequestWithCallback(actionState *action.State, failOnError bool,
	request *RestRequest, logEntry *logger.LogEntry, callback func(err error, req *RestRequest)) {
	handler.pending.IncPending()

	startTS := time.Now()
	go func() {
		var errRequest error
		stall := time.Since(startTS)
		failRequest := func(err error) {
			errRequest = err
			actionState.AddErrors(err)
		}

		var panicErr error
		defer func() {
			defer handler.pending.DecPending()
			if callback != nil {
				if panicErr != nil { // propagate panic error to callback
					errRequest = panicErr
				}
				// recover from and error report any panic inside callback
				if err := helpers.RecoverWithErrorFunc(func() {
					callback(errRequest, request)
				}); err != nil {
					actionState.AddErrors(err)
				}
				return
			}
			// no callback handling errors, report panic error directly
			if panicErr != nil {
				actionState.AddErrors(panicErr)
			}
		}()

		panicErr = helpers.RecoverWithErrorFunc(func() {

			if stall > constant.MaxStallTime {
				logEntry.LogDetail(logger.WarningLevel, "Goroutine stall", strconv.FormatInt(stall.Nanoseconds(), 10))
			}

			if handler.Client == nil {
				failRequest(errors.New("no REST client initialized"))
				return
			}

			host, err := getHost(request.Destination)
			if err != nil {
				failRequest(errors.Wrapf(err, `Failed to extract host from "%s"`, request.Destination))
				return
			}

			if err := handler.addVirtualProxy(request); err != nil {
				failRequest(errors.WithStack(err))
				return
			}

			req, err := newStdRequest(handler.ctx, request, logEntry, handler.headers.GetHeader(host))
			if err != nil {
				failRequest(errors.WithStack(err))
				return
			}
			doTs := time.Now()
			request.response, errRequest = handler.Client.Do(req)
			if errRequest != nil {
				WarnOrError(actionState, logEntry, failOnError, errors.Wrap(errRequest, "HTTP request fail"))
			}
			if request.response != nil {
				defer func() {
					if err := request.response.Body.Close(); err != nil {
						WarnOrError(actionState, logEntry, failOnError, errors.Wrap(err, "failed to close request body"))
					}
				}()
				request.ResponseStatus = request.response.Status
				request.ResponseStatusCode = request.response.StatusCode
				request.ResponseHeaders = request.response.Header
				request.ResponseBody, errRequest = io.ReadAll(request.response.Body)
				contentType := request.response.Header.Get("Content-Type")
				mediaType := ""
				if contentType != "" {
					mediaType, _, err = mime.ParseMediaType(contentType)
					if err != nil {
						logEntry.Logf(logger.WarningLevel, "failed to parse content type %s", request.response.Header.Get("Content-Type"))
					}
				}

				// When content type is a stream normal metric log will be time to response without starting to stream the body. Thus this will log response time to stream end
				if _, ok := streamContentTypes[mediaType]; ok && logEntry.ShouldLogTrafficMetrics() {
					logEntry.LogTrafficMetric(time.Since(doTs).Nanoseconds(), 0, uint64(len(request.ResponseBody)), -1, req.URL.Path, "", "STREAM", "")
				}
			}
		})
	}()
}

func (handler *RestHandler) addVirtualProxy(request *RestRequest) error {
	if handler.virtualProxy != "" && !request.NoVirtualProxy {
		destination, err := prependURLPath(request.Destination, handler.virtualProxy)
		if err != nil {
			return errors.Wrapf(err, "failed to prepend virtual proxy<%s> to url<%s>", destination, handler.virtualProxy)
		}
		if destination == "" {
			return errors.Errorf("appending virtualproxy<%s> to destination<%s> failed", handler.virtualProxy, request.Destination)
		}
		request.Destination = destination
	}
	return nil
}

func prependURLPath(aURL, pathToPrepend string) (string, error) {
	urlObj, err := url.Parse(aURL)
	if err != nil {
		return aURL, errors.WithStack(err)
	}
	urlObj.Path = path.Join(pathToPrepend, urlObj.Path)

	return urlObj.String(), nil
}

func newStdRequest(ctx context.Context, request *RestRequest, logEntry *logger.LogEntry, mainHeader http.Header) (*http.Request, error) {
	var req *http.Request
	var err error

	switch request.Method {
	case HEAD:
		req, err = http.NewRequest(http.MethodHead, request.Destination, nil)
	case GET:
		req, err = http.NewRequest(http.MethodGet, request.Destination, nil)
	case DELETE:
		req, err = http.NewRequest(http.MethodDelete, request.Destination, nil)
	case POST:
		req, err = http.NewRequest(http.MethodPost, request.Destination, getRequestReader(request))
	case PUT:
		req, err = http.NewRequest(http.MethodPut, request.Destination, getRequestReader(request))
	case PATCH:
		req, err = http.NewRequest(http.MethodPatch, request.Destination, getRequestReader(request))
	case OPTIONS:
		req, err = http.NewRequest(http.MethodOptions, request.Destination, nil)
	default:
		return nil, errors.Errorf("Unsupported REST method<%v>", request.Method)
	}
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create HTTP request")
	}
	addCustomHeaders(req, logEntry)
	//Set user-agent as special "gopherciser version". version is set from the version package during build.
	req.Header.Set("User-Agent", globals.UserAgent())
	for k, v := range mainHeader {
		req.Header[k] = v
	}
	req.Header.Set("Content-Type", request.ContentType)
	for k, v := range request.ExtraHeaders {
		req.Header.Set(k, v)
	}
	return req.WithContext(ctx), nil
}

func getRequestReader(request *RestRequest) io.Reader {
	if request.ContentReader != nil {
		return request.ContentReader
	}
	return bytes.NewReader(request.Content)
}

// RoundTrip implement RoundTripper interface
func (transport *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	sentTS := time.Now()

	// log errors on current action if we have one
	logErrors := func(err error) {
		// Set error on action state if we have one.
		if transport.State != nil && transport.CurrentActionState != nil {
			transport.CurrentActionState.AddErrors(err)
			return
		}
		if transport.LogEntry != nil {
			transport.LogEntry.LogError(err)
		}
	}

	isApp := false
	// bug req.ContentLength seems to be 0 for our POST messages with upload app, instead try to check if we're posting an app
	// this also means sent size logged is incorrect
	if (req.Method == http.MethodPost || req.Method == http.MethodPut) && req.Header["Content-Type"] != nil {
		isApp = contentIsBinary(req.Header)
	}

	requestID := transport.Counters.RestRequestID.Inc()

	reqSize := int64(0)
	if req.ContentLength > 0 {
		reqSize = req.ContentLength
	}

	if err := transport.RequestMetrics.UpdateSent(sentTS, reqSize); err != nil {
		logErrors(errors.Wrap(err, "failed to update sent request metrics"))
	}

	body := !isApp && reqSize <= constant.MaxBodySize // avoid logging large bodies
	LogTrafficOut(req, body, transport.trafficLogger, transport.LogEntry, requestID)

	resp, err := transport.Transport.RoundTrip(req)
	if err != nil || resp == nil {
		logErrors(errors.Wrapf(err, "failed to perform HTTP request<%s>", req.URL))
		return resp, err
	}

	recTS := time.Now()

	apiPath := apiCallFromPath(req.URL.Path)
	if apiPath != "" {
		actionString := "unknown"
		labelString := ""
		if transport.LogEntry.Action != nil {
			actionString = transport.LogEntry.Action.Action
			labelString = transport.LogEntry.Action.Label
		}
		buildmetrics.ReportApiResult(actionString, labelString,
			apiPath, req.Method, resp.StatusCode, recTS.Sub(sentTS))
	}

	respSize := int64(0)
	if resp.ContentLength > 0 {
		respSize = resp.ContentLength
	}

	if err := transport.RequestMetrics.UpdateReceived(recTS, respSize); err != nil {
		logErrors(errors.Wrap(err, "failed to update received request metrics"))
	}

	// get request statics collector
	var requestStats *statistics.RequestStats
	if req.URL != nil && req.URL.Path != "" {
		requestStats = transport.Counters.StatisticsCollector.GetOrAddRequestStats(req.Method, req.URL.Path)
	}

	if transport.LogEntry.ShouldLogTrafficMetrics() || requestStats != nil {
		respTime := recTS.Sub(sentTS).Nanoseconds()
		sent := uint64(reqSize)
		received := uint64(respSize)

		if requestStats != nil {
			requestStats.RespAvg.AddSample(uint64(respTime))
			requestStats.Sent.Add(sent)
			requestStats.Received.Add(received)
		}

		if transport.LogEntry.ShouldLogTrafficMetrics() {
			// todo somehow id sent/rec requests?
			buf := helpers.GlobalBufferPool.Get()
			defer helpers.GlobalBufferPool.Put(buf)
			buf.WriteString(req.Method)
			var query string
			if req.URL != nil {
				if _, err := buf.WriteString(" "); err != nil {
					transport.LogEntry.Logf(logger.WarningLevel, "failed writing to buffer: %v", err)
				}
				if _, err := buf.WriteString(req.URL.Path); err != nil {
					transport.LogEntry.Logf(logger.WarningLevel, "failed writing to buffer: %v", err)
				}
				query = req.URL.RawQuery
			}

			// Add trace ID to metric message if exist as header
			traceID := ""
			if resp.Header != nil {
				traceID = resp.Header.Get("x-b3-traceid")
			}
			msg := ""
			if traceID != "" {
				msg = fmt.Sprintf("traceID:%s", traceID)
			}

			transport.LogEntry.LogTrafficMetric(recTS.Sub(sentTS).Nanoseconds(), sent, received, -1, buf.String(), query, "REST", msg)
		}
	}

	LogTrafficIn(resp, transport.trafficLogger, transport.LogEntry, requestID)

	return resp, err
}

func LogTrafficIn(resp *http.Response, trafficLogger MinimalTrafficLogger, logEntry *logger.LogEntry, requestID uint64) {
	if logEntry.ShouldLogTraffic() && trafficLogger != nil {
		body := !contentIsBinary(resp.Header)
		respSize := int64(0)
		if resp.ContentLength > 0 {
			respSize = resp.ContentLength
		}

		if respSize > constant.MaxBodySize {
			body = false // avoid logging large bodies
		}
		if trafficIn, err := httputil.DumpResponse(resp, body); err == nil {
			trafficLogger.Received(append([]byte(fmt.Sprintf("[%d] ", requestID)), trafficIn...))
		} else {
			logEntry.Log(logger.WarningLevel, "error dumping response", err)
		}
	}
}

func LogTrafficOut(req *http.Request, doLogBody bool, trafficLogger MinimalTrafficLogger, logEntry *logger.LogEntry, requestID uint64) {
	if logEntry.ShouldLogTraffic() && trafficLogger != nil {
		if trafficOut, err := httputil.DumpRequestOut(req, doLogBody); err == nil {
			trafficLogger.Sent(append([]byte(fmt.Sprintf("[%d] ", requestID)), trafficOut...))
		} else {
			logEntry.Log(logger.WarningLevel, "error dumping request", err)
		}
	} else {
		// make sure global and local request counters ticks.
		trafficLogger.Sent(nil)
	}
}

func contentIsBinary(header http.Header) bool {
	if header["Content-Type"] == nil {
		return false
	}
	for _, contentType := range header["Content-Type"] {
		if strings.ToLower(contentType) == "application/octet-stream" || strings.HasSuffix(contentType, ".app") {
			return true
		}
	}
	return false
}

const apiSeparator = "api/v1/"

func apiCallFromPath(path string) string {
	splitApiV1 := strings.SplitN(path, apiSeparator, 2)
	if len(splitApiV1) < 2 {
		return "" // No api call found in path
	}
	apiCall := splitApiV1[1]
	splitSlash := strings.SplitN(apiCall, "/", 2)
	if len(splitSlash) < 1 {
		return "" // Nothing after apiSeparator (which is weird)
	}
	return fmt.Sprintf("%s%s", apiSeparator, splitSlash[0])
}
