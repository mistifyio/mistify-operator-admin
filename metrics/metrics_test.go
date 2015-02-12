package metrics_test

import (
	"fmt"
	"testing"

	gmetrics "github.com/armon/go-metrics"
	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/config"
	"github.com/mistifyio/mistify-operator-admin/metrics"
)

var configFileName = "../cmd/mistify-operator-admin/testconfig.json"

func TestContext_Build(t *testing.T) {
	conf := &config.Metrics{}
	conf.ServiceName = "test-build"
	mc, err := metrics.NewContext(conf)
	h.Ok(t, err)
	h.Equals(t, mc.Metrics.Config.ServiceName, "test-build")
	h.Equals(t, fmt.Sprintf("%T", *mc.MapSink), "mapsink.MapSink")
	h.Equals(t, mc.StatsdSink, (*gmetrics.StatsdSink)(nil))
}

func TestContext_BuildStatsd(t *testing.T) {
	conf := &config.Metrics{}
	conf.ServiceName = "test-build-statsd"
	conf.StatsdAddress = "example.com:http"
	mc, err := metrics.NewContext(conf)
	h.Ok(t, err)
	h.Equals(t, mc.Metrics.Config.ServiceName, "test-build-statsd")
	h.Equals(t, fmt.Sprintf("%T", *mc.MapSink), "mapsink.MapSink")
	h.Equals(t, fmt.Sprintf("%T", *mc.StatsdSink), "metrics.StatsdSink")
}

func TestContext_BuildConfig(t *testing.T) {
	config.Load(configFileName)
	mc, err := metrics.NewContext(nil)
	h.Ok(t, err)
	h.Equals(t, mc.Metrics.Config.ServiceName, "operator-admin")
	h.Equals(t, fmt.Sprintf("%T", *mc.MapSink), "mapsink.MapSink")
	h.Equals(t, mc.StatsdSink, (*gmetrics.StatsdSink)(nil))
}

func TestMapSink_Counter(t *testing.T) {
	conf := &config.Metrics{}
	conf.ServiceName = "test-counter"
	mc, err := metrics.NewContext(conf)
	h.Ok(t, err)

	mc.Metrics.IncrCounter([]string{"my_key"}, float32(1))
	json, err := mc.MapSink.MarshalJSON()
	h.Equals(t, err, nil)
	h.Equals(t, string(json), "{\"test-counter.my_key\":1}")
}

func TestContext_Reuse(t *testing.T) {
	conf := &config.Metrics{}
	conf.ServiceName = "test-reuse"
	conf.StatsdAddress = "example.com:http"
	mc, err := metrics.GetContext(conf)
	h.Ok(t, err)

	mc2, err := metrics.GetContext(conf)
	h.Ok(t, err)
	h.Equals(t, mc, mc2)
}
