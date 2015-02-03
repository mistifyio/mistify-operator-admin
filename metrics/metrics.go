package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"
	"time"

	gmetrics "github.com/armon/go-metrics"
	conf "github.com/mistifyio/mistify-operator-admin/config"
)

// Keep track of metrics objects
var metricsObjects map[string]*gmetrics.Metrics = make(map[string]*gmetrics.Metrics)

// Duration options for config
var durations = map[string]time.Duration{
	"Nanosecond":  time.Nanosecond,
	"Microsecond": time.Microsecond,
	"Millisecond": time.Millisecond,
	"Second":      time.Second,
	"Minute":      time.Minute,
	"Hour":        time.Hour,
}

// Get a metrics object with a particular config, or reuse one that matches
func GetObject(cfg *conf.Metrics) (*gmetrics.Metrics, error) {
	// Use the loaded default if one is not provided
	if cfg == nil {
		cfg = conf.Get()
	}
	apiConfig := cfg.Metrics

	// Use the json config to look up the metrics object
	lookup, _ := json.Marshal(apiConfig)
	metricsObj, ok := metricsObjects[string(lookup)]
	if ok {
		return metricsObj, nil
	}

	// Spin up the metrics config
	metricsConfig := gmetrics.DefaultConfig(apiConfig["ServiceName"])

	// Host name (allow for automatic detection)
	apiHostName := apiConfig["HostName"]
	if apiHostName != "" && apiHostName != "auto" {
		metricsConfig.HostName = apiHostName
	}

	// Booleans
	if apiConfig["EnableHostname"] == "true" {
		metricsConfig.EnableHostname = true
	} else {
		metricsConfig.EnableHostname = false
	}
	if apiConfig["EnableRuntimeMetrics"] == "true" {
		metricsConfig.EnableRuntimeMetrics = true
	} else {
		metricsConfig.EnableRuntimeMetrics = false
	}
	if apiConfig["EnableTypePrefix"] == "true" {
		metricsConfig.EnableTypePrefix = true
	} else {
		metricsConfig.EnableTypePrefix = false
	}

	// Duaration
	metricsConfig.TimerGranularity = durations[apiConfig["TimerGranularity"]]
	metricsConfig.ProfileInterval = durations[apiConfig["ProfileInterval"]]

	// Make use of some custom configuration for our end
	var sink gmetrics.MetricSink
	if apiConfig["SinkType"] == "Test" {
		sink := gmetrics.NewInmemSink(10*time.Second, 5*time.Minute)
		sig := gmetrics.NewInmemSignal(sink, syscall.SIGQUIT, os.Stdout)
		fmt.Println(sig)
	} else {
		// TODO: set up other sinks
	}

	// Create and store
	metricsObj, _ = gmetrics.New(metricsConfig, sink)
	metricsObjects[string(lookup)] = metricsObj

	return metricsObj, nil
}
