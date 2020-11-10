package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type (
	// RegressionLogger logs data associated with unique ids, with optional
	// meta data.
	RegressionLogger interface {
		Log(dataID string, data interface{}, meta map[string]interface{}) error
	}

	// RegressionLoggerCloser is a closable RegressionLogger, which is typically
	// used when writing log to file.
	RegressionLoggerCloser interface {
		io.Closer
		RegressionLogger
	}

	regressionLogger struct {
		w io.WriteCloser
	}

	filterType string

	// HeaderEntry contains a Key mapped to a regression log meta data Value.
	HeaderEntry struct {
		Key   string
		Value string
	}
)

var filters = marshalFilters(
	"+qStateCounts",
	"+qGrandTotalRow",
	"+qDataPages",
	"+qPivotDataPages",
	"+qStackedDataPages",
	"-qNum",
)

func marshalFilters(filters ...filterType) []byte {
	rawJson, err := json.Marshal(filters)
	if err != nil {
		panic(err)
	}
	return rawJson
}

// NewRegressionLogger creates a new RegressionLoggerCloser with headerEntries
// written in the header of the log.
func NewRegressionLogger(w io.WriteCloser, headerEntries ...HeaderEntry) RegressionLoggerCloser {
	fmt.Fprintf(w, "HEADER_KEY\tHEADER_VALUE\n")
	fmt.Fprintf(w, "FILTERS\t%s\n", filters)
	for _, he := range headerEntries {
		fmt.Fprintf(w, "%s\t%s\n", strings.ToUpper(strings.TrimSpace(he.Key)), strings.TrimSpace(he.Value))
	}
	fmt.Fprintln(w, "---")
	fmt.Fprintln(w, "ID\tMETA\tDATA")
	return &regressionLogger{w}
}

// Close the io.WriteCloser used to create the regressionLogger
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
