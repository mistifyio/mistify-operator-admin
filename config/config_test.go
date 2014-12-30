package config_test

import (
	"io/ioutil"
	"syscall"
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/config"
)

var configFileName = "../cmd/mistify-operator-admin/testconfig.json"

func TestConfigLoad(t *testing.T) {
	h.Ok(t, config.Load(configFileName))
	h.Assert(t, config.Load("thisisnotarealfile") != nil, "expected ReadFile error")

	f, err := ioutil.TempFile("", "BadJSON")
	if err != nil {
		panic(err)
	}
	defer syscall.Unlink(f.Name())
	ioutil.WriteFile(f.Name(), []byte("foobar"), 0644)
	h.Assert(t, config.Load(f.Name()) != nil, "expected Unmarshal error")

	f, err = ioutil.TempFile("", "BadJSON")
	if err != nil {
		panic(err)
	}
	defer syscall.Unlink(f.Name())
	ioutil.WriteFile(f.Name(), []byte(`{"db":{}}`), 0644)

	h.Assert(t, config.Load(f.Name()) != nil, "expected DB Validate error")
}

func TestConfigGet(t *testing.T) {
	err := config.Load(configFileName)
	h.Ok(t, err)
	conf := config.Get()
	h.Assert(t, conf != nil, "did not expect conf to be nil")
}
