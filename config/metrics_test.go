package config_test

import (
	"testing"
	"time"

	h "github.com/bakins/test-helpers"
	"github.com/hashicorp/go-multierror"
	"github.com/mistifyio/mistify-operator-admin/config"
)

func TestMetricsValidate(t *testing.T) {
	metrics := &config.Metrics{}
	var err error

	err = metrics.Validate()
	h.Assert(t, errContains(config.ErrMetricsNoServiceName, err), "expected 'no service name' error")

	metrics.ServiceName = "foo"
	err = metrics.Validate()
	h.Assert(t, errDoesNotContain(config.ErrMetricsNoServiceName, err), "did not expect 'no service name' error")

	metrics.TimerGranularity = "foo"
	err = metrics.Validate()
	h.Assert(t, errContains(config.ErrMetricsBadTimerGranularity, err), "expected 'bad timer granularity' error")

	metrics.TimerGranularity = "5*Second"
	err = metrics.Validate()
	h.Assert(t, errDoesNotContain(config.ErrMetricsBadTimerGranularity, err), "did not expect 'bad timer granularity' error")

	metrics.ProfileInterval = "foo"
	err = metrics.Validate()
	h.Assert(t, errContains(config.ErrMetricsBadProfileInterval, err), "expected 'bad profile interval' error")

	metrics.ProfileInterval = "5*Second"
	err = metrics.Validate()
	h.Assert(t, errDoesNotContain(config.ErrMetricsBadProfileInterval, err), "did not expect 'bad profile interval' error")

	// NB: Sink's Validate() is tested in metrics_sink_test.go
	sink := config.MetricSink{}
	metrics.Sinks = []config.MetricSink{sink}
	err = metrics.Validate()
	h.Assert(t, errContains(config.ErrMetricsBadSinkType, err), "expected 'bad sink type' error")
}

func TestParseDuration(t *testing.T) {
	duration, err := config.ParseDuration("foo")
	h.Assert(t, err == config.ErrMetricsBadDuration, "expected a duration of 'foo' to generate an error")
	duration, err = config.ParseDuration("Millisecond")
	h.Assert(t, err == nil, "expected a duration of 'Millisecond' not to generate an error")
	h.Equals(t, duration, time.Millisecond)
	duration, err = config.ParseDuration("5*Millisecond")
	h.Assert(t, err == nil, "expected a duration of '5*Millisecond' not to generate an error")
	h.Equals(t, duration, 5*time.Millisecond)
}

func TestTimerGranularityDuration(t *testing.T) {
	metrics := &config.Metrics{}
	metrics.TimerGranularity = "foo"
	duration, err := metrics.TimerGranularityDuration()
	h.Assert(t, err == config.ErrMetricsBadDuration, "expected a timer granularity of 'foo' to generate an error")

	metrics.TimerGranularity = "5*Millisecond"
	duration, err = metrics.TimerGranularityDuration()
	h.Assert(t, err == nil, "expected a timer granularity of '5*Millisecond' not to generate an error")
	h.Equals(t, duration, 5*time.Millisecond)
}

func TestProfileIntervalDuration(t *testing.T) {
	metrics := &config.Metrics{}
	metrics.ProfileInterval = "foo"
	duration, err := metrics.ProfileIntervalDuration()
	h.Assert(t, err == config.ErrMetricsBadDuration, "expected a profile interval of 'foo' to generate an error")

	metrics.ProfileInterval = "5*Millisecond"
	duration, err = metrics.ProfileIntervalDuration()
	h.Assert(t, err == nil, "expected a profile interval of '5*Millisecond' not to generate an error")
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
