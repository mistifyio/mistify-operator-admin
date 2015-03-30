package metrics

import (
	gmetrics "github.com/armon/go-metrics"
	"github.com/bakins/go-metrics-map"
	"github.com/bakins/go-metrics-middleware"
	"github.com/mistifyio/mistify-operator-admin/config"
)

// Context contains information necessary to add time and count metrics
// to routes, show collected metrics, or emit custom metrics from within a route
type Context struct {
	Metrics    *gmetrics.Metrics
	Middleware *mmw.Middleware
	MapSink    *mapsink.MapSink
	StatsdSink *gmetrics.StatsdSink
}

// One context only
var context *Context
var contextLoaded = false

// LoadContext loads the context from config
func LoadContext() error {
	conf := config.Get()
	apiConfig := &conf.Metrics
	c, err := NewContext(apiConfig)
	if err != nil {
		return err
	}
	context = c
	contextLoaded = true
	return nil
}

// GetContext retrieves the context, loading if it has not yet
func GetContext() *Context {
	if !contextLoaded {
		_ = LoadContext()
	}
	return context
}

// NewContext creates a new context from the config
func NewContext(apiConfig *config.Metrics) (*Context, error) {
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
	return &Context{metrics, mmw, mapSink, statsdSink}, nil
}
