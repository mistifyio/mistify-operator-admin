package config_test

import (
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/hashicorp/go-multierror"
	"github.com/mistifyio/mistify-operator-admin/config"
)

func TestMetricsValidate(t *testing.T) {
	metrics := &config.Metrics{}
	var err error

	err = metrics.Validate()
	h.Assert(t, errContains1(config.ErrMetricsNoServiceName, err), "expected 'no service name' error")

	metrics.ServiceName = "foo"
	err = metrics.Validate()
	h.Assert(t, errDoesNotContain1(config.ErrMetricsNoServiceName, err), "did not expect 'no service name' error")

	metrics.StatsdAddress = "foo"
	err = metrics.Validate()
	h.Assert(t, errContains1(config.ErrMetricsBadStatsdAddress, err), "expected 'bad statsd address' error")

	metrics.StatsdAddress = "example.com:http"
	err = metrics.Validate()
	h.Assert(t, errDoesNotContain1(config.ErrMetricsBadStatsdAddress, err), "did not expect 'bad statsd address' error")
}

func errContains1(err error, list error) bool {
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

func errDoesNotContain1(err error, list error) bool {
	return !errContains1(err, list)
}
