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
	"github.com/stretchr/testify/assert"
)

func TestResthandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprint(w, "get!")
			return
		}
		if r.Method == http.MethodPost {
			fmt.Fprint(w, r.Body)
			return
		}
		w.WriteHeader(400)
	}))
	defer ts.Close()

	actionState := action.State{}

	restHandler := NewRestHandler(context.Background(), 32, &enigmahandlers.TrafficLogger{}, NewHeaderJar(), "", 10*time.Second)
	restHandler.Client = http.DefaultClient
	getRequest := RestRequest{
		Method:      GET,
		ContentType: "application/json",
		Destination: ts.URL,
	}
	restHandler.QueueRequest(&actionState, true, &getRequest, &logger.LogEntry{})
	restHandler.WaitForPending()

	assert.Equal(t, "get!", string(getRequest.ResponseBody))

	postRequest := RestRequest{
		Method:      POST,
		ContentType: "application/json",
		Destination: ts.URL,
		Content:     []byte("data!"),
	}
	restHandler.QueueRequest(&actionState, true, &postRequest, &logger.LogEntry{})
	restHandler.WaitForPending()

	assert.Equal(t, "data!", string(postRequest.Content))
}

func TestReqOptions(t *testing.T) {
	options := DefaultReqOptions()
	defaultReqOptions.ExpectedStatusCode[0] = 404

	if options.ExpectedStatusCode[0] == defaultReqOptions.ExpectedStatusCode[0] {
		t.Errorf("Default options changed when modifying instance returned from DefaultReqOptions()")
	}
}

func TestApiExtract(t *testing.T) {
	test1 := "http://myserver:9565/api/v1/items/abc123/action"
	test2 := "http://myserver:9565/api/dcaas"
	test3 := "http://myserver.com/api/v1/evaluation"

	assert.Equal(t, "api/v1/items", apiCallFromPath(test1))
	assert.Equal(t, "", apiCallFromPath(test2))
	assert.Equal(t, "api/v1/evaluation", apiCallFromPath(test3))

}
