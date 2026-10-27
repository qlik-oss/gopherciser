package senseobjects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/logger"
)

type (
	// fakeEngineSocket answers the few engine methods needed by the tests below
	fakeEngineSocket struct {
		responses  chan []byte
		closeOnce  sync.Once
		nextHandle atomic.Int32
	}

	// fakeSessionState executes requests synchronously in the calling goroutine
	fakeSessionState struct{}
)

func newFakeEngineSocket() *fakeEngineSocket {
	socket := &fakeEngineSocket{responses: make(chan []byte, 1024)}
	socket.nextHandle.Store(1)
	return socket
}

func (socket *fakeEngineSocket) WriteMessage(_ int, data []byte) error {
	var request struct {
		ID     int               `json:"id"`
		Method string            `json:"method"`
		Params []json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(data, &request); err != nil {
		return err
	}

	message := map[string]any{"jsonrpc": "2.0", "id": request.ID}
	switch request.Method {
	case "OpenDoc":
		message["result"] = map[string]any{"qReturn": map[string]any{"qType": "Doc", "qHandle": socket.nextHandle.Add(1), "qGenericId": "app"}}
	case "GetDimension":
		var id string
		if err := json.Unmarshal(request.Params[0], &id); err != nil {
			return err
		}
		message["result"] = map[string]any{"qReturn": map[string]any{"qType": "GenericDimension", "qHandle": socket.nextHandle.Add(1), "qGenericId": id}}
	case "CreateSessionObject":
		message["result"] = map[string]any{"qReturn": map[string]any{"qType": "GenericObject", "qHandle": socket.nextHandle.Add(1), "qGenericId": "sessionobject"}}
	case "GetLayout":
		message["result"] = map[string]any{"qLayout": map[string]any{}}
	case "GetProperties", "GetEffectiveProperties":
		message["result"] = map[string]any{"qProp": map[string]any{}}
	default:
		// reply with an error rather than failing the write, a failed write leaves the request waiting for a response
		message["error"] = map[string]any{"code": -32601, "message": fmt.Sprintf("fake engine: unexpected method<%s>", request.Method)}
	}

	response, err := json.Marshal(message)
	if err != nil {
		return err
	}
	socket.responses <- response
	return nil
}

func (socket *fakeEngineSocket) ReadMessage() (int, []byte, error) {
	response, ok := <-socket.responses
	if !ok {
		return 0, nil, fmt.Errorf("fake engine: socket closed")
	}
	return 1, response, nil
}

func (socket *fakeEngineSocket) Close() error {
	socket.closeOnce.Do(func() { close(socket.responses) })
	return nil
}

func (fakeSessionState) BaseContext() context.Context { return context.Background() }
func (fakeSessionState) QueueRequest(f func(ctx context.Context) error, _ *action.State, _ bool, _ string) {
	_ = f(context.Background())
}
func (fakeSessionState) SendRequest(_ *action.State, f func(ctx context.Context) error) error {
	return f(context.Background())
}
func (fakeSessionState) RegisterEvent(int, func(ctx context.Context, actionState *action.State) error, func(), bool) {
}
func (fakeSessionState) DeRegisterEvent(int) {}
func (fakeSessionState) GetEventFunc(int) func(ctx context.Context, actionState *action.State) error {
	return nil
}
func (fakeSessionState) LogDebugf(string, ...interface{})          {}
func (fakeSessionState) LogDetail(logger.LogLevel, string, string) {}
func (fakeSessionState) LogError(error)                            {}
func (fakeSessionState) LogInfo(string, string)                    {}
func (fakeSessionState) LogTrafficMetric(int64, uint64, uint64, int, string, string, string, string) {
}

func newFakeApp(t *testing.T) *App {
	t.Helper()
	ctx := context.Background()
	dialer := enigma.Dialer{
		CreateSocket: func(ctx context.Context, url string, httpHeader http.Header) (enigma.Socket, error) {
			return newFakeEngineSocket(), nil
		},
	}
	global, err := dialer.Dial(ctx, "ws://fake", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { global.DisconnectFromServer() })

	doc, err := global.OpenDoc(ctx, "app", "", "", "", false)
	if err != nil {
		t.Fatal(err)
	}
	return &App{GUID: "app", Doc: doc}
}

// TestAppConcurrentGetDimension reproduces "concurrent map read and map write" when several object layouts
// with cyclic dimensions are handled in parallel. Run with -race.
func TestAppConcurrentGetDimension(t *testing.T) {
	app := newFakeApp(t)
	sessionState := fakeSessionState{}
	actionState := &action.State{}

	// Each goroutine walks the ids from a different offset, so cache hits (no engine request) on one goroutine
	// overlap with another goroutine adding a dimension it has not seen yet.
	const goroutines = 10
	const dimensions = 100
	var wg sync.WaitGroup
	for i := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range dimensions {
				id := fmt.Sprintf("dim%d", (i*dimensions/goroutines+j)%dimensions)
				dim, err := app.GetDimension(sessionState, actionState, id)
				if err != nil {
					t.Error(err)
					return
				}
				if dim.GenericId != id {
					t.Errorf("expected dimension<%s> got<%s>", id, dim.GenericId)
				}
			}
		}()
	}
	wg.Wait()
}

// TestAppConcurrentGetDimensionList verifies that the cached session object pointer is not read unlocked
// while being set. Run with -race.
func TestAppConcurrentGetDimensionList(t *testing.T) {
	app := newFakeApp(t)
	sessionState := fakeSessionState{}
	actionState := &action.State{}

	const goroutines = 20
	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dl, err := app.GetDimensionList(sessionState, actionState)
			if err != nil {
				t.Error(err)
				return
			}
			if dl == nil {
				t.Error("expected dimension list, got nil")
			}
		}()
	}
	wg.Wait()
}
