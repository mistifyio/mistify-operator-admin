package db

import (
	"database/sql"
	"local/mistify-operator-admin/config"

	_ "github.com/lib/pq"
)

// The DB structure handles connection pooling, so keep track of opened ones
var dbConnections map[string]*sql.DB = make(map[string]*sql.DB)

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
	db, err := sql.Open(dbConfig.Driver, dsn)
	if err != nil {
		return nil, err
	}
	dbConnections[dsn] = db
	return db, nil
}
