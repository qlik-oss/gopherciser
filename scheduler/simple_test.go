package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/users"
)

func TestSimpleSched(t *testing.T) {
	sched := &SimpleScheduler{}

	// Validate execution time
	validateSched := func() error {
		_, err := sched.Validate()
		return err
	}
	if err := errors.Cause(validateSched()); err == nil || err.Error() !=
		"Invalid simple scheduler setting:  ExecutionTime<0>" {
		t.Log(err)
		t.Error("ExecutionTime validation failed")
	}
	sched.Settings.ExecutionTime = -1

	// Validate iterations
	if err := errors.Cause(validateSched()); err == nil || err.Error() !=
		"Invalid simple scheduler setting:  Iterations<0>" {
		t.Log(err)
		t.Error("Iterations validation failed")
	}
	sched.Settings.Iterations = -1

	// Validate RampupDelay
	if err := errors.Cause(validateSched()); err == nil || err.Error() !=
		"Invalid simple scheduler setting:  RampupDelay<0.000000>" {
		t.Log(err)
		t.Error("RampupDelay validation failed")
	}
	sched.Settings.RampupDelay = 1.0

	// Validate ConcurrentUsers
	if err := errors.Cause(validateSched()); err == nil || err.Error() !=
		"Invalid simple scheduler setting:  ConcurrentUsers<0>" {
		t.Log(err)
		t.Error("ConcurrentUsers validation failed")
	}
	sched.Settings.ConcurrentUsers = 1

	if err := errors.Cause(validateSched()); err != nil {
		t.Log(err)
		t.Error("validation failed")
	}
}

type resultCollectorAction struct {
	Result []int
}

func (settings resultCollectorAction) Validate() ([]string, error) {
	return nil, nil
}
func (settings *resultCollectorAction) Execute(sessionState *session.State,
	actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	settings.Result = append(settings.Result, sessionState.Randomizer().Rand(100))
}

func TestOnlyInstanceSeed(t *testing.T) {

	iterations := 2
	sched1 := &SimpleScheduler{
		Scheduler: Scheduler{
			SchedType:      SchedSimple,
			InstanceNumber: 2,
			ConnectionSettings: &connection.ConnectionSettings{
				ConnectionSettingsCore: connection.ConnectionSettingsCore{
					Server: "localhost",
					Mode:   connection.WS,
				},
			},
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
	counters := &statistics.ExecutionCounters{}

	deRef := *sched1
	sched2 := &deRef
	sched2.Settings.OnlyInstanceSeed = false

	actionsToAdd := 10
	actions, resultP := FillScenario(actionsToAdd)

	runUserIteration(t, sched1, actions, counters)
	runUserIteration(t, sched2, actions, counters)

	result := *resultP
	t.Log("results:", result)

	// check we have to correct amount of results
	if len(result) != actionsToAdd*iterations*2 {
		t.Fatalf("results<%v> not of expected length<%d>", result, actionsToAdd*iterations)
	}

	// sched 1 iteration 1 and 2 sequences should be the same
	if err := compareSequenceEqual(result[0:actionsToAdd], result[actionsToAdd:(actionsToAdd*2)]); err != nil {
		t.Fatalf("iteration sequences differs although using OnlyInstanceSeed flag: %v", err)
	}

	// sched 1 results should differ from sched 2 results
	if err := compareSequenceNotEqual(result[0:(actionsToAdd*iterations)], result[(actionsToAdd*iterations):(actionsToAdd*iterations*2)]); err != nil {
		t.Fatalf("scheduler sequences doesn't differ when using OnlyInstanceSeed and not, %v", err)
	}
}

func TestReuseUserRandomizer(t *testing.T) {
	// Test to make sure each iteration of re-use users has a unique randomizer
	sched := &SimpleScheduler{
		Scheduler: Scheduler{
			SchedType:      SchedSimple,
			InstanceNumber: 1,
			ConnectionSettings: &connection.ConnectionSettings{
				ConnectionSettingsCore: connection.ConnectionSettingsCore{
					Server: "localhost",
					Mode:   connection.WS,
				},
			},
		},
		Settings: SimpleSchedSettings{
			ExecutionTime:    -1,
			Iterations:       1,
			RampupDelay:      0.1,
			ConcurrentUsers:  1,
			ReuseUsers:       true,
			OnlyInstanceSeed: false,
		},
	}

	counters := &statistics.ExecutionCounters{}

	actionsToAdd := 16
	if actionsToAdd%2 != 0 { // actionsToAdd need to be an even number
		t.Fatalf("actionsToAdd<%d> not divisible by 2", actionsToAdd)
	}
	actions, resultP := FillScenario(actionsToAdd)

	counters.Sessions.Reset()
	runUserIterationReuseUser(t, sched, actions, counters)
	sched.Settings.Iterations = 2
	counters.Sessions.Reset()
	runUserIterationReuseUser(t, sched, actions[:(actionsToAdd/2)], counters)

	result := *resultP
	t.Log("results:", result)

	// verify we have expected amount of results
	if len(result) != actionsToAdd*2 {
		t.Fatalf("results not of expected length<%d>", actionsToAdd*2)
	}

	// divide into 4 sequences
	var seqDiv = actionsToAdd / 2
	seq1 := result[:seqDiv]
	seq2 := result[seqDiv : seqDiv*2]
	seq3 := result[seqDiv*2 : seqDiv*3]
	seq4 := result[seqDiv*3:]

	t.Logf("seq1: %v\n", seq1)
	t.Logf("seq2: %v\n", seq2)
	t.Logf("seq3: %v\n", seq3)
	t.Logf("seq4: %v\n", seq4)

	if err := compareSequenceEqual(seq1, seq3); err != nil {
		t.Fatal(err)
	}

	if err := compareSequenceNotEqual(seq2, seq4); err != nil {
		t.Fatal(err)
	}
}

func runUserIteration(t *testing.T, sched *SimpleScheduler, actions []scenario.Action, counters *statistics.ExecutionCounters) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := sched.iterator(ctx, time.Minute, nil, actions, "", users.NewUserGeneratorNone(), counters); err != nil {
		t.Fatal(err)
	}
}

func runUserIterationReuseUser(t *testing.T, sched *SimpleScheduler, actions []scenario.Action, counters *statistics.ExecutionCounters) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := sched.iterator(ctx, time.Minute, nil, actions, "", users.NewUserGeneratorNone(), counters); err != nil {
		t.Fatal(err)
	}
}

func FillScenario(actionsToAdd int) ([]scenario.Action, *[]int) {
	actionSettings := &resultCollectorAction{}
	testAction := scenario.Action{ActionCore: scenario.ActionCore{}, Settings: actionSettings}

	actions := make([]scenario.Action, 0, actionsToAdd)
	for i := 0; i < actionsToAdd; i++ {
		actions = append(actions, testAction)
	}
	return actions, &actionSettings.Result
}

func compareSequenceEqual(seq1, seq2 []int) error {
	len1 := len(seq1)
	if len(seq2) != len1 {
		return errors.Errorf("%v and %v of different lengths", seq1, seq2)
	}

	for i := 0; i < len1; i++ {
		if seq1[i] != seq2[i] {
			return errors.Errorf("sequence<%v> and sequence<%v> not equal", seq1, seq2)
		}
	}

	return nil
}

func compareSequenceNotEqual(seq1, seq2 []int) error {
	len1 := len(seq1)
	if len(seq2) != len1 {
		return errors.Errorf("%v and %v of different lengths", seq1, seq2)
	}

	for i := 0; i < len1; i++ {
		if seq1[i] != seq2[i] {
			return nil // difference found
		}
	}

	return errors.Errorf("sequence<%v> and sequence<%v> are equal", seq1, seq2)
}

// TestScale can be turned on to test if user creator loop scales well,
// but should not be included in normal tests
func TestScale(t *testing.T) {
	t.SkipNow()

	concurrentusers := 99999999999999999
	executionTime := 300
	rampupdelay := 0.003

	connectionSettings := &connection.ConnectionSettings{
		ConnectionSettingsCore: connection.ConnectionSettingsCore{
			Server: "localhost",
			Mode:   connection.WS,
		},
	}
	sched := &SimpleScheduler{
		Settings: SimpleSchedSettings{
			ConcurrentUsers: concurrentusers,
			ExecutionTime:   executionTime,
			RampupDelay:     rampupdelay,
			Iterations:      1,
		},
		Scheduler: Scheduler{
			ConnectionSettings: connectionSettings,
		},
	}
	scenario := []scenario.Action{
		{
			ActionCore: scenario.ActionCore{
				Type: scenario.ActionThinkTime,
			},
			Settings: scenario.ThinkTimeSettings{
				DistributionSettings: helpers.DistributionSettings{
					Type:  helpers.StaticDistribution,
					Delay: 0.00000001,
				},
			},
		},
	}

	counters := &statistics.ExecutionCounters{}
	err := sched.Execute(context.Background(), nil, time.Second, scenario, "", users.NewUserGeneratorNone(), connectionSettings, counters)
	if err != nil {
		t.Fatal(err)
	}
	totalSessions := counters.Sessions.Current()
	t.Logf("Total sessions: %d", totalSessions)
	expectedSessions := uint64(float64(executionTime)/rampupdelay) + 1
	if expectedSessions != totalSessions {
		t.Errorf("Total sessions: %d, expected: %d", totalSessions, expectedSessions)
	}
}

func BenchmarkSimple(b *testing.B) {
	connectionSettings := &connection.ConnectionSettings{
		ConnectionSettingsCore: connection.ConnectionSettingsCore{
			Server: "localhost",
			Mode:   connection.WS,
		},
	}

	sched := &SimpleScheduler{
		Settings: SimpleSchedSettings{
			ConcurrentUsers: b.N,
			ExecutionTime:   -1,
			RampupDelay:     0.006,
			Iterations:      1,
		},
		Scheduler: Scheduler{
			ConnectionSettings: connectionSettings,
		},
	}
	scenario := []scenario.Action{
		{
			ActionCore: scenario.ActionCore{
				Type: scenario.ActionThinkTime,
			},
			Settings: scenario.ThinkTimeSettings{
				DistributionSettings: helpers.DistributionSettings{
					Type:  helpers.StaticDistribution,
					Delay: 0.00000001,
				},
			},
		},
	}
	counters := &statistics.ExecutionCounters{}
	err := sched.Execute(context.Background(), nil, time.Second, scenario, "", users.NewUserGeneratorNone(), connectionSettings, counters)
	if err != nil {
		b.Fatal(err)
	}
}
