package logger

import (
	"encoding/json"
	"fmt"
	"io"
)

type (
	RegressionLogger interface {
		Log(dataID string, data interface{}, meta map[string]interface{}) error
	}

	RegressionLoggerCloser interface {
		io.Closer
		RegressionLogger
	}

	regressionLogger struct {
		w io.WriteCloser
	}

	filterType string
)

var filters = marshalFilters(
	"- qNum",
	"+ qStateCounts",
	"+ qGrandTotalRow",
	"+ qPivotDataPages",
	"+ qStackedDataPages",
	"+ qDataPages",
)

func marshalFilters(filters ...filterType) []byte {
	rawJson, err := json.Marshal(filters)
	if err != nil {
		panic(err)
	}
	return rawJson
}

func NewRegressionLogger(w io.WriteCloser) RegressionLoggerCloser {
	fmt.Fprintf(w, "FILTERS %s\n", filters)
	return &regressionLogger{w}
}

func (logger *regressionLogger) Close() error {
	return logger.w.Close()
}

func (logger *regressionLogger) Log(dataID string, data interface{}, meta map[string]interface{}) error {
	dataIDJSON, err := json.Marshal(dataID)
	if err != nil {
		return err
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	fmt.Fprintf(logger.w, "\nID %s\nDATA %s\nMETA %s\n", dataIDJSON, dataJSON, metaJSON)
	return nil
}
