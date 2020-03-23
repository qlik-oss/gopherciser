package statistics

import "github.com/qlik-oss/gopherciser/atomichandlers"

type (
	RequestStatsMap map[string]*RequestStats

	// RequestStats statistics collector for a REST request
	RequestStats struct {
		method   string
		path     string
		RespAvg  *SampleCollector
		Sent     atomichandlers.AtomicCounter
		Received atomichandlers.AtomicCounter
	}
)

// NewRequestStats creates a new REST request statistics collector
func NewRequestStats(method, path string) *RequestStats {
	return &RequestStats{
		method:  method,
		path:    path,
		RespAvg: NewSampleCollector(),
	}
}

// Method of request
func (request *RequestStats) Method() string {
	if request == nil {
		return ""
	}
	return request.method
}

// Path of request
func (request *RequestStats) Path() string {
	if request == nil {
		return ""
	}
	return request.path
}
