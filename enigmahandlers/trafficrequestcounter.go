package enigmahandlers

import (
	"github.com/qlik-oss/gopherciser/atomichandlers"
	"github.com/qlik-oss/gopherciser/globals"
)

type (
	// TrafficRequestCounter implementation of enigma.trafficLogger interface
	TrafficRequestCounter struct {
		Requests *atomichandlers.AtomicCounter
	}
)

// NewTrafficRequestCounter create new instance of traffic request counter
func NewTrafficRequestCounter() *TrafficRequestCounter {
	var req atomichandlers.AtomicCounter
	return &TrafficRequestCounter{
		Requests: &req,
	}
}

// Opened implements trafficLogger interface
func (tl *TrafficRequestCounter) Opened() {}

// Sent count sent requests
func (tl *TrafficRequestCounter) Sent(message []byte) {
	if tl.Requests != nil {
		tl.Requests.Inc()
	}
	globals.Requests.Inc()
}

// Received implements trafficLogger interface
func (tl *TrafficRequestCounter) Received(message []byte) {}

// Closed implements trafficLogger interface
func (tl *TrafficRequestCounter) Closed() {}

// RequestCount get current request count
func (tl *TrafficRequestCounter) RequestCount() uint64 {
	return tl.Requests.Current()
}

// ResetRequestCount reset current request count
func (tl *TrafficRequestCounter) ResetRequestCount() {
	tl.Requests.Reset()
}
