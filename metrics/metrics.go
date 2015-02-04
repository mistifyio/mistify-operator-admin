package metrics

import (
	"encoding/json"
	"os"
	"sync"
	"syscall"
	"time"

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

	// Spin up the metrics config
	metricsConfig := apiConfig.MetricsObjectConfig()

	// Set up the sink
	var sink gmetrics.MetricSink
	if apiConfig.SinkType == "Test" {
		sink = gmetrics.NewInmemSink(10*time.Second, 5*time.Minute)
		gmetrics.NewInmemSignal(sink.(*gmetrics.InmemSink), syscall.SIGQUIT, os.Stdout)
	} else {
		// TODO: set up other sinks
	}

	// Create and store
	metricsObj, err = gmetrics.New(metricsConfig, sink)
	if err != nil {
		return nil, err
	}
	metricsObjects[string(lookup)] = metricsObj

	return metricsObj, nil
}
