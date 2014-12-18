// Package config allows for reading configuration from a JSON file
package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/mistifyio/mistify-agent/log"
)

var conf *Config

type (
	// Config struct holds data from a JSON config file
	Config struct {
		DB      DB                           `json:"db"`
		Mistify map[string]map[string]string `json:"mistify"`
	}
)

// Load parses a JSON config file
func Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	newConfig := &Config{}

	if err := json.Unmarshal(data, newConfig); err != nil {
		return err
	}

	if err := newConfig.DB.Validate(); err != nil {
		return err
	}

	conf = newConfig

	return nil
}

// Get returns the configuration data and dies if the config is not loaded
func Get() *Config {
	if conf == nil {
		log.Fatal("attempted to access config while config not loaded")
	}
	return conf
}
