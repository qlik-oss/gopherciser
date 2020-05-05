package helpers

import (
	"fmt"
	"testing"
)

func TestMarshalTimeDuration(t *testing.T) {
	td := TimeDuration(2000000000)
	expected := "\"1s\""
	err := td.UnmarshalJSON([]byte(expected))
	if err != nil {
		t.Errorf("got error during unmarshal <%s>", err)
	}
	json, err := td.MarshalJSON()
	if err != nil {
		t.Errorf("got error during marshal <%s>", err)
	}
	if fmt.Sprintf("%s", json) != expected {
		t.Errorf("expected<%s> got %s", expected, json)
	}
}
