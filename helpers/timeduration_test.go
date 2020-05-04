package helpers

import (
	"fmt"
	"testing"
)

func TestMarshalTimeDuration(t *testing.T) {
	td := TimeDuration(2000000000)
	expected := "\"1s\""
	td.UnmarshalJSON([]byte(expected))
	json, err := td.MarshalJSON()
	if err != nil {
		t.Errorf("got error during marshal <%s>", err)
	}
	if fmt.Sprintf("%s", json) != expected {
		t.Errorf("expected<%s> got %s", expected, json)
	}
}
