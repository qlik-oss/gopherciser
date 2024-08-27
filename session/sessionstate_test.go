package session

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/synced"
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

	marshalBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatal("failed to marshal RestMethod:", err)
	}

	if expectedJSON != string(marshalBytes) {
		t.Errorf("unexpected marshal of RestMethod: expected<%s> result<%s>", expectedJSON, marshalBytes)
	}

	// Test Unmarshal
	err = json.Unmarshal([]byte(expectedJSON), &data)
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
		nameTemplate, err := synced.New(str)
		if err != nil {
			t.Errorf("synced.New failed for str<%s>, err:%v", str, err)
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
	if nameTemplate, err := synced.New("{{.Local}}"); err != nil {
		t.Errorf("synced.New failed for str<mystring>, err:%v", err)
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
	if nameTemplate, err := synced.New("{{.Local.Filename}}"); err != nil {
		t.Errorf("synced.New failed for str<mystring>, err:%v", err)
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
	state.Rest = NewRestHandler(state.ctx, state.trafficLogger, state.HeaderJar, state.VirtualProxy, state.Timeout)

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

func TestStateTemplateArtifactMap(t *testing.T) {

	t.Run("nil templateArtifactmap", func(t *testing.T) {
		var tmplAM *TemplateArtifactMap
		id, err := tmplAM.GetIDByTypeAndName("type", "name")
		if id != "" {
			t.Error("id is not empty")
		}
		expectedError := "templateArtifactMap is nil"
		if err.Error() != expectedError {
			t.Errorf("expected error<%s> got error<%s>", expectedError, err.Error())
		}
	})

	t.Run("nil artifactmap", func(t *testing.T) {
		tmplAM := &TemplateArtifactMap{nil}
		id, err := tmplAM.GetIDByTypeAndName("type", "name")
		if id != "" {
			t.Error("id is not empty")
		}
		expectedError := "artifactMap is nil"
		if err.Error() != expectedError {
			t.Errorf("expected error<%s> got error<%s>", expectedError, err.Error())
		}
	})

	am := NewArtifactMap()
	am.Append("type1", &ArtifactEntry{
		ID:     "id1",
		ItemID: "itemID1",
		Name:   "name1",
	})
	am.Append("type2", &ArtifactEntry{
		ID:     "id2",
		ItemID: "itemID2",
		Name:   "name2",
	})
	am.Append("typeNil", nil)
	am.Append("typeNoName", &ArtifactEntry{
		ID:     "IDnoName",
		ItemID: "",
		Name:   "",
	})
	am.Append("typeNoID", &ArtifactEntry{
		ID:     "",
		ItemID: "",
		Name:   "nameNoID",
	})
	artifacts := &TemplateArtifactMap{am}

	t.Run("get type by id", func(t *testing.T) {
		name, err := artifacts.GetNameByTypeAndID("type1", "id1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if name != "name1" {
			t.Errorf("expected<name1> got<%v>", name)
		}
	})
	t.Run("get id by name", func(t *testing.T) {
		id, err := artifacts.GetIDByTypeAndName("type2", "name2")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if id != "id2" {
			t.Errorf("expected<name1> got<%v>", id)
		}
	})

	t.Run("get nil artifact", func(t *testing.T) {
		id, err := artifacts.GetIDByTypeAndName("typeNil", "name2")
		if id != "" {
			t.Error("id is not empty")
		}
		expectedError := "item type<typeNil> id<name2> not found in artifact map"
		if err.Error() != expectedError {
			t.Errorf("expected error<%s> got error<%s>", expectedError, err.Error())
		}
	})

	t.Run("get no name", func(t *testing.T) {
		name, err := artifacts.GetNameByTypeAndID("typeNoName", "IDnoName")
		if name != "" {
			t.Error("id is not empty")
		}
		expectedError := "name is empty string"
		if err.Error() != expectedError {
			t.Errorf("expected error<%s> got error<%s>", expectedError, err.Error())
		}
	})

	t.Run("get no id", func(t *testing.T) {
		id, err := artifacts.GetIDByTypeAndName("typeNoID", "nameNoID")
		if id != "" {
			t.Error("id is not empty")
		}
		expectedError := "id is empty string"
		if err.Error() != expectedError {
			t.Errorf("expected error<%s> got error<%s>", expectedError, err.Error())
		}
	})

	t.Run("empty args", func(t *testing.T) {
		id, err := artifacts.GetIDByTypeAndName("", "")
		if id != "" {
			t.Error("id is not empty")
		}
		expectedError := "first argument artifactType is empty string"
		if err.Error() != expectedError {
			t.Errorf("expected error<%s> got error<%s>", expectedError, err.Error())
		}
	})

	t.Run("empty arg 2", func(t *testing.T) {
		id, err := artifacts.GetIDByTypeAndName("x", "")
		if id != "" {
			t.Error("id is not empty")
		}
		expectedError := "second argument name is empty string"
		if err.Error() != expectedError {
			t.Errorf("expected error<%s> got error<%s>", expectedError, err.Error())
		}
	})

}

func TestPendingWaiter(t *testing.T) {
	counters := &statistics.ExecutionCounters{}
	state := New(context.Background(), "", 120, nil, 1, 1, "", false, counters)
	state.Rest = NewRestHandler(state.ctx, state.trafficLogger, state.HeaderJar, state.VirtualProxy, state.Timeout)
	actionState := &action.State{}

	firstDone := false
	secondDone := false
	thirdDone := false

	state.QueueRequest(func(ctx context.Context) error {
		<-time.After(50 * time.Millisecond)
		firstDone = true
		return nil
	}, actionState, true, "")

	state.Rest.QueueRequestWithCallback(actionState, true, nil, state.LogEntry, func(err error, req *RestRequest) {
		<-time.After(100 * time.Millisecond)
		secondDone = true
		state.QueueRequest(func(ctx context.Context) error {
			<-time.After(500 * time.Millisecond)
			thirdDone = true
			return nil
		}, actionState, true, "")
	})

	state.Wait(actionState)

	if !firstDone {
		t.Error("first request not waited for")
	}
	if !secondDone {
		t.Error("second request not waited for")
	}
	if !thirdDone {
		t.Error("third request not waited for")
	}
}
