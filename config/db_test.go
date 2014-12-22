package config_test

import (
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/hashicorp/go-multierror"
	"github.com/mistifyio/mistify-operator-admin/config"
)

func TestDBValidate(t *testing.T) {
	db := &config.DB{}
	var err error

	err = db.Validate()
	h.Assert(t, errContains(config.ErrDBBadDriver, err), "expected 'bad driver' error")
	h.Assert(t, errContains(config.ErrDBNoDatabase, err), "expected 'no database' error")
	h.Assert(t, errContains(config.ErrDBNoUsername, err), "expected 'no username' error")
	h.Assert(t, errContains(config.ErrDBNoHost, err), "expected 'no host' error")
	h.Assert(t, errContains(config.ErrDBBadPort, err), "expected 'bad port' error")

	db.Driver = "foobar"
	err = db.Validate()
	h.Assert(t, errContains(config.ErrDBBadDriver, err), "expected 'bad driver' error")
	db.Driver = "postgres"
	err = db.Validate()
	h.Assert(t, errDoesNotContain(config.ErrDBBadDriver, err), "did not expect 'bad driver' error")

	db.Database = "foobar"
	err = db.Validate()
	h.Assert(t, errDoesNotContain(config.ErrDBNoDatabase, err), "did not expect 'no database' error")
	db.Username = "foobar"
	err = db.Validate()
	h.Assert(t, errDoesNotContain(config.ErrDBNoUsername, err), "did not expect 'no username' error")
	db.Host = "localhost"
	err = db.Validate()
	h.Assert(t, errDoesNotContain(config.ErrDBNoHost, err), "did not expect 'no host' error")

	db.Port = 0
	err = db.Validate()
	h.Assert(t, errContains(config.ErrDBBadPort, err), "expected 'bad port' error")
	db.Port = 70000
	err = db.Validate()
	h.Assert(t, errContains(config.ErrDBBadPort, err), "expected 'bad port' error")
	db.Port = 10000
	err = db.Validate()
	h.Assert(t, errDoesNotContain(config.ErrDBBadPort, err), "expected 'bad port' error")
}

func TestDataSourceName(t *testing.T) {
	db := &config.DB{
		Driver:   "postgres",
		Database: "adb",
		Username: "auser",
		Password: "apassword",
		Host:     "localhost",
		Port:     1337,
	}
	h.Equals(t, db.DataSourceName(), "postgres://auser:apassword@localhost:1337/adb")
}

func errContains(err error, list error) bool {
	merr, ok := list.(*multierror.Error)
	if !ok {
		return false
	}

	errList := merr.Errors
	for _, e := range errList {
		if err == e {
			return true
		}
	}
	return false
}

func errDoesNotContain(err error, list error) bool {
	return !errContains(err, list)
}
