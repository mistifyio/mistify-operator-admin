package config

import "errors"

// ErrDBBadDriver is for bad or missing database driver in the config
var ErrDBBadDriver = errors.New("missing or invalid database driver")

// ErrDBNoDatabase is for missing a database in the config
var ErrDBNoDatabase = errors.New("missing database")

// ErrDBNoUsername is for a missing username in the config
var ErrDBNoUsername = errors.New("missing database username")

// ErrDBNoHost is for a missing database host in the config
var ErrDBNoHost = errors.New("missing database host")

// ErrDBBadPort is for a bad or missing database port in the config
var ErrDBBadPort = errors.New("missing or invalid database port")

// ErrMetricsNoServiceName is for a missing service name in the config
var ErrMetricsNoServiceName = errors.New("missing service name")

// ErrMetricsBadStatsdAddress is for a bad statsd address in the config
var ErrMetricsBadStatsdAddress = errors.New("invalid address for statsd")
