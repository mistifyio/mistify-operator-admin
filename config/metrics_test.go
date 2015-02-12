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
	h.Assert(t, errContains(config.ErrMetricsNoServiceName, err), "expected 'no service name' error")

	metrics.ServiceName = "foo"
	err = metrics.Validate()
	h.Assert(t, errDoesNotContain(config.ErrMetricsNoServiceName, err), "did not expect 'no service name' error")

	metrics.StatsdAddress = "foo"
	err = metrics.Validate()
	h.Assert(t, errContains(config.ErrMetricsBadStatsdAddress, err), "expected 'bad statsd address' error")

	metrics.StatsdAddress = "example.com:http"
	err = metrics.Validate()
	h.Assert(t, errDoesNotContain(config.ErrMetricsBadStatsdAddress, err), "did not expect 'bad statsd address' error")
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
