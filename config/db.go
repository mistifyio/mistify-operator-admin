package config

import (
	"errors"
	"fmt"
)

// Available drivers
var drivers = map[string]bool{
	"postgres": true,
}

// DB is the JSON structure and validation for database configuration
type DB struct {
	Driver   string `json:"driver"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     uint   `json:"port"`
}

// Validate ensures that the database configuration is reasonable
func (db *DB) Validate() error {
	if _, ok := drivers[db.Driver]; !ok {
		return fmt.Errorf("'%s': not an available database driver", db.Driver)
	}
	if db.Database == "" {
		return errors.New("database cannot be empty")
	}
	if db.Username == "" {
		return errors.New("username cannot be empty")
	}
	if db.Host == "" {
		return errors.New("host cannot be empty")
	}
	if db.Port <= 0 || db.Port > 65535 {
		return fmt.Errorf("%d: not a valid port", db.Port)
	}
	return nil
}

// DataSourceName generates the dsn for connecting to the database from the
// configured values
func (db *DB) DataSourceName() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s",
		db.Driver,
		db.Username,
		db.Password,
		db.Host,
		db.Port,
		db.Database,
	)
}
