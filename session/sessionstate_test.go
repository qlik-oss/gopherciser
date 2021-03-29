package session

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/users"
)

type (
	eventCounter struct {
		trigger int
		close   int
	}
)

func TestState_SendRequestChangeList(t *testing.T) {
	state, event0, event1, event2, event3 := setupStateForCLTest()
	actionState := &action.State{}

	if err := state.SendRequest(actionState, func(ctx context.Context) error {

		if innerErr := state.SendRequest(actionState, func(ctx context.Context) error {
			return manipulateCtxCLValue(ctx, []int{0}, []int{2})
		}); innerErr != nil {
			return errors.Wrap(innerErr, "failed sending inner request")
		}

		return manipulateCtxCLValue(ctx, []int{1}, []int{3})
	}); err != nil {
		t.Fatal("SendRequest failed", err)
	}

	if state.Wait(actionState) {
		t.Fatal("errors during async methods", actionState.Errors())
	}

	if err := assertEventCounter(event0, 1, 0); err != nil {
		t.Error("event0", err)
	}
	if err := assertEventCounter(event1, 1, 0); err != nil {
		t.Error("event1", err)
	}
	if err := assertEventCounter(event2, 0, 1); err != nil {
		t.Error("event2", err)
	}
	if err := assertEventCounter(event3, 0, 1); err != nil {
		t.Error("event3", err)
	}
}

func TestState_QueueRequestChangeList(t *testing.T) {
	state, event0, event1, event2, event3 := setupStateForCLTest()
	actionState := &action.State{}

	state.QueueRequest(func(ctx context.Context) error {
		state.QueueRequest(func(ctx context.Context) error {
			return manipulateCtxCLValue(ctx, []int{0}, []int{2})
		}, actionState, true, "")
		return manipulateCtxCLValue(ctx, []int{1}, []int{3})
	}, actionState, true, "")

	state.Wait(actionState)

	if err := assertEventCounter(event0, 1, 0); err != nil {
		t.Error("event0", err)
	}
	if err := assertEventCounter(event1, 1, 0); err != nil {
		t.Error("event1", err)
	}
	if err := assertEventCounter(event2, 0, 1); err != nil {
		t.Error("event2", err)
	}
	if err := assertEventCounter(event3, 0, 1); err != nil {
		t.Error("event3", err)
	}
}

func TestState_IDMap(t *testing.T) {
	var idmap IDMap
	idmap.newIfNil()

	const KeyID = "keyID"

	// Add a id pair
	valueID := "valueID"
	if err := idmap.Add(KeyID, valueID, nil); err != nil {
		t.Fatal("failed to add id pair:", err)
	}

	// Get id value
	resultValueID := idmap.Get(KeyID)
	if resultValueID != valueID {
		t.Errorf("incorrect value id, expected<%s>, got<%s>", valueID, resultValueID)
	}

	// add duplicate key
	dupErr := idmap.Add("keyID", "valueIDDup", nil)
	switch errors.Cause(dupErr).(type) {
	case DuplicateKeyError:
	default:
		t.Fatalf("expected duplicate key error, got err<%T>:%v", dupErr, dupErr)
	}

	// id should still be first id
	resultValueID = idmap.Get(KeyID)
	if resultValueID != valueID {
		t.Errorf("incorrect value id, expected<%s>, got<%s>", valueID, resultValueID)
	}
}

func TestState_RestMethod(t *testing.T) {
	type dataType = struct {
		Method RestMethod `json:"method"`
	}

	// Test Marshal
	expectedJSON := `{"method":"POST"}`
	data := dataType{
		Method: POST,
	}

	marshalBytes, err := jsonit.Marshal(data)
	if err != nil {
		t.Fatal("failed to marshal RestMethod:", err)
	}

	if expectedJSON != string(marshalBytes) {
		t.Errorf("unexpected marshal of RestMethod: expected<%s> result<%s>", expectedJSON, marshalBytes)
	}

	// Test Unmarshal
	err = jsonit.Unmarshal([]byte(expectedJSON), &data)
	if err != nil {
		t.Fatalf("failed to unmarshal data<%s> to RestMethod: %v", expectedJSON, err)
	}
	if data.Method != POST {
		t.Errorf("unexpected unmarshal of RestMethod: expected<POST> result<%s>", data.Method)
	}
}

func TestState_SessionVariables(t *testing.T) {
	counters := &statistics.ExecutionCounters{}
	state := New(context.Background(), "", 60, nil, 1, 1, "", false, counters)
	state.SetLogEntry(&logger.LogEntry{
		Session: &logger.SessionEntry{
			Thread:  5,
			Session: 56,
		},
	})
	state.User = &users.User{
		UserName: "myuser",
	}

	standardStrings := []string{"{{.UserName}}", "{{.Thread}}", "{{.Session}}"}
	expectedResults := []string{"myuser", "5", "56"}

	if len(standardStrings) != len(expectedResults) {
		t.Fatal("inconsistent count of patterns to test and expected results")
	}

	for i, str := range standardStrings {
		nameTemplate, err := NewSyncedTemplate(str)
		if err != nil {
			t.Errorf("NewSyncedTemplate failed for str<%s>, err:%v", str, err)
			continue
		}
		result, err := state.ReplaceSessionVariables(nameTemplate)
		if err != nil {
			t.Errorf("ReplaceSessionVariables failed for str<%s>, err:%v", str, err)
		}

		if result != expectedResults[i] {
			t.Errorf("ReplaceSessionVariables with str<%s> did not get result<%s>", str, expectedResults[i])
		}
	}

	localString := "mystring"
	if nameTemplate, err := NewSyncedTemplate("{{.Local}}"); err != nil {
		t.Errorf("NewSyncedTemplate failed for str<mystring>, err:%v", err)
	} else {
		result, err := state.ReplaceSessionVariablesWithLocalData(nameTemplate, localString)
		if err != nil {
			t.Error("Failed using ReplaceSessionVariablesWithLocalData:", err)
		} else if result != localString {
			t.Errorf("result<%s> != expected<%s>", result, localString)
		}
	}

	data := struct {
		Filename string
	}{Filename: localString}
	if nameTemplate, err := NewSyncedTemplate("{{.Local.Filename}}"); err != nil {
		t.Errorf("NewSyncedTemplate failed for str<mystring>, err:%v", err)
	} else {
		result, err := state.ReplaceSessionVariablesWithLocalData(nameTemplate, data)
		if err != nil {
			t.Error("Failed using ReplaceSessionVariablesWithLocalData:", err)
		} else if result != localString {
			t.Errorf("result<%s> != expected<%s>", result, localString)
		}
	}
}

func setupStateForCLTest() (*State, *eventCounter, *eventCounter, *eventCounter, *eventCounter) {
	counters := &statistics.ExecutionCounters{}
	state := New(context.Background(), "", 60, nil, 1, 1, "", false, counters)
	state.Rest = NewRestHandler(state.ctx, 64, state.trafficLogger, state.HeaderJar, state.VirtualProxy, state.Timeout)

	event0 := registerEvent(state, 0)
	event1 := registerEvent(state, 1)
	event2 := registerEvent(state, 2)
	event3 := registerEvent(state, 3)

	return state, event0, event1, event2, event3
}

func registerEvent(state *State, handle int) *eventCounter {
	ec := eventCounter{
		trigger: 0,
		close:   0,
	}

	state.RegisterEvent(handle, func(ctx context.Context, actionState *action.State) error {
		ec.trigger++
		return nil
	}, func() {
		ec.close++
	}, true)
	return &ec
}

func manipulateCtxCLValue(ctx context.Context, addChangeValues, addCloseValues []int) error {
	Icl := ctx.Value(enigma.ChangeListsKey{})
	if Icl == nil {
		return nil
	}
	cl, ok := Icl.(*enigma.ChangeLists)
	if !ok {
		return errors.New("ChangeListsKey exists but value not of type *ChangeLists")
	}
	if len(addChangeValues) > 0 {
		cl.Changed = append(cl.Changed, addChangeValues...)
	}
	if len(addCloseValues) > 0 {
		cl.Closed = append(cl.Closed, addCloseValues...)
	}

	return nil
}

func assertEventCounter(ec *eventCounter, ech, ecl int) error {
	if ec == nil {
		return errors.New("event counter is nil")
	}

	errString := ""
	if ec.trigger != ech {
		errString = fmt.Sprintf("trigger changes<%d>, expected<%d>", ec.trigger, ech)
	}

	if ec.close != ecl {
		errString += fmt.Sprintf(" trigger closes<%d>, expected<%d>", ec.close, ecl)
	}

	if errString != "" {
		return errors.New(errString)
	}

	return nil
}
