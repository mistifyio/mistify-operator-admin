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
func GetObject(apiConfig *config.Metrics) (*gmetrics.Metrics, error) {
	// Use the loaded default if one is not provided
	if apiConfig == nil {
		conf := config.Get()
		apiConfig = &conf.Metrics
	}

	// Use the json config to look up the metrics object
	lookup, err := json.Marshal(apiConfig)
	if err != nil {
		return nil, err
	}
	metricsObj, ok := metricsObjects[string(lookup)]
	if ok {
		return metricsObj, nil
	}

	// Make sure multiple processes don't step on each others' toes
	mutex.Lock()
	defer mutex.Unlock()

	// Build the object and store it
	metricsObj, err = buildMetricsObject(apiConfig)
	if err != nil {
		return nil, err
	}
	metricsObjects[string(lookup)] = metricsObj

	return metricsObj, nil
}

// buildMetricsObject generates the metrics object defined by the config
func buildMetricsObject(apiConfig *config.Metrics) (*gmetrics.Metrics, error) {
	metricsConfig := buildMetricsObjectConfig(apiConfig)
	mainSink := make(gmetrics.FanoutSink, len(apiConfig.Sinks))
	for i, sinkConfig := range apiConfig.Sinks {
		sink, err := buildSink(sinkConfig)
		if err != nil {
			return nil, err
		}
		mainSink[i] = sink
	}
	return gmetrics.New(metricsConfig, mainSink)
}

// buildMetricsObjectConfig generates the config object used by go-metrics
func buildMetricsObjectConfig(apiConfig *config.Metrics) *gmetrics.Config {
	metricsConfig := gmetrics.DefaultConfig(apiConfig.ServiceName)
	myHostName := apiConfig.HostName
	if myHostName != "" && myHostName != "auto" {
		metricsConfig.HostName = myHostName
	}
	metricsConfig.EnableHostname = apiConfig.EnableHostname
	metricsConfig.EnableRuntimeMetrics = apiConfig.EnableRuntimeMetrics
	metricsConfig.EnableTypePrefix = apiConfig.EnableTypePrefix
	metricsConfig.TimerGranularity = apiConfig.TimerGranularityDuration()
	metricsConfig.ProfileInterval = apiConfig.ProfileIntervalDuration()
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
		sink := gmetrics.NewInmemSink(sinkConfig.IntervalDuration(), sinkConfig.RetainDuration())
		return sink, nil
	}
	if sinkConfig.SinkType == "Test" {
		sink := gmetrics.NewInmemSink(sinkConfig.IntervalDuration(), sinkConfig.RetainDuration())
		gmetrics.NewInmemSignal(sink, syscall.SIGQUIT, os.Stdout)
		return sink, nil
	}
	return &gmetrics.BlackholeSink{}, nil
}
