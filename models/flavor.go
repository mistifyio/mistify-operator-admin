package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"local/mistify-operator-admin/db"

	"code.google.com/p/go-uuid/uuid"
)

type Flavor struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	CPU      int               `json:"cpu"`    // Number of Cores
	Memory   int               `json:"memory"` // Size in MB
	Disk     int               `json:"disk"`   // Size in MB
	Metadata map[string]string `json:"metadata"`
}

func (flavor *Flavor) Validate() error {
	if flavor.ID == "" {
		return errors.New("missing id")
	}
	if uuid.Parse(flavor.ID) == nil {
		return errors.New("invalid id. must be uuid")
	}
	if flavor.Name == "" {
		return errors.New("missing name")
	}
	if flavor.CPU <= 0 {
		return errors.New("cpu must be > 0")
	}
	if flavor.Memory <= 0 {
		return errors.New("memory must be > 0")
	}
	if flavor.Disk <= 0 {
		return errors.New("disk must be > 0")
	}
	if flavor.Metadata == nil {
		return errors.New("metadata must not be nil")
	}
	return nil
}

func (flavor *Flavor) Save() error {
	err := flavor.Validate()
	if err != nil {
		return err
	}

	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	// Writable CTE for an Upsert
	// See: http://stackoverflow.com/a/8702291
	// And: http://dba.stackexchange.com/a/78535
	sql := `
	WITH new_values (flavor_id, name, cpu, memory, disk, metadata) as (
		VALUES ($1::uuid, $2, $3::integer, $4::integer, $5::integer, $6::json)
	),
	upsert as (
		UPDATE flavors f SET
			name = nv.name,
			cpu = nv.cpu,
			memory = nv.memory,
			disk = nv.disk,
			metadata = nv.metadata
		FROM new_values nv
		WHERE f.flavor_id = nv.flavor_id
		RETURNING nv.flavor_id
	)
	INSERT INTO flavors
		(flavor_id, name, cpu, memory, disk, metadata)
	SELECT flavor_id, name, cpu, memory, disk, metadata
	FROM new_values nv
	WHERE NOT EXISTS (SELECT 1 FROM upsert u WHERE nv.flavor_id = u.flavor_id)
	`
	metadata, err := json.Marshal(flavor.Metadata)
	if err != nil {
		return err
	}
	_, err = d.Exec(sql,
		flavor.ID,
		flavor.Name,
		flavor.CPU,
		flavor.Memory,
		flavor.Disk,
		string(metadata),
	)
	return err
}

func (flavor *Flavor) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM flavors WHERE flavor_id = $1"
	_, err = d.Exec(sql, flavor.ID)
	return err
}

func (flavor *Flavor) Load() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	SELECT flavor_id, name, cpu, memory, disk, metadata
	FROM flavors
	WHERE flavor_id = $1
	`
	rows, err := d.Query(sql, flavor.ID)
	if err != nil {
		return err
	}
	rows.Next()
	return flavor.fromRows(rows)
}

func (flavor *Flavor) fromRows(rows *sql.Rows) error {
	var metadata string
	err := rows.Scan(
		&flavor.ID,
		&flavor.Name,
		&flavor.CPU,
		&flavor.Memory,
		&flavor.Disk,
		&metadata,
	)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(metadata), &flavor.Metadata)
}

func (flavor *Flavor) Decode(data io.Reader) error {
	if err := json.NewDecoder(data).Decode(flavor); err != nil {
		return err
	}
	if flavor.Metadata == nil {
		flavor.Metadata = make(map[string]string)
	} else {
		for key, value := range flavor.Metadata {
			if value == "" {
				delete(flavor.Metadata, key)
			}
		}
	}
	return nil
}

func (flavor *Flavor) NewID() string {
	flavor.ID = uuid.New()
	return flavor.ID
}

func NewFlavor() *Flavor {
	flavor := &Flavor{
		ID:       uuid.New(),
		Metadata: make(map[string]string),
	}
	return flavor
}

func FetchFlavor(id string) (*Flavor, error) {
	flavor := &Flavor{
		ID: id,
	}
	if err := flavor.Load(); err != nil {
		return nil, err
	}
	return flavor, nil
}

func ListFlavors() ([]*Flavor, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT flavor_id, name, cpu, memory, disk, metadata
	FROM flavors
	ORDER BY flavor_id asc
	`
	rows, err := d.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flavors := make([]*Flavor, 0, 1)
	for rows.Next() {
		flavor := &Flavor{}
		if err := flavor.fromRows(rows); err != nil {
			return nil, err
		}
		flavors = append(flavors, flavor)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return flavors, nil
}
