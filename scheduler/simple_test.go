package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/users"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/connection"
)

func TestSimpleSched(t *testing.T) {
	sched := &SimpleScheduler{}

	// Validate execution time
	if err := errors.Cause(sched.Validate()); err == nil || err.Error() !=
		"Invalid simple scheduler setting:  ExecutionTime<0>" {
		t.Log(err)
		t.Error("ExecutionTime validation failed")
	}
	sched.Settings.ExecutionTime = -1

	// Validate iterations
	if err := errors.Cause(sched.Validate()); err == nil || err.Error() !=
		"Invalid simple scheduler setting:  Iterations<0>" {
		t.Log(err)
		t.Error("Iterations validation failed")
	}
	sched.Settings.Iterations = -1

	// Validate RampupDelay
	if err := errors.Cause(sched.Validate()); err == nil || err.Error() !=
		"Invalid simple scheduler setting:  RampupDelay<0.000000>" {
		t.Log(err)
		t.Error("RampupDelay validation failed")
	}
	sched.Settings.RampupDelay = 1.0

	// Validate ConcurrentUsers
	if err := errors.Cause(sched.Validate()); err == nil || err.Error() !=
		"Invalid simple scheduler setting:  ConcurrentUsers<0>" {
		t.Log(err)
		t.Error("ConcurrentUsers validation failed")
	}
	sched.Settings.ConcurrentUsers = 1

	if err := errors.Cause(sched.Validate()); err != nil {
		t.Log(err)
		t.Error("validation failed")
	}
}

type TestOnlyInstanceSeedAction struct {
	Result []int
}

func (settings TestOnlyInstanceSeedAction) Validate() error {
	return nil
}
func (settings *TestOnlyInstanceSeedAction) Execute(sessionState *session.State,
	actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	settings.Result = append(settings.Result, sessionState.Randomizer().Rand(100))
}

func TestOnlyInstanceSeed(t *testing.T) {

	iterations := 2
	sched1 := &SimpleScheduler{
		Scheduler: Scheduler{
			SchedType:          SchedSimple,
			InstanceNumber:     2,
			connectionSettings: &connection.ConnectionSettings{},
		},
		Settings: SimpleSchedSettings{
			ExecutionTime:    -1,
			Iterations:       iterations,
			RampupDelay:      0.1,
			ConcurrentUsers:  1,
			ReuseUsers:       false,
			OnlyInstanceSeed: true,
		},
	}

	deRef := *sched1
	sched2 := &deRef
	sched2.Settings.OnlyInstanceSeed = false

	actionSettings := &TestOnlyInstanceSeedAction{}
	testAction := scenario.Action{ActionCore: scenario.ActionCore{}, Settings: actionSettings}

	actionsToAdd := 10

	actions := make([]scenario.Action, 0, actionsToAdd)
	for i := 0; i < actionsToAdd; i++ {
		actions = append(actions, testAction)
	}

	runUserIteration(t, sched1, actions)
	runUserIteration(t, sched2, actions)

	result := actionSettings.Result
	t.Log("results:", result)

	// check we have to correct amount of results
	if len(result) != actionsToAdd*iterations*2 {
		t.Fatalf("results<%v> not of expected length<%d>", result, actionsToAdd*iterations)
	}

	// sched 1 iteration 1 and 2 sequences should be the same
	seq11 := result[0:actionsToAdd]
	seq12 := result[actionsToAdd:(actionsToAdd * 2)]
	for i := 0; i < actionsToAdd; i++ {
		if seq11[i] != seq12[i] {
			t.Fatalf("iteration sequences differs although using OnlyInstanceSeed flag, sequences: 1<%v> 2<%v>", seq11, seq12)
		}
	}

	// sched 1 results should differ from sched 2 results
	seq1 := result[0:(actionsToAdd * iterations)]
	seq2 := result[(actionsToAdd * iterations):(actionsToAdd * iterations * 2)]
	differenceFound := false
	for i := 0; i < actionsToAdd*iterations; i++ {
		if seq1[i] != seq2[i] {
			differenceFound = true
			break
		}
	}
	if !differenceFound {
		t.Fatalf("scheduler sequences doesn't differ when using OnlyInstanceSeed and not, sequences: 1<%v> 2<%v>", seq1, seq2)
	}
}

func runUserIteration(t *testing.T, sched *SimpleScheduler, actions []scenario.Action) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := sched.iteratorNewUsers(ctx, time.Minute, nil, actions, "", users.NewUserGeneratorNone()); err != nil {
		t.Fatal(err)
	}
}
