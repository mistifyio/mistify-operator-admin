package config

import (
	"time"

	"github.com/hashicorp/go-multierror"
)

// Duration options
var Durations = map[string]time.Duration{
	"Nanosecond":  time.Nanosecond,
	"Microsecond": time.Microsecond,
	"Millisecond": time.Millisecond,
	"Second":      time.Second,
	"Minute":      time.Minute,
	"Hour":        time.Hour,
}

// Enable flag options
var EnableFlags = map[string]bool{
	"true":  true,
	"True":  true,
	"TRUE":  true,
	"false": false,
	"False": false,
	"FALSE": false,
}

// Metrics is the JSON structure and validation for metrics configuration
type Metrics struct {
	ServiceName          string       `json:"service_name"`
	HostName             string       `json:"host_name"`
	EnableHostname       string       `json:"enable_hostname"`
	EnableRuntimeMetrics string       `json:"enable_runtime_metrics"`
	EnableTypePrefix     string       `json:"enable_type_prefix"`
	TimerGranularity     string       `json:"timer_granularity"`
	ProfileInterval      string       `json:"profile_interval"`
	Sinks                []MetricSink `json:"sinks"`
}

// Validate ensures that the metrics configuration is reasonable
func (self *Metrics) Validate() error {
	var result *multierror.Error
	if self.ServiceName == "" {
		result = multierror.Append(result, ErrMetricsNoServiceName)
	}
	if self.EnableHostname != "" {
		if _, ok := EnableFlags[self.EnableHostname]; !ok {
			result = multierror.Append(result, ErrMetricsBadEnableFlag)
		}
	}
	if self.EnableRuntimeMetrics != "" {
		if _, ok := EnableFlags[self.EnableRuntimeMetrics]; !ok {
			result = multierror.Append(result, ErrMetricsBadEnableFlag)
		}
	}
	if self.EnableTypePrefix != "" {
		if _, ok := EnableFlags[self.EnableTypePrefix]; !ok {
			result = multierror.Append(result, ErrMetricsBadEnableFlag)
		}
	}
	if self.TimerGranularity != "" {
		if _, err := time.ParseDuration(self.TimerGranularity); err != nil {
			result = multierror.Append(result, ErrMetricsBadTimerGranularity)
		}
	}
	if self.ProfileInterval != "" {
		if _, err := time.ParseDuration(self.ProfileInterval); err != nil {
			result = multierror.Append(result, ErrMetricsBadProfileInterval)
		}
	}
	for _, sink := range self.Sinks {
		err := sink.Validate()
		if err != nil {
			for _, e := range err.(*multierror.Error).WrappedErrors() {
				result = multierror.Append(result, e)
			}
		}
	}
	return result.ErrorOrNil()
}

// TimerGranularityDuration parses the "TimerGranularity" option and returns a time duration object
func (self *Metrics) TimerGranularityDuration() (time.Duration, error) {
	return time.ParseDuration(self.TimerGranularity)
}

// ProfileIntervalDuration parses the "ProfileInterval" option and returns a time duration object
func (self *Metrics) ProfileIntervalDuration() (time.Duration, error) {
	return time.ParseDuration(self.ProfileInterval)
}
