package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	conf "github.com/mistifyio/mistify-operator-admin/config"
	"github.com/mistifyio/mistify-operator-admin/db"
)

type (
	// Config persists and retrieves arbitrary namespaced key/value pairs with
	// support for default values.
	Config struct {
		// Since set values overlay defaults, prevent direct access to the data
		// and force use of appropriate getters/setters
		data map[string]map[string]string
	}
)

// Validate checks that the properties of Config are valid
func (config *Config) Validate() error {
	if config.data == nil {
		return errors.New("data must not be nil")
	}
	return nil
}

// Get retrieves the full config with set values taking precedence over defaults
func (config *Config) Get() map[string]map[string]string {
	data := make(map[string]map[string]string)
	defaultConfig := conf.Get().Mistify
	for namespace := range defaultConfig {
		data[namespace] = config.GetNamespace(namespace)
	}
	for namespace := range config.data {
		if _, ok := data[namespace]; !ok {
			data[namespace] = config.GetNamespace(namespace)
		}
	}
	return data
}

// GetNamespace returns a map of config key/value pairs with set values merged
// on top of defaults. It is a new map, so modifications will not be stored
func (config *Config) GetNamespace(namespace string) map[string]string {
	ns := make(map[string]string)
	// Start with defaults
	defaultNS, ok := conf.Get().Mistify[namespace]
	if ok {
		for i, v := range defaultNS {
			ns[i] = v
		}
	}
	// Overlay configured values
	dataNS, ok := config.data[namespace]
	if ok {
		for i, v := range dataNS {
			ns[i] = v
		}
	}
	return ns
}

// SetNamespace places a set of key/value pairs under a given namespace
func (config *Config) SetNamespace(namespace string, value map[string]string) {
	if value == nil {
		config.DeleteNamespace(namespace)
	} else {
		config.data[namespace] = value
	}
}

// DeleteNamespace deletes a set namespace. Defaults are not deleted.
func (config *Config) DeleteNamespace(namespace string) {
	delete(config.data, namespace)
}

// GetValue retrieves the value of a namespaced key, with set values taking
// precidence over defaults
func (config *Config) GetValue(namespace string, key string) (string, bool) {
	// Go compiler isn't smart enough to use the two-value in a direct return
	value, ok := config.GetNamespace(namespace)[key]
	return value, ok
}

// SetValue sets the value of a namespaced key, creating the namespace if it
// does not already exist.
func (config *Config) SetValue(namespace string, key string, value string) {
	ns, ok := config.data[namespace]
	if !ok {
		config.data[namespace] = make(map[string]string)
		ns = config.data[namespace]
	}
	ns[key] = value
}

// DeleteValue deletes a set namespaced key. Defaults are not deleted.
func (config *Config) DeleteValue(namespace string, key string) {
	ns, ok := config.data[namespace]
	if !ok {
		return
	}
	delete(ns, key)
}

// Merge merges in set values from another Config into the current one
func (config *Config) Merge(updates *Config) {
	for ns, data := range updates.data {
		for key, value := range data {
			config.SetValue(ns, key, value)
		}
	}
}

// Clean will remove any set values that match defaults and any empty set
// namespaces. Primarilly used before persisting to reduce redundant and
// unnecessary data.
func (config *Config) Clean() {
	for name, ns := range config.data {
		// Discard anything set the same as default values
		for key, v := range ns {
			if v == conf.Get().Mistify[name][key] {
				delete(ns, key)
			}
		}
		// Discard completely empty namespaces
		if len(ns) == 0 {
			delete(config.data, name)
		}
	}
}

// Load retrieves the persisted set data
func (config *Config) Load() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}

	sql := `
	SELECT namespace, data
	FROM config
	`
	rows, err := d.Query(sql)
	if err != nil {
		return err
	}

	for rows.Next() {
		var namespace, dataJSON string
		rows.Scan(&namespace, &dataJSON)
		var data map[string]string
		if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
			return err
		}
		config.data[namespace] = data
	}

	return rows.Err()
}

// Decode decodes JSON into a Config
func (config *Config) Decode(data io.Reader) error {
	if err := json.NewDecoder(data).Decode(&config.data); err != nil {
		return err
	}
	if config.data == nil {
		config.data = make(map[string]map[string]string)
	}
	return nil
}

// Save persists the set data
func (config *Config) Save() error {
	if err := config.Validate(); err != nil {
		return err
	}

	config.Clean()

	d, err := db.Connect(nil)
	if err != nil {
		return err
	}

	// Clearing and rebuilding the table
	// If we need accurate auditing, switch from a txn to a writable CTE
	// that handles updates/inserts/deletes more granularly
	txn, err := d.Begin()
	if err != nil {
		return err
	}

	// Clear the table
	if _, err := txn.Exec("TRUNCATE config"); err != nil {
		txn.Rollback()
		return err
	}

	// Commit if we don't have anything to save
	if len(config.data) == 0 {
		return txn.Commit()
	}

	// Build the variable length values sql and vars array
	placeholders := make([]string, len(config.data))
	values := make([]interface{}, len(config.data)*2)
	i := 0
	for namespace, data := range config.data {
		placeholders[i] = fmt.Sprintf("($%d, $%d::json)", (i*2)+1, (i*2)+2)
		dataJSON, err := json.Marshal(data)
		if err != nil {
			txn.Rollback()
			return err
		}
		values[2*i] = interface{}(namespace)
		values[(2*i)+1] = interface{}(string(dataJSON))
		i++
	}
	sql := `
	INSERT INTO config
		(namespace, data)
	VALUES
	`
	sql += strings.Join(placeholders, ",")
	if _, err = txn.Exec(sql, values...); err != nil {
		txn.Rollback()
		return err
	}
	return txn.Commit()
}

// NewConfig creates a new Config instance and initializes the internal data map
func NewConfig() *Config {
	return &Config{
		data: make(map[string]map[string]string),
	}
}
