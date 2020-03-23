package helpers

import (
	"context"
	"time"
)

// IsContextTriggered check if context is "done"
func IsContextTriggered(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// WaitFor context or timeout
func WaitFor(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(duration):
		return
	}
}
