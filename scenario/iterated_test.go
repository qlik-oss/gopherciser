package scenario

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/qlik-oss/gopherciser/connection"

	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

func TestIterated(t *testing.T) {
	raw := `{
				"label" : "Iterated action to test",
				"action" : "iterated",
				
				"settings" : {
					"iterations" : 3,
					"actions" : [
						{
							"label" : "0.1 seconds delay",
							"action" : "thinktime",
							"settings" : {
								"type": "static",
								"delay" : 0.1
							}
				            
						},
						{
							"label" : "0.2 seconds delay",
							"action" : "thinktime",
							"settings" : {
								"type": "static",
								"delay" : 0.2
							}
				            
						}
					] 
					
				}
			}`

	var item Action
	if err := jsonit.Unmarshal([]byte(raw), &item); err != nil {
		t.Fatal(err)
	}

	if item.Type != ActionIterated {
		t.Fatalf("Invalid action expected<ItemParallel> got<%s>", item.Type)
	}

	settings, ok := item.Settings.(*IteratedSettings)
	if !ok {
		t.Fatalf("Failed to cast settings to ParallelSettings is Type<%T>", item.Settings)
	}

	err := settingsIteratedDecodeCheck(settings)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	session := session.New(ctx, "", time.Second*10, nil, 1, 1, "")
	defer session.Disconnect()

	session.Connection = new(enigmahandlers.SenseConnection)
	sense := enigmahandlers.NewSenseUplink(ctx, nil, session.RequestMetrics, nil)
	sense.MockMode = true
	session.Connection.SetSense(sense)
	session.LogEntry = logger.NewLogEntry(&logger.Log{})
	session.LogEntry.Session = &logger.SessionEntry{}

	startTime := time.Now()
	if err := item.Execute(session, &connection.ConnectionSettings{}); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(startTime)

	if elapsed < (850*time.Millisecond) || elapsed > (950*time.Millisecond) {
		t.Errorf("Unexpected parallel action duration<%v>, expected<0.9s>", elapsed)
	}

}

func settingsIteratedDecodeCheck(settings *IteratedSettings) error {
	if len(settings.Actions) != 2 {
		return fmt.Errorf("expected parallel actions len 2 got <%d>", len(settings.Actions))
	}

	if settings.Iterations != 3 {
		return fmt.Errorf("expected ItemParallel iterations<3> got <%d>", settings.Iterations)
	}

	if settings.Actions[0].Type != ActionThinkTime {
		return fmt.Errorf("Invalid action expected<ItemThinkTime> got<%s>", settings.Actions[0].Type)
	}

	settingsInnerAction, ok := settings.Actions[0].Settings.(*ThinkTimeSettings)
	if !ok {
		return fmt.Errorf("Failed to cast settings to ThinkTimeSettings is Type<%T>", settings.Actions[0].Settings)
	}

	if settingsInnerAction.Delay != 0.1 {
		return fmt.Errorf("expected ItemThinkTime delay<1.0> got <%f>", settingsInnerAction.Delay)
	}

	if settings.Actions[1].Type != ActionThinkTime {
		return fmt.Errorf("Invalid action expected<ItemThinkTime> got<%s>", settings.Actions[1].Type)
	}

	settingsInnerAction, ok = settings.Actions[1].Settings.(*ThinkTimeSettings)
	if !ok {
		return fmt.Errorf("Failed to cast settings to ThinkTimeSettings is Type<%T>", settings.Actions[1].Settings)
	}

	if settingsInnerAction.Delay != 0.2 {
		return fmt.Errorf("expected ItemThinkTime delay<0.5> got <%f>", settingsInnerAction.Delay)
	}

	return nil
}
