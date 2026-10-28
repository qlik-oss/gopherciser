package senseobjects

import (
	"sync"
	"sync/atomic"
)

type (
	// lazySessionObject creates a session object once and publishes it only after it is fully initialized.
	// Concurrent callers wait for an ongoing initialization and get the same object.
	lazySessionObject[T any] struct {
		initMu sync.Mutex
		obj    atomic.Pointer[T]
	}
)

// get returns the initialized object, running create if no object has been successfully initialized yet. If create
// returns an error nothing is published and the next call retries. create must not call get on the same object.
func (l *lazySessionObject[T]) get(create func() (T, error)) (T, error) {
	if obj := l.obj.Load(); obj != nil {
		return *obj, nil
	}

	l.initMu.Lock()
	defer l.initMu.Unlock()
	if obj := l.obj.Load(); obj != nil {
		return *obj, nil
	}

	obj, err := create()
	if err != nil {
		var zero T
		return zero, err
	}
	l.obj.Store(&obj)
	return obj, nil
}

// peek returns the object if initialized, without waiting for an ongoing initialization
func (l *lazySessionObject[T]) peek() (T, bool) {
	if obj := l.obj.Load(); obj != nil {
		return *obj, true
	}
	var zero T
	return zero, false
}
