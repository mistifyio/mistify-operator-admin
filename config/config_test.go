package config_test

import (
	"io/ioutil"
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

func tempConfigFile() string {
	f, err := ioutil.TempFile("", "testconf")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(f.Name(),
		[]byte(`
		{
			"db": {
				"driver": "postgres",
				"database": "mistify",
				"username": "foobar",
				"password": "baz",
				"host": "localhost",
				"port": 10000
			},
			"mistify":{
				"foo":{
					"bar":"baz"
				}
			}
		}
		`),
		0644,
	)
	return f.Name()
}
