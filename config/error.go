package config

import "errors"

var ErrDBBadDriver = errors.New("missing or invalid database driver")
var ErrDBNoDatabase = errors.New("missing database")
var ErrDBNoUsername = errors.New("missing database username")
var ErrDBNoHost = errors.New("missing database host")
var ErrDBBadPort = errors.New("missing or invalid database port")

var ErrMetricsBadSinkType = errors.New("missing or invalid sink type")
var ErrMetricsNoServiceName = errors.New("missing service name")
var ErrMetricsBadTimerGranularity = errors.New("invalid duration for timer granularity")
var ErrMetricsBadProfileInterval = errors.New("invalid duration for profile interval")
