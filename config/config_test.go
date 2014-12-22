package config_test

import (
	"io/ioutil"
	"syscall"
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/config"
)

func TestConfigLoad(t *testing.T) {
	fileName := loadConfig(t)
	defer syscall.Unlink(fileName)
}

func TestConfigGet(t *testing.T) {
	fileName := loadConfig(t)
	defer syscall.Unlink(fileName)
	conf := config.Get()
	h.Assert(t, conf != nil, "did not expect conf to be nil")
}

func loadConfig(t *testing.T) string {
	fileName := tempConfigFile()
	err := config.Load(fileName)
	h.Ok(t, err)
	return fileName
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
