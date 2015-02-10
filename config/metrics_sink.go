package config

import (
	"time"

	"github.com/hashicorp/go-multierror"
)

// Sink type options
var sinkTypes = map[string]bool{
	"Blackhole": true,
	"Inmem":     true,
	"Statsd":    true,
	"Statsite":  true,
	"Test":      true,
}

// MetricSink is the JSON structure and validation for metric sink configuration
type MetricSink struct {
	SinkType string `json:"sink_type"`
	Address  string `json:"address"`  // Used by Statsd and Statsite
	Interval string `json:"interval"` // Used by Inmem and Test
	Retain   string `json:"retain"`   // Used by Inmem and Test
}

// Validate ensures that the metrics configuration is reasonable
func (self *MetricSink) Validate() error {
	var result *multierror.Error
	if _, ok := sinkTypes[self.SinkType]; !ok {
		result = multierror.Append(result, ErrMetricsBadSinkType)
	}
	if self.SinkType == "Inmem" {
		if _, err := time.ParseDuration(self.Interval); err != nil {
			result = multierror.Append(result, ErrMetricsBadInmemInterval)
		}
		if _, err := time.ParseDuration(self.Retain); err != nil {
			result = multierror.Append(result, ErrMetricsBadInmemRetain)
		}
	}
	if self.SinkType == "Statsd" || self.SinkType == "Statsite" {
		if self.Address == "" {
			result = multierror.Append(result, ErrMetricsNoAddress)
		}
	}
	return result.ErrorOrNil()
}

// IntervalDuration parses the "Interval" option and returns a time duration object
func (self *MetricSink) IntervalDuration() (time.Duration, error) {
	return time.ParseDuration(self.Interval)
}

// RetainDuration parses the "Retain" option and returns a time duration object
func (self *MetricSink) RetainDuration() (time.Duration, error) {
	return time.ParseDuration(self.Retain)
}
