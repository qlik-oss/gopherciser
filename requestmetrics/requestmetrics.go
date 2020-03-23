package requestmetrics

import (
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/atomichandlers"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	// RequestMetrics keep count on data sent and received. First sent and last received message time
	RequestMetrics struct {
		first    atomichandlers.AtomicTimeStamp
		last     atomichandlers.AtomicTimeStamp
		sent     atomichandlers.AtomicCounter
		received atomichandlers.AtomicCounter
	}
)

// Reset action metrics
func (resp *RequestMetrics) Reset() {
	resp.first.Reset()
	resp.last.Reset()
	resp.sent.Reset()
	resp.received.Reset()
}

// Update action metrics with more data
func (resp *RequestMetrics) Update(sentTS, receivedTS time.Time, sentData, receivedData int64) error {
	var mErr *multierror.Error
	if err := resp.UpdateSent(sentTS, sentData); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	if err := resp.UpdateReceived(receivedTS, receivedData); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	return errors.WithStack(helpers.FlattenMultiError(mErr))
}

// UpdateSent metrics
func (resp *RequestMetrics) UpdateSent(sentTS time.Time, sentData int64) error {
	if sentData < 0 {
		return errors.Errorf("Negative sent data<%d>", sentData)
	}
	resp.sent.Add(uint64(sentData))

	if err := resp.first.SetIfOlder(sentTS); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// UpdateReceived metrics
func (resp *RequestMetrics) UpdateReceived(receivedTS time.Time, receivedData int64) error {
	if receivedData < 0 {
		return errors.Errorf("Negative received data<%d>", receivedData)
	}
	resp.received.Add(uint64(receivedData))
	last := resp.last.Current()
	if receivedTS.After(last) {
		resp.last.Set(receivedTS)
	}

	return nil
}

// Metrics get action metrics
func (resp *RequestMetrics) Metrics() (time.Duration, uint64, uint64) {
	first := resp.first.Current()
	last := resp.last.Current()
	if last.IsZero() {
		last = time.Now()
	}

	var respTime time.Duration
	if first.IsZero() {
		respTime = time.Duration(0) // we never managed to send any requests so there's no response time
	} else {
		respTime = last.Sub(first)
	}

	return respTime, resp.sent.Current(), resp.received.Current()
}
