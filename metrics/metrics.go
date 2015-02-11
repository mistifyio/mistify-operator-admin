package metrics

import (
	"encoding/json"
	"os"
	"sync"
	"syscall"

	gmetrics "github.com/armon/go-metrics"
	"github.com/mistifyio/mistify-operator-admin/config"
)

// Keep track of metrics objects
var metricsObjects map[string]*gmetrics.Metrics = make(map[string]*gmetrics.Metrics)
var mutex sync.Mutex

// Get a metrics object with a particular config, or reuse one that matches
func GetObject(apiConfig *config.Metrics, overrideSink gmetrics.MetricSink) (*gmetrics.Metrics, error) {
	apiConfig = fetchConfig(apiConfig)
	lookup, err := json.Marshal(apiConfig)
	if err != nil {
		return nil, err
	}
	metricsObj, ok := metricsObjects[string(lookup)]
	if ok {
		return metricsObj, nil
	}
	mutex.Lock()
	defer mutex.Unlock()
	metricsObj, err = buildMetricsObject(apiConfig, overrideSink)
	if err != nil {
		return nil, err
	}
	metricsObjects[string(lookup)] = metricsObj
	return metricsObj, nil
}

// Get a new metrics object with a particular config
func NewObject(apiConfig *config.Metrics, overrideSink gmetrics.MetricSink) (*gmetrics.Metrics, error) {
	apiConfig = fetchConfig(apiConfig)
	metricsObj, err := buildMetricsObject(apiConfig, overrideSink)
	if err != nil {
		return nil, err
	}
	return metricsObj, nil
}

// fetchConfig gets the config from arguments or loads the default if one is not provided
func fetchConfig(apiConfig *config.Metrics) *config.Metrics {
	if apiConfig == nil {
		conf := config.Get()
		apiConfig = &conf.Metrics
	}
	return apiConfig
}

// buildMetricsObject generates the metrics object defined by the config
func buildMetricsObject(apiConfig *config.Metrics, overrideSink gmetrics.MetricSink) (*gmetrics.Metrics, error) {
	metricsConfig := buildMetricsObjectConfig(apiConfig)
	var mainSink gmetrics.FanoutSink
	if overrideSink != nil {
		mainSink = make(gmetrics.FanoutSink, 1)
		mainSink[0] = overrideSink
	} else {
		mainSink = make(gmetrics.FanoutSink, len(apiConfig.Sinks))
		for i, sinkConfig := range apiConfig.Sinks {
			sink, err := buildSink(sinkConfig)
			if err != nil {
				return nil, err
			}
			mainSink[i] = sink
		}
	}
	m, err := gmetrics.New(metricsConfig, mainSink)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// buildMetricsObjectConfig generates the config object used by go-metrics
func buildMetricsObjectConfig(apiConfig *config.Metrics) *gmetrics.Config {
	metricsConfig := gmetrics.DefaultConfig(apiConfig.ServiceName)
	myHostName := apiConfig.HostName
	if myHostName != "" && myHostName != "auto" {
		metricsConfig.HostName = myHostName
	}
	if apiConfig.EnableTypePrefix != "" {
		metricsConfig.EnableHostname = config.EnableFlags[apiConfig.EnableHostname]
	}
	if apiConfig.EnableRuntimeMetrics != "" {
		metricsConfig.EnableRuntimeMetrics = config.EnableFlags[apiConfig.EnableRuntimeMetrics]
	}
	if apiConfig.EnableTypePrefix != "" {
		metricsConfig.EnableTypePrefix = config.EnableFlags[apiConfig.EnableTypePrefix]
	}
	if apiConfig.TimerGranularity != "" {
		duration, _ := apiConfig.TimerGranularityDuration()
		metricsConfig.TimerGranularity = duration
	}
	if apiConfig.ProfileInterval != "" {
		duration, _ := apiConfig.ProfileIntervalDuration()
		metricsConfig.ProfileInterval = duration
	}
	return metricsConfig
}

// buildSink creates a sink from the config options
func buildSink(sinkConfig config.MetricSink) (gmetrics.MetricSink, error) {
	if sinkConfig.SinkType == "Statsd" {
		sink, err := gmetrics.NewStatsdSink(sinkConfig.Address)
		if err != nil {
			return nil, err
		}
		return sink, nil
	}
	if sinkConfig.SinkType == "Statsite" {
		sink, err := gmetrics.NewStatsiteSink(sinkConfig.Address)
		if err != nil {
			return nil, err
		}
		return sink, nil
	}
	if sinkConfig.SinkType == "Inmem" {
		interval, _ := sinkConfig.IntervalDuration()
		retain, _ := sinkConfig.RetainDuration()
		sink := gmetrics.NewInmemSink(interval, retain)
		return sink, nil
	}
	if sinkConfig.SinkType == "Test" {
		interval, _ := sinkConfig.IntervalDuration()
		retain, _ := sinkConfig.RetainDuration()
		sink := gmetrics.NewInmemSink(interval, retain)
		gmetrics.NewInmemSignal(sink, syscall.SIGQUIT, os.Stdout)
		return sink, nil
	}
	return &gmetrics.BlackholeSink{}, nil
}
