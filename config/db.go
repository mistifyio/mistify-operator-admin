package config

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
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
	var result *multierror.Error
	if _, ok := drivers[db.Driver]; !ok {
		result = multierror.Append(result, ErrDBBadDriver)
	}
	if db.Database == "" {
		result = multierror.Append(result, ErrDBNoDatabase)
	}
	if db.Username == "" {
		result = multierror.Append(result, ErrDBNoUsername)
	}
	if db.Host == "" {
		result = multierror.Append(result, ErrDBNoHost)
	}
	if db.Port <= 0 || db.Port > 65535 {
		result = multierror.Append(result, ErrDBBadPort)
	}
	return result.ErrorOrNil()
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
