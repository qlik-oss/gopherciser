package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/qlik-oss/gopherciser/helpers"
)

func TestTimeBufConstant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	timeBuf := TimeBuffer{
		Mode:     TimeBufConstant,
		Duration: helpers.TimeDuration(time.Second),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := timeBuf.Wait(ctx, false); err != nil {
		t.Errorf("Error waiting: %+v", err)
	}
	if helpers.IsContextTriggered(ctx) {
		t.Error("context was triggered for mode TimeBufConstant")
	}
}

func TestTimeBufMinDur(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	timeBuf := TimeBuffer{
		Mode:     TimeBufMinDur,
		Duration: helpers.TimeDuration(2 * time.Second),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	expectedError := "illegal start time<0001-01-01 00:00:00 +0000 UTC>"
	if err := timeBuf.Wait(ctx, false); err == nil || err.Error() != expectedError {
		t.Errorf("Expected error<%s>  got: %+v", expectedError, err)
	}
	if helpers.IsContextTriggered(ctx) {
		t.Error("context was triggered for mode TimeBufConstant")
	}

	now := time.Now()
	timeBuf.SetDurationStart(now)
	<-time.After(time.Millisecond * 500)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := timeBuf.Wait(ctx, false); err != nil {
		t.Errorf("Error waiting: %+v", err)
	}

	if now.Add(2 * time.Second).After(time.Now()) {
		t.Error("duration less than expected 2 seconds")
	}
	if helpers.IsContextTriggered(ctx) {
		t.Error("context was triggered for mode TimeBufConstant")
	}
}

func TestTimeBufOnError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	timeBuf := TimeBuffer{
		Mode:     TimeBufOnError,
		Duration: helpers.TimeDuration(time.Second),
	}

	start := time.Now()
	// test no errors
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := timeBuf.Wait(ctx, false); err != nil {
		t.Errorf("Error waiting: %+v", err)
	}
	if helpers.IsContextTriggered(ctx) {
		t.Error("context was triggered for mode TimeBufConstant")
	}
	if start.Add(time.Second + 50*time.Millisecond).Before(time.Now()) {
		t.Error("mode<TimeBufOnError> waited despite errors false")
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := timeBuf.Wait(ctx, true); err != nil {
		t.Errorf("Error waiting: %+v", err)
	}
	if helpers.IsContextTriggered(ctx) {
		t.Error("context was triggered for mode TimeBufConstant")
	}
	if start.Add(time.Second).After(time.Now()) {
		t.Error("mode<TimeBufOnError> did not wait despite errors true")
	}
}

func TestTimeBufMarshaling(t *testing.T) {
	raw := `{"mode":"constant","duration":"25s"}`

	var timeBuf TimeBuffer
	if err := jsonit.Unmarshal([]byte(raw), &timeBuf); err != nil {
		t.Fatal("Failed to unmarshal TimeBuffer", err)
	}

	marshaled, err := jsonit.Marshal(timeBuf)
	if err != nil {
		t.Fatal("Failed to marshal TimeBuffer:", err)
	}
	if raw != string(marshaled) {
		t.Errorf("Failed to marshal TimeBuffer. Expected:\n%s\nGot:\n%s", raw, string(marshaled))
	}
}
