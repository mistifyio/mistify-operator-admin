package config

import (
	"github.com/hashicorp/go-multierror"
)

// MetricKeys is the JSON structure and validation for configuring how keys are made from urls in the middleware
type MetricKeys struct {
	AllowedChunks []string `json:"allowed_chunks"`
}

// Validate ensures that the metrics configuration is reasonable
func (self *MetricKeys) Validate() error {
	var result *multierror.Error
	return result.ErrorOrNil()
}
