package enigmahandlers

import enigma "github.com/qlik-oss/enigma-go/v3"

type (
	// ITrafficLogger interface for traffic logger
	ITrafficLogger interface {
		enigma.TrafficLogger
		RequestCount() uint64
		ResetRequestCount()
	}
)
