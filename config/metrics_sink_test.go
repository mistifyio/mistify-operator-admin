package config_test

import (
	"testing"
	"time"

	h "github.com/bakins/test-helpers"
	"github.com/hashicorp/go-multierror"
	"github.com/mistifyio/mistify-operator-admin/config"
)

func TestMetricSinkValidate(t *testing.T) {
	sink := &config.MetricSink{}
	var err error

	err = sink.Validate()
	h.Assert(t, errContains(config.ErrMetricsBadSinkType, err), "expected 'bad sink type' error")

	sink.SinkType = "foo"
	err = sink.Validate()
	h.Assert(t, errContains(config.ErrMetricsBadSinkType, err), "expected 'bad sink type' error")

	sink.SinkType = "Inmem"
	err = sink.Validate()
	h.Assert(t, errDoesNotContain(config.ErrMetricsBadSinkType, err), "did not expect 'bad sink type' error")

	sink.Interval = "foo"
	err = sink.Validate()
	h.Assert(t, errContains(config.ErrMetricsBadInmemInterval, err), "expected 'bad in-memory interval' error")

	sink.Interval = "5ms"
	err = sink.Validate()
	h.Assert(t, errDoesNotContain(config.ErrMetricsBadInmemInterval, err), "did not expect 'bad in-memory interval' error")

	sink.Retain = "bar"
	err = sink.Validate()
	h.Assert(t, errContains(config.ErrMetricsBadInmemRetain, err), "expected 'bad in-memory retain duration' error")

	sink.Retain = "3s"
	err = sink.Validate()
	h.Assert(t, errDoesNotContain(config.ErrMetricsBadInmemRetain, err), "did not expect 'bad in-memory retain duration' error")

	sink.SinkType = "Statsd"
	sink.Interval = "foo"
	sink.Retain = "bar"
	err = sink.Validate()
	h.Assert(t, errDoesNotContain(config.ErrMetricsBadInmemInterval, err), "did not expect 'bad in-memory interval' error")
	h.Assert(t, errDoesNotContain(config.ErrMetricsBadInmemRetain, err), "did not expect 'bad in-memory retain duration' error")
	h.Assert(t, errContains(config.ErrMetricsNoAddress, err), "expected 'no address' error")

	sink.Address = "foobar"
	err = sink.Validate()
	h.Assert(t, errDoesNotContain(config.ErrMetricsNoAddress, err), "did not expect 'no address' error")
}

func TestIntervalDuration(t *testing.T) {
	sink := &config.MetricSink{}
	sink.Interval = "foo"
	duration, err := sink.IntervalDuration()
	h.Assert(t, err != nil, "expected an interval of 'foo' to generate an error")

	sink.Interval = "5ms"
	duration, err = sink.IntervalDuration()
	h.Assert(t, err == nil, "expected an interval of '5ms' not to generate an error")
	h.Equals(t, duration, 5*time.Millisecond)
}

func TestRetainDuration(t *testing.T) {
	sink := &config.MetricSink{}
	sink.Retain = "foo"
	duration, err := sink.RetainDuration()
	h.Assert(t, err != nil, "expected a retain duration of 'foo' to generate an error")

	sink.Retain = "5ms"
	duration, err = sink.RetainDuration()
	h.Assert(t, err == nil, "expected a retain duration of '5ms' not to generate an error")
	h.Equals(t, duration, 5*time.Millisecond)
}

func errContains(err error, list error) bool {
	merr, ok := list.(*multierror.Error)
	if !ok {
		return false
	}

	errList := merr.Errors
	for _, e := range errList {
		if err == e {
			return true
		}
	}
	return false
}

func errDoesNotContain(err error, list error) bool {
	return !errContains(err, list)
}
