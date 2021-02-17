package session

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qlik-oss/gopherciser/buildmetrics"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/globals/constant"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/version"
)

type (
	// RestMethod method with which to execute request
	RestMethod int

	// RestHandler handles waiting for pending requests and responses
	RestHandler struct {
		reqCounterCond *sync.Cond
		reqCounter     int
		timeout        time.Duration
		Client         *http.Client
		trafficLogger  enigma.TrafficLogger
		headers        *HeaderJar
		virtualProxy   string
		ctx            context.Context
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
)

var (
	restMethodEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
		"get":    int(GET),
		"post":   int(POST),
		"delete": int(DELETE),
		"put":    int(PUT),
	})

	defaultReqOptions = ReqOptions{
		ExpectedStatusCode: []int{http.StatusOK},
		FailOnError:        true,
		ContentType:        "application/json",
	}
)

// NewRestHandler new instance of RestHandler
func NewRestHandler(ctx context.Context, size int, trafficLogger enigma.TrafficLogger, headerjar *HeaderJar, virtualProxy string, timeout time.Duration) *RestHandler {
	return &RestHandler{
		reqCounter:     0,
		reqCounterCond: sync.NewCond(&sync.Mutex{}),
		trafficLogger:  trafficLogger,
		headers:        headerjar,
		virtualProxy:   virtualProxy,
		timeout:        timeout,
		ctx:            ctx,
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

// WaitForPending uses double locking of mutex to wait until mutex is unlocked by
// loop listening for pending req/resp
func (handler *RestHandler) WaitForPending() {
	handler.reqCounterCond.L.Lock()
	for handler.reqCounter > 0 {
		handler.reqCounterCond.Wait()
	}
	handler.reqCounterCond.L.Unlock()
}

// IncPending increase pending requests
func (handler *RestHandler) IncPending() {
	handler.reqCounterCond.L.Lock()
	handler.reqCounter++
	handler.reqCounterCond.Broadcast()
	handler.reqCounterCond.L.Unlock()
}

// DecPending increase finished requests
func (handler *RestHandler) DecPending(request *RestRequest) {
	handler.reqCounterCond.L.Lock()
	handler.reqCounter--
	handler.reqCounterCond.Broadcast()
	handler.reqCounterCond.L.Unlock()
}

// DefaultClient creates client instance with default client settings
func DefaultClient(allowUntrusted bool, state *State) (*http.Client, error) {
	// todo client values are currently from http.DefaultClient, should choose better values depending on
	// configured timeout etc

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return nil },
		Transport: &Transport{
			&http.Transport{
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
func (handler *RestHandler) SetClient(client *http.Client) {
	handler.Client = client
}

// FireOffGet request
// Deprecated: Replace with GetAsync or QueueRequest
func (handler *RestHandler) FireOffGet(url string, actionState *action.State, failOnError bool) *RestRequest {
	return handler.FireOffGetWithCallback(url, actionState, failOnError, nil)
}

// FireOffGet request with callback
// Deprecated: Replace with GetAsyncWithCallback or QueueRequestWithCallback
func (handler *RestHandler) FireOffGetWithCallback(url string, actionState *action.State, failOnError bool, callback func(err error, req *RestRequest)) *RestRequest {
	getRequest := RestRequest{
		Method:      GET,
		Destination: url,
	}
	handler.QueueRequestWithCallback(actionState, failOnError, &getRequest, &logger.LogEntry{}, callback)
	return &getRequest
}

// GetSync sends synchronous GET request with options, using options=nil default options are used
func (handler *RestHandler) GetSync(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions) (*RestRequest, error) {
	return handler.GetSyncWithCallback(url, actionState, logEntry, options, nil)
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

// GetWithHeadersAsync send async GET request with headers and options, using options=nil default options are used
func (handler *RestHandler) GetWithHeadersAsync(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.getAsyncWithCallback(url, actionState, logEntry, headers, options, callback)
}

// GetAsyncWithCallback send async GET request with options and callback, with options=nil default options are used
func (handler *RestHandler) GetAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	return handler.getAsyncWithCallback(url, actionState, logEntry, nil, options, callback)
}

func (handler *RestHandler) getAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	if options == nil {
		options = &defaultReqOptions
	}

	getRequest := RestRequest{
		Method:       GET,
		Destination:  url,
		ContentType:  options.ContentType,
		ExtraHeaders: headers,
	}

	handler.QueueRequestWithCallback(actionState, options.FailOnError, &getRequest, logEntry, createStatusCallback(actionState, logEntry, &getRequest, options, callback))

	return &getRequest
}

// PostAsync send async POST request with options, using options=nil default options are used
func (handler *RestHandler) PostAsync(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, options *ReqOptions) *RestRequest {
	return handler.PostAsyncWithCallback(url, actionState, logEntry, content, nil, options, nil)
}

// PostWithHeaderAsync send async POST request with options and headers, using options=nil default options are used
func (handler *RestHandler) PostWithHeadersAsync(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions) *RestRequest {
	return handler.PostAsyncWithCallback(url, actionState, logEntry, content, headers, options, nil)
}

// PostAsync send async POST request with options and callback, using options=nil default options are used
func (handler *RestHandler) PostAsyncWithCallback(url string, actionState *action.State, logEntry *logger.LogEntry, content []byte, headers map[string]string, options *ReqOptions, callback func(err error, req *RestRequest)) *RestRequest {
	if options == nil {
		options = &defaultReqOptions
	}

	postRequest := RestRequest{
		Method:         POST,
		ContentType:    options.ContentType,
		Content:        content,
		Destination:    url,
		NoVirtualProxy: options.NoVirtualProxy,
		ExtraHeaders:   headers,
	}

	handler.QueueRequestWithCallback(actionState, options.FailOnError, &postRequest, logEntry, createStatusCallback(actionState, logEntry, &postRequest, options, callback))

	return &postRequest
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

func getUrlObj(fullURL string) (*url.URL, error) {
	urlObj, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}
	return urlObj, nil
}

// QueueRequest Async request
func (handler *RestHandler) QueueRequest(actionState *action.State, failOnError bool,
	request *RestRequest, logEntry *logger.LogEntry) {
	handler.QueueRequestWithCallback(actionState, failOnError, request, logEntry, nil)
}

// QueueRequestWithCallback Async request with callback, set warnOnError to log warning instead of registering error for request
func (handler *RestHandler) QueueRequestWithCallback(actionState *action.State, failOnError bool,
	request *RestRequest, logEntry *logger.LogEntry, callback func(err error, req *RestRequest)) {
	handler.IncPending()

	startTS := time.Now()
	go func() {
		stall := time.Since(startTS)
		defer handler.DecPending(request)
		var errRequest error
		var panicErr error
		defer helpers.RecoverWithError(&panicErr)
		if callback != nil {
			defer func() {
				if panicErr != nil {
					errRequest = panicErr
				}
				callback(errRequest, request)
			}()
		}

		if stall > constant.MaxStallTime {
			logEntry.LogDetail(logger.WarningLevel, "Goroutine stall", strconv.FormatInt(stall.Nanoseconds(), 10))
		}

		if handler.Client == nil {
			errRequest = errors.New("no REST client initialized")
			actionState.AddErrors(errRequest)
			return
		}

		var host string
		host, errRequest = getHost(request.Destination)
		if errRequest != nil {
			WarnOrError(actionState, logEntry, failOnError, errors.Wrapf(errRequest, "Failed to read REST response to %s", request.Destination))
		}

		if request.ContentReader == nil {
			if errRequest = handler.performRestCall(handler.ctx, request, handler.Client, handler.headers.GetHeader(host)); errRequest != nil {
				WarnOrError(actionState, logEntry, failOnError, errors.WithStack(errRequest))
			}
		} else {
			if errRequest = handler.postWithReader(handler.ctx, request, handler.Client, logEntry, handler.headers.GetHeader(host)); errRequest != nil {
				WarnOrError(actionState, logEntry, failOnError, errors.WithStack(errRequest))
			}
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
			request.ResponseBody, errRequest = ioutil.ReadAll(request.response.Body)
		}
	}()
}

func ReadAll(r io.Reader) ([]byte, error) {
	buf := helpers.GlobalBufferPool.Get()
	defer helpers.GlobalBufferPool.Put(buf)

	capacity := int64(bytes.MinRead)
	var err error
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	if int64(int(capacity)) == capacity {
		buf.Grow(int(capacity))
	}
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}

func (handler *RestHandler) performRestCall(ctx context.Context, request *RestRequest, client *http.Client, headers http.Header) error {

	destination := request.Destination
	if handler.virtualProxy != "" && !request.NoVirtualProxy {
		host, err := getHost(request.Destination)
		if err != nil {
			return err
		}
		urlObj, err := getUrlObj(request.Destination)
		if err != nil {
			return err
		}
		destination = fmt.Sprintf("%s://%s/%s%s", urlObj.Scheme, host, handler.virtualProxy, urlObj.Path)
	}

	var req *http.Request
	var err error

	switch request.Method {
	case GET:
		req, err = http.NewRequest(http.MethodGet, destination, nil)
		if err != nil {
			return errors.Wrapf(err, "Failed to create HTTP request")
		}
	case DELETE:
		req, err = http.NewRequest(http.MethodDelete, destination, nil)
		if err != nil {
			return errors.Wrapf(err, "Failed to create HTTP request")
		}
	case POST:
		req, err = http.NewRequest(http.MethodPost, destination, bytes.NewReader(request.Content))
		if err != nil {
			return errors.Wrap(err, "Failed to create HTTP request")
		}
	case PUT:
		req, err = http.NewRequest(http.MethodPut, destination, bytes.NewReader(request.Content))
		if err != nil {
			return errors.Wrap(err, "Failed to create HTTP request")
		}
	default:
		return errors.Errorf("Unsupported REST method<%v>", request.Method)
	}
	req = req.WithContext(ctx)
	handler.newHeader(headers, request, req.Header)

	res, err := client.Do(req)
	request.response = res
	if err != nil {
		return errors.Wrap(err, "HTTP request fail")
	}
	return nil
}

func (handler *RestHandler) newHeader(mainHeader http.Header, request *RestRequest, reqHeader http.Header) {
	//Set user-agent as special "gopherciser version". version is set from the version package during build.
	useragent := fmt.Sprintf("gopherciser %s", version.Version)
	reqHeader.Set("User-Agent", useragent)

	for k, v := range mainHeader {
		reqHeader[k] = v
	}
	reqHeader.Set("Content-Type", request.ContentType)
	for k, v := range request.ExtraHeaders {
		reqHeader.Set(k, v)
	}
}

func (handler *RestHandler) postWithReader(ctx context.Context, request *RestRequest, client *http.Client, logEntry *logger.LogEntry, headers http.Header) error {
	var method string
	switch request.Method {
	case POST:
		method = http.MethodPost
	case PUT:
		method = http.MethodPut
	default:
		return errors.Errorf("Can only send io.Reader payload with a POST or PUT request. Method<%v>", request.Method)
	}

	req, err := http.NewRequest(method, request.Destination, request.ContentReader)
	if err != nil {
		return errors.Wrap(err, "Failed to create HTTP request")
	}
	req = req.WithContext(ctx)
	handler.newHeader(headers, request, req.Header)
	res, err := client.Do(req)
	request.response = res
	if err != nil {
		return errors.Wrap(err, "HTTP request fail")
	}
	return nil
}

// RoundTrip implement RoundTripper interface
func (transport *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	sentTS := time.Now()

	// log errors on current action if we have one
	logErrors := func(err error) {
		// Set error on action state if we have one.
		if transport.State != nil && transport.State.CurrentActionState != nil {
			transport.State.CurrentActionState.AddErrors(err)
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

	// traffic logging turned on?
	logTraffic := transport.LogEntry.ShouldLogTraffic()

	if logTraffic && transport.trafficLogger != nil {
		body := true
		if isApp || reqSize > constant.MaxBodySize {
			body = false // avoid logging large bodies
		}
		if trafficOut, err := httputil.DumpRequestOut(req, body); err == nil {
			transport.trafficLogger.Sent(append([]byte(fmt.Sprintf("[%d] ", requestID)), trafficOut...))
		} else {
			transport.LogEntry.Log(logger.WarningLevel, "error dumping request", err)
		}
	} else {
		// make sure global and local request counters ticks.
		transport.trafficLogger.Sent(nil)
	}

	resp, err := transport.Transport.RoundTrip(req)
	if err != nil || resp == nil {
		logErrors(errors.Wrapf(err, "failed to perform HTTP request<%s>", req.URL))
		return resp, err
	}

	recTS := time.Now()

	apiPath := apiCallFromPath(req.URL.Path)
	if apiPath != "" {
		buildmetrics.ReportApiResult(apiPath, req.Method, resp.StatusCode, recTS.Sub(sentTS))
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

			transport.LogEntry.LogTrafficMetric(recTS.Sub(sentTS).Nanoseconds(), sent, received, -1, buf.String(), query, "REST")
		}
	}

	if logTraffic && transport.trafficLogger != nil {
		body := !contentIsBinary(resp.Header)
		if respSize > constant.MaxBodySize {
			body = false // avoid logging large bodies
		}
		if trafficIn, err := httputil.DumpResponse(resp, body); err == nil {
			transport.trafficLogger.Received(append([]byte(fmt.Sprintf("[%d] ", requestID)), trafficIn...))
		} else {
			transport.LogEntry.Log(logger.WarningLevel, "error dumping response", err)
		}
	}

	return resp, err
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
	splitApiV1 := strings.Split(path, apiSeparator)
	if len(splitApiV1) < 2 {
		return "" // No api call found in path
	}
	apiCall := splitApiV1[1]
	splitSlash := strings.Split(apiCall, "/")
	if len(splitSlash) < 1 {
		return "" // Nothing after apiSeparator (which is weird)
	}
	return fmt.Sprintf("%s%s", apiSeparator, splitSlash[0])
}
