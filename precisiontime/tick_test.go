package precisiontime

import (
	"sync"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func TestTick(t *testing.T) {
	if _, err := Tick(); err != nil {
		t.Fatal("Error getting tick:", err)
	}
}

func TestTickFormat(t *testing.T) {
	// Comment this out if precision should be evaluated
	t.Skip("Not really needed unless to investigate time precision used for ordering purposes")

	var (
		ticks         = make(map[int64]struct{})
		ticksMapMutex = sync.RWMutex{}
		mu            sync.Mutex // protects multiErrors
		mErr          *multierror.Error
	)

	PrintTime := func(wg *sync.WaitGroup) {
		defer wg.Done()

		if tick, err := Tick(); err == nil {
			checkMapForValue := func() bool {
				defer ticksMapMutex.RUnlock()
				ticksMapMutex.RLock()
				_, ok := ticks[tick]
				return ok
			}

			valueFound := checkMapForValue()

			if valueFound {
				defer mu.Unlock()
				mu.Lock()
				mErr = multierror.Append(mErr, errors.Errorf("Tick already generated so precision is compromised: %d", tick))
			} else {
				ticksMapMutex.Lock()
				defer ticksMapMutex.Unlock()
				ticks[tick] = struct{}{}
			}
		}

	}

	var wg sync.WaitGroup
	// A high number does acutally yield issues..such as 10000, but frequency increases as the number is increased
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go PrintTime(&wg)
	}
	wg.Wait()

	if mErr != nil {
		t.Log("Amount of errors: ", mErr.Len())
		t.Fatal("First error:\n ", mErr.Errors[0].Error())
	}
}
