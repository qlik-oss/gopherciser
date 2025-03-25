package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/goccy/go-json"

	"github.com/qlik-oss/gopherciser/version"
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
	"-qInExtRow",
)

func marshalFilters(filters ...filterType) []byte {
	rawJson, err := json.Marshal(filters)
	if err != nil {
		panic(err)
	}
	return rawJson
}

func (logger *regressionLogger) write(record ...string) {
	for i, r := range record {
		record[i] = replacer.Replace(r)
	}
	if _, err := fmt.Fprintln(logger.w, strings.Join(record, "\t")); err != nil {
		_, _ = fmt.Fprint(os.Stderr, "error writting to regressionlogger", err)
	}
}

// NewRegressionLogger creates a new RegressionLoggerCloser with headerEntries
// written in the header of the log.
func NewRegressionLogger(w io.WriteCloser, headerEntries ...HeaderEntry) RegressionLoggerCloser {
	logger := &regressionLogger{w}
	logger.write("HEADER_KEY", "HEADER_VALUE")
	logger.write("FILTERS", string(filters))
	logger.write("VERSION", version.Version)
	for _, he := range headerEntries {
		logger.write(strings.ToUpper(strings.TrimSpace(he.Key)), strings.TrimSpace(he.Value))
	}
	logger.write("---")
	logger.write("ID", "META", "DATA")
	return logger
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

	logger.write(string(dataIDJSON), string(metaJSON), string(dataJSON))
	return nil
}
