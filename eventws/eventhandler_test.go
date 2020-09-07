package eventws_test

import (
	"testing"
	"time"

	"github.com/qlik-oss/gopherciser/eventws"
)

func TestEventHandler(t *testing.T) {
	handler := eventws.NewEventHandler()
	if handler == nil {
		t.Fatal("NewEventHandler return nil")
	}

	myevent := "myevent"

	eventTriggerChan := make(chan struct{})

	t.Log("Test normal trigger of event")
	eventFunc1 := handler.RegisterFunc(myevent, func(event eventws.Event) {
		eventTriggerChan <- struct{}{}
	}, false)
	if eventFunc1 == nil {
		t.Fatal("event registered as nil function")
	}

	handler.FakeEvent(eventws.Event{Operation: myevent})

	select {
	case <-eventTriggerChan:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("event failed to trigger")
	}

	handler.DeRegisterFunc(eventFunc1)

	t.Log("Test event after deregistered")
	handler.FakeEvent(eventws.Event{Operation: myevent})

	select {
	case <-eventTriggerChan:
		t.Fatal("event triggered after de-register")
	case <-time.After(500 * time.Millisecond):
	}

	t.Log("Test replayed events")
	handler.RegisterFunc(myevent, func(event eventws.Event) {
		eventTriggerChan <- struct{}{}
	}, true)

loop:
	for i := 0; ; i++ {
		select {
		case <-eventTriggerChan:
		case <-time.After(time.Millisecond):
			if i != 2 {
				t.Errorf("replay triggered %d times, expected 2", i)
			}
			break loop
		}
	}
	handler.DeRegisterFunc(eventFunc1)
}
