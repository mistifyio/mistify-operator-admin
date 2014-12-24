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
	db, err := db.Connect(nil)
	h.Ok(t, err)
	h.Ok(t, db.Ping())
}
