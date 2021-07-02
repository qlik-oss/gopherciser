package enigmahandlers

import enigma "github.com/qlik-oss/enigma-go/v2"

type (
	// ITrafficLogger interface for traffic logger
	ITrafficLogger interface {
		enigma.TrafficLogger
		RequestCount() uint64
		ResetRequestCount()
	}
)
