package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
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

	HeaderLine struct {
		Key   string
		Value string
	}
)

var filters = marshalFilters(
	"-qNum",
	"+qStateCounts",
	"+qGrandTotalRow",
	"+qPivotDataPages",
	"+qStackedDataPages",
	"+qDataPages",
)

func marshalFilters(filters ...filterType) []byte {
	rawJson, err := json.Marshal(filters)
	if err != nil {
		panic(err)
	}
	return rawJson
}

// NewRegressionLogger creates a new RegressionLoggerCloser with headerLines are
// written at the top of the log file.
func NewRegressionLogger(w io.WriteCloser, headerLines ...HeaderLine) RegressionLoggerCloser {
	fmt.Fprintf(w, "HEADER_KEY\tHEADER_VALUE\n")
	fmt.Fprintf(w, "FILTERS\t%s\n", filters)
	for _, hl := range headerLines {
		fmt.Fprintf(w, "%s\t%s\n", strings.ToUpper(strings.TrimSpace(hl.Key)), strings.TrimSpace(hl.Value))
	}
	fmt.Fprintln(w, "---")
	fmt.Fprintln(w, "ID\tMETA\tDATA")
	return &regressionLogger{w}
}

func (logger *regressionLogger) Close() error {
	return logger.w.Close()
}

// Log the regression analysis data associated with a unique id. Caller is
// responsible for setting a unique id. Pass meta data to support interpretaton
// of log and regression analysis results.
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

	fmt.Fprintf(logger.w, "%s\t%s\t%s\n", dataIDJSON, metaJSON, dataJSON)
	return nil
}
