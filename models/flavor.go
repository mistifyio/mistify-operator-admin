package models

import (
	"errors"
	"local/mistify-operator-admin/db"

	"code.google.com/p/go-uuid/uuid"
)

type Flavor struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	CPU    int    `json:"cpu"`    // Number of Cores
	Memory int    `json:"memory"` // Size in MB
	Disk   int    `json:"disk"`   // Size in MB
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
	WITH new_values (flavor_id, name, cpu, memory, disk) as (
		VALUES ($1::uuid, $2, $3::integer, $4::integer, $5::integer)
	),
	upsert as (
		UPDATE flavors f SET
			name = nv.name,
			cpu = nv.cpu,
			memory = nv.memory,
			disk = nv.disk
		FROM new_values nv
		WHERE f.flavor_id = nv.flavor_id
		RETURNING nv.flavor_id
	)
	INSERT INTO flavors
		(flavor_id, name, cpu, memory, disk)
	SELECT flavor_id, name, cpu, memory, disk
	FROM new_values nv
	WHERE NOT EXISTS (SELECT 1 FROM upsert u WHERE nv.flavor_id = u.flavor_id)
	`
	_, err = d.Exec(sql,
		flavor.ID,
		flavor.Name,
		flavor.CPU,
		flavor.Memory,
		flavor.Disk,
	)
	return err
}

func (flavor *Flavor) Apply(update *Flavor) {
	if update.Name != "" {
		flavor.Name = update.Name
	}
	if update.CPU != 0 {
		flavor.CPU = update.CPU
	}
	if update.Memory != 0 {
		flavor.Memory = update.Memory
	}
	if update.Disk != 0 {
		flavor.Disk = update.Disk
	}
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
	SELECT name, cpu, memory, disk
	FROM flavors
	WHERE flavor_id = $1
	`
	err = d.QueryRow(sql, flavor.ID).Scan(
		&flavor.Name,
		&flavor.CPU,
		&flavor.Memory,
		&flavor.Disk,
	)
	return err
}

func (flavor *Flavor) NewID() string {
	flavor.ID = uuid.New()
	return flavor.ID
}

func NewFlavor() *Flavor {
	flavor := &Flavor{
		ID: uuid.New(),
	}
	return flavor
}

func FetchFlavor(id string) (*Flavor, error) {
	flavor := &Flavor{
		ID: id,
	}
	err := flavor.Load()
	if err != nil {
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
	SELECT flavor_id, name, cpu, memory, disk
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
		rows.Scan(
			&flavor.ID,
			&flavor.Name,
			&flavor.CPU,
			&flavor.Memory,
			&flavor.Disk,
		)
		flavors = append(flavors, flavor)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return flavors, nil
}
