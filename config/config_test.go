package config_test

import (
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/config"
)

var configFileName = "../cmd/mistify-operator-admin/testconfig.json"

func TestConfigLoad(t *testing.T) {
	err := config.Load(configFileName)
	h.Ok(t, err)
}

func TestConfigGet(t *testing.T) {
	err := config.Load(configFileName)
	h.Ok(t, err)
	conf := config.Get()
	h.Assert(t, conf != nil, "did not expect conf to be nil")
}
