package session

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/pending"
	"github.com/qlik-oss/gopherciser/requestmetrics"
	"github.com/stretchr/testify/assert"
)

func TestResthandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if _, err := fmt.Fprint(w, "get!"); err != nil {
				t.Error(err)
			}
			return
		}
		if r.Method == http.MethodPost {
			if _, err := fmt.Fprint(w, r.Body); err != nil {
				t.Error(err)
			}
			return
		}
		w.WriteHeader(400)
	}))
	defer ts.Close()

	actionState := action.State{}
	pendingHandler := pending.NewHandler()

	restHandler := NewRestHandler(context.Background(), &enigmahandlers.TrafficLogger{}, NewHeaderJar(), "", 10*time.Second, &pendingHandler, &requestmetrics.RequestMetrics{})
	restHandler.Client = http.DefaultClient
	getRequest := RestRequest{
		Method:      GET,
		ContentType: "application/json",
		Destination: ts.URL,
	}
	restHandler.QueueRequest(&actionState, true, &getRequest, &logger.LogEntry{})
	timeOutContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pendingHandler.WaitForPending(timeOutContext)

	assert.Equal(t, "get!", string(getRequest.ResponseBody))

	postRequest := RestRequest{
		Method:      POST,
		ContentType: "application/json",
		Destination: ts.URL,
		Content:     []byte("data!"),
	}
	restHandler.QueueRequest(&actionState, true, &postRequest, &logger.LogEntry{})
	pendingHandler.WaitForPending(timeOutContext)

	assert.Equal(t, "data!", string(postRequest.Content))
}

func TestReqOptions(t *testing.T) {
	options := DefaultReqOptions()
	defaultReqOptions.ExpectedStatusCode[0] = 404

	if options.ExpectedStatusCode[0] == defaultReqOptions.ExpectedStatusCode[0] {
		t.Errorf("Default options changed when modifying instance returned from DefaultReqOptions()")
	}
}

func TestApiCallFromPath(t *testing.T) {
	test1 := "api/v1/items/abc123/action"
	test2 := "api/dcaas"
	test3 := "api/v1/evaluation"

	assert.Equal(t, "api/v1/items", apiCallFromPath(test1))
	assert.Equal(t, "", apiCallFromPath(test2))
	assert.Equal(t, "api/v1/evaluation", apiCallFromPath(test3))
}

func TestPrependURLPath(t *testing.T) {
	for _, tc := range []struct {
		inputURL          string
		inputPath         string
		expectedOutputURL string
	}{
		{
			inputURL:          "http://something.com/a/b/def",
			inputPath:         "x",
			expectedOutputURL: "http://something.com/x/a/b/def",
		},
		{
			inputURL:          "http://something.com/a/b/def",
			inputPath:         "/x/y/z",
			expectedOutputURL: "http://something.com/x/y/z/a/b/def",
		},
		{
			inputURL:          "http://something.com/path/to/endpoint?myVar=10&myvar2=aString",
			inputPath:         "/myProxy/",
			expectedOutputURL: "http://something.com/myProxy/path/to/endpoint?myVar=10&myvar2=aString",
		},
	} {
		outputURL, err := prependURLPath(tc.inputURL, tc.inputPath)
		if err != nil {
			t.Error(err)
		}
		if outputURL != tc.expectedOutputURL {
			t.Errorf(`Expeted url<%s>, but got url<%s>`, tc.expectedOutputURL, outputURL)
		}
	}
}
