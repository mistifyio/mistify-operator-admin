package metrics

import (
	"sync"

	gmetrics "github.com/armon/go-metrics"
	"github.com/bakins/go-metrics-map"
	"github.com/bakins/go-metrics-middleware"
	"github.com/mistifyio/mistify-operator-admin/config"
)

// MetricsContext contains information necessary to add time and count metrics
// to routes, show collected metrics, or emit custom metrics from within a route
type MetricsContext struct {
	Metrics    *gmetrics.Metrics
	Middleware *mmw.Middleware
	MapSink    *mapsink.MapSink
	StatsdSink *gmetrics.StatsdSink
}

// If different statsd addresses are requested, they should result in entirely
// different contexts
var contexts map[string]*MetricsContext = make(map[string]*MetricsContext)
var mutex sync.Mutex

// Get the metrics context for a statsd address, or create one
func GetContext(apiConfig *config.Metrics) (*MetricsContext, error) {
	// Use the loaded default if one is not provided
	if apiConfig == nil {
		conf := config.Get()
		apiConfig = &conf.Metrics
	}

	// Look up the address and return if necessary
	var key = apiConfig.ServiceName + apiConfig.StatsdAddress
	var err error
	mc, ok := contexts[key]
	if ok {
		return mc, nil
	}

	// Build a new context, store it, and return
	mutex.Lock()
	defer mutex.Unlock()
	mc, err = NewContext(apiConfig)
	if err != nil {
		return nil, err
	}
	contexts[key] = mc
	return mc, nil
}

// Get a new metrics context given a statsd address
func NewContext(apiConfig *config.Metrics) (*MetricsContext, error) {
	// Use the loaded default if one is not provided
	if apiConfig == nil {
		conf := config.Get()
		apiConfig = &conf.Metrics
	}

	// Build the fanout sink, with statsd if required
	var mainSink *gmetrics.FanoutSink
	var statsdSink *gmetrics.StatsdSink
	var err error
	mapSink := mapsink.New()
	if apiConfig.StatsdAddress != "" {
		statsdSink, err = gmetrics.NewStatsdSink(apiConfig.StatsdAddress)
		if err != nil {
			return nil, err
		}
		mainSink = &gmetrics.FanoutSink{statsdSink, mapSink}
	} else {
		mainSink = &gmetrics.FanoutSink{mapSink}
	}

	// Build the metrics object
	cfg := gmetrics.DefaultConfig(apiConfig.ServiceName)
	cfg.EnableHostname = false
	metrics, err := gmetrics.New(cfg, mainSink)
	if err != nil {
		return nil, err
	}

	// Create the middleware and return everything
	mmw := mmw.New(metrics)
	return &MetricsContext{metrics, mmw, mapSink, statsdSink}, nil
}
