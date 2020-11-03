package regression

import (
	"encoding/json"
	"fmt"
	"io"
)

type (
	Logger struct {
		w io.WriteCloser
	}
	filterType string
)

const (
	enabled = true
)

var filters = marshalFilters(
	"- qNum",
	"+ qStateCounts",
	"+ qGrandTotalRow",
	"+ qPivotDataPages",
	"+ qStackedDataPages",
	"+ qNum",
)

func marshalFilters(filters ...filterType) []byte {
	rawJson, err := json.Marshal(filters)
	if err != nil {
		panic(err)
	}
	return rawJson
}

func NewLogger(w io.WriteCloser) *Logger {
	fmt.Fprintf(w, "FILTERS %s\n\n", filters)
	return &Logger{w}
}

func (logger *Logger) Close() error {
	return logger.w.Close()
}

func (logger *Logger) Log(dataID string, data interface{}, meta map[string]interface{}) error {
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

	fmt.Fprintf(logger.w, `ID %s\nDATA %s\nMETA %s\n\n`, dataIDJSON, string(dataJSON), string(metaJSON))
	return nil
}
