package config

import (
	"time"

	gmetrics "github.com/armon/go-metrics"
	"github.com/hashicorp/go-multierror"
)

// Sink type options
var sinkTypes = map[string]bool{
	"Test": true,
}

// Duration options
var durations = map[string]time.Duration{
	"Nanosecond":  time.Nanosecond,
	"Microsecond": time.Microsecond,
	"Millisecond": time.Millisecond,
	"Second":      time.Second,
	"Minute":      time.Minute,
	"Hour":        time.Hour,
}

// Metrics is the JSON structure and validation for metrics configuration
type Metrics struct {
	SinkType             string `json:"sink_type"`
	ServiceName          string `json:"service_name"`
	HostName             string `json:"host_name"`
	EnableHostname       bool   `json:"enable_hostname"`
	EnableRuntimeMetrics bool   `json:"enable_runtime_metrics"`
	EnableTypePrefix     bool   `json:"enable_type_prefix"`
	TimerGranularity     string `json:"timer_granularity"`
	ProfileInterval      string `json:"profile_interval"`
}

// Validate ensures that the metrics configuration is reasonable
func (metrics *Metrics) Validate() error {
	var result *multierror.Error
	if _, ok := sinkTypes[metrics.SinkType]; !ok {
		result = multierror.Append(result, ErrMetricsBadSinkType)
	}
	if metrics.ServiceName == "" {
		result = multierror.Append(result, ErrMetricsNoServiceName)
	}
	if _, ok := durations[metrics.TimerGranularity]; !ok {
		result = multierror.Append(result, ErrMetricsBadTimerGranularity)
	}
	if _, ok := durations[metrics.ProfileInterval]; !ok {
		result = multierror.Append(result, ErrMetricsBadProfileInterval)
	}
	return result.ErrorOrNil()
}

// MetricsObjectConfig generates the config object used by go-metrics
func (self *Metrics) MetricsObjectConfig() *gmetrics.Config {
	metricsConfig := gmetrics.DefaultConfig(self.ServiceName)
	myHostName := self.HostName
	if myHostName != "" && myHostName != "auto" {
		metricsConfig.HostName = myHostName
	}
	metricsConfig.EnableHostname = self.EnableHostname
	metricsConfig.EnableRuntimeMetrics = self.EnableRuntimeMetrics
	metricsConfig.EnableTypePrefix = self.EnableTypePrefix
	metricsConfig.TimerGranularity = durations[self.TimerGranularity]
	metricsConfig.ProfileInterval = durations[self.ProfileInterval]
	return metricsConfig
}
