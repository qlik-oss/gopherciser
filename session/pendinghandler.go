package session

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/qlik-oss/gopherciser/atomichandlers"
	"github.com/qlik-oss/gopherciser/globals/constant"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
)

type (
	// PendingHandler handles waiting for pending requests and responses
	PendingHandler struct {
		cond *sync.Cond
		pc   atomichandlers.AtomicCounter
	}
)

// NewPendingHandler new instance of PendingHandler
func NewPendingHandler(size int) PendingHandler {
	return PendingHandler{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

// WaitForPending uses double locking of mutex to wait until mutex is unlocked by
// loop listening for pending req/resp
func (pending *PendingHandler) WaitForPending(ctx context.Context) {
	if helpers.IsContextTriggered(ctx) {
		return
	}

	// Wait until all pending is done
	pending.cond.L.Lock()
	for pending.pc.Current() > 0 {
		pending.cond.Wait()
	}
	pending.cond.L.Unlock()
}

// IncPending increase pending requests
func (pending *PendingHandler) IncPending() {
	pending.pc.Inc()
}

// DecPending increase finished requests
func (pending *PendingHandler) DecPending() {
	pending.pc.Dec()
	if pending.pc.Current() < 1 {
		pending.cond.Broadcast()
	}
}

// QueueRequest Async request,
func (pending *PendingHandler) QueueRequest(baseCtx context.Context, timeout time.Duration,
	f func(ctx context.Context) error, logEntry *logger.LogEntry, onFinished func(err error)) {
	pending.IncPending()

	startTS := time.Now()
	go func() {
		stall := time.Since(startTS)
		defer pending.DecPending()
		ctx, cancel := context.WithTimeout(baseCtx, timeout)
		defer cancel()

		var err error
		var panicErr error

		if onFinished != nil {
			defer func() {
				if panicErr != nil {
					err = panicErr
				}
				onFinished(err)
			}()
		}

		defer helpers.RecoverWithError(&panicErr)

		if stall > constant.MaxStallTime {
			logEntry.LogDetail(logger.WarningLevel, "Goroutine stall", strconv.FormatInt(stall.Nanoseconds(), 10))
		}

		err = f(ctx)
	}()
}
