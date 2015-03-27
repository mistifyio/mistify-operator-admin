package db

import (
	"database/sql"
	"sync"

	_ "github.com/lib/pq" // doesn't need to be used directly
	"github.com/mistifyio/mistify-operator-admin/config"
)

// The DB structure handles connection pooling, so keep track of opened ones
var dbConnections = make(map[string]*sql.DB)
var mutex sync.Mutex

// Connect creates or returns an existing database connection based on the data
// source name (dsn)
func Connect(dbConfig *config.DB) (*sql.DB, error) {
	// Use the loaded default if one is not provided
	if dbConfig == nil {
		conf := config.Get()
		dbConfig = &conf.DB
	}

	dsn := dbConfig.DataSourceName()

	// Reuse existing open DB
	db, ok := dbConnections[dsn]
	if ok {
		return db, nil
	}

	// Open a new DB and keep track of it
	mutex.Lock()
	defer mutex.Unlock()
	db, err := sql.Open(dbConfig.Driver, dsn)
	if err != nil {
		return nil, err
	}
	dbConnections[dsn] = db
	return db, nil
}
