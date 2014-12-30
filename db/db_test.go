package db_test

import (
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/config"
	"github.com/mistifyio/mistify-operator-admin/db"
)

var configFileName = "../cmd/mistify-operator-admin/testconfig.json"

func TestConnect(t *testing.T) {
	config.Load(configFileName)
	d, err := db.Connect(nil)
	h.Ok(t, err)
	h.Ok(t, d.Ping())

	// Reuse existing
	d2, err := db.Connect(nil)
	h.Ok(t, err)
	h.Equals(t, d, d2)
}
