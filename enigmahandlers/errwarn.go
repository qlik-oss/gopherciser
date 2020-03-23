package enigmahandlers

import "github.com/qlik-oss/gopherciser/atomichandlers"

type (
	// ErrWarn keeps track of errors and warnings within action + total
	ErrWarn struct {
		warnings    atomichandlers.AtomicCounter
		errors      atomichandlers.AtomicCounter
		totWarnings atomichandlers.AtomicCounter
		totErrors   atomichandlers.AtomicCounter
	}
)

// IncWarn Add 1 to warnings counter
func (ew *ErrWarn) IncWarn() {
	ew.warnings.Inc()
	ew.totWarnings.Inc()
}

// IncErr Add 1 to errors counter
func (ew *ErrWarn) IncErr() {
	ew.errors.Inc()
	ew.totErrors.Inc()
}

// Reset reset error and warnings counters
func (ew *ErrWarn) Reset() {
	ew.errors.Reset()
	ew.warnings.Reset()
}

// Errors report error count
func (ew *ErrWarn) Errors() uint64 {
	return ew.errors.Current()
}

// Warnings report warning count
func (ew *ErrWarn) Warnings() uint64 {
	return ew.warnings.Current()
}

// TotErrors report total error count
func (ew *ErrWarn) TotErrors() uint64 {
	return ew.totErrors.Current()
}

// TotWarnings report total warnings count
func (ew *ErrWarn) TotWarnings() uint64 {
	return ew.totWarnings.Current()
}
