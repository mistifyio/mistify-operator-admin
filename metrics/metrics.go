package metrics

import (
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

// One metrics context only
var context *MetricsContext
var contextLoaded bool = false

// Load the metrics context
func LoadContext() error {
	conf := config.Get()
	apiConfig := &conf.Metrics
	mc, err := NewContext(apiConfig)
	if err != nil {
		return err
	}
	context = mc
	contextLoaded = true
	return nil
}

// Get the metrics context
func GetContext() *MetricsContext {
	if !contextLoaded {
		LoadContext()
	}
	return context
}

// Get a new metrics context given a statsd address and service name
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
