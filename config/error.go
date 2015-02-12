package config

import "errors"

var ErrDBBadDriver = errors.New("missing or invalid database driver")
var ErrDBNoDatabase = errors.New("missing database")
var ErrDBNoUsername = errors.New("missing database username")
var ErrDBNoHost = errors.New("missing database host")
var ErrDBBadPort = errors.New("missing or invalid database port")

var ErrMetricsNoServiceName = errors.New("missing service name")
var ErrMetricsBadStatsdAddress = errors.New("invalid address for statsd")
