package scheduler

import (
	"testing"

	"github.com/pkg/errors"
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
