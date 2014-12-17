package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/mistifyio/mistify-agent/log"
)

var conf *Config

type (
	Config struct {
		DB      DB                           `json:"db"`
		Mistify map[string]map[string]string `json:"mistify"`
	}
)

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

func Get() *Config {
	if conf == nil {
		log.Fatal("attempted to access config while config not loaded")
	}
	return conf
}
