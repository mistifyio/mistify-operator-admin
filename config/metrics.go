package config

import (
	"net"

	"github.com/hashicorp/go-multierror"
)

// Metrics is the JSON structure and validation for metrics configuration
type Metrics struct {
	ServiceName   string `json:"service_name"`
	StatsdAddress string `json:"statsd_address"`
}

// Validate ensures that the metrics configuration is reasonable
func (self *Metrics) Validate() error {
	var result *multierror.Error
	if self.ServiceName == "" {
		result = multierror.Append(result, ErrMetricsNoServiceName)
	}
	if self.StatsdAddress != "" {
		_, _, err := net.SplitHostPort(self.StatsdAddress)
		if err != nil {
			result = multierror.Append(result, ErrMetricsBadStatsdAddress)
			result = multierror.Append(result, err)
		}
	}
	return result.ErrorOrNil()
}
