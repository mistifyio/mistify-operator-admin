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

// Mock sink type
// Copied from github.com/armon/go-metrics/sink_test.go
type MockSink struct {
	keys [][]string
	vals []float32
}

func (m *MockSink) SetGauge(key []string, val float32) {
	m.keys = append(m.keys, key)
	m.vals = append(m.vals, val)
}
func (m *MockSink) EmitKey(key []string, val float32) {
	m.keys = append(m.keys, key)
	m.vals = append(m.vals, val)
}
func (m *MockSink) IncrCounter(key []string, val float32) {
	m.keys = append(m.keys, key)
	m.vals = append(m.vals, val)
}
func (m *MockSink) AddSample(key []string, val float32) {
	m.keys = append(m.keys, key)
	m.vals = append(m.vals, val)
}

// Get a new metric object with a mock sink
func newTestMetric() (*gmetrics.Metrics, *MockSink, error) {
	config.Load(configFileName)
	s := &MockSink{}
	m, err := metrics.NewObject(nil, s)
	return m, s, err
}

func TestMetrics_Build(t *testing.T) {
	m, s, err := newTestMetric()
	h.Ok(t, err)
	h.Equals(t, m.Config.ServiceName, "operator-admin")
	h.Equals(t, fmt.Sprintf("%T", *s), "metrics_test.MockSink")
}

func TestMetrics_Counter(t *testing.T) {
	m, s, err := newTestMetric()
	h.Ok(t, err)

	m.IncrCounter([]string{"my_key"}, float32(1))
	h.Equals(t, s.keys[0][0], "operator-admin") // Service Name
	h.Equals(t, s.keys[0][1], "my_key")         // Key
	h.Equals(t, s.vals[0], float32(1))          // Value
}

func TestMetrics_Reuse(t *testing.T) {
	config.Load(configFileName)
	s := &MockSink{}
	m, err := metrics.GetObject(nil, s)
	h.Ok(t, err)

	m2, err := metrics.GetObject(nil, s)
	h.Ok(t, err)
	h.Equals(t, m, m2)
}
