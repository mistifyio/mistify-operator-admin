package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"

	"code.google.com/p/go-uuid/uuid"
	"github.com/mistifyio/mistify-operator-admin/db"
)

type (
	// Permission represents an action that can be taken on an entity by a user
	Permission struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		Service     string            `json:"service"`
		Action      string            `json:"action"`
		EntityType  string            `json:"entityType"`
		Owner       bool              `json:"owner"`
		Description string            `json:"description"`
		Metadata    map[string]string `json:"metadata"`
	}
)

// Validate ensures the permission properties are set correctly
func (permission *Permission) Validate() error {
	if permission.ID == "" {
		return errors.New("missing id")
	}
	if uuid.Parse(permission.ID) == nil {
		return errors.New("invalid id. must be uuid")
	}
	if permission.Service == "" {
		return errors.New("missing service")
	}
	if permission.Action == "" {
		return errors.New("missing action")
	}
	if permission.Metadata == nil {
		return errors.New("metadata must not be nil")
	}
	return nil
}

// Save persists the permission to the database
func (permission *Permission) Save() error {
	if err := permission.Validate(); err != nil {
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
	WITH new_values (permission_id, name, service, action, entitytype, owner,
		description, metadata) AS (
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8::json)
	)
	upsert as (
		UPDATE permissions p SET
			permission_id = nv.permission_id,
			name = nv.name,
			service = nv.service,
			action = nv.action.
			entitytype = nv.entitytype,
			owner = nv.owner,
			meatadata = nv.metadata
		FROM new_values nv
		WHERE p.permission_id = nv.permission_id
		RETURNING nv.permission_id
	)
	INSERT INTO permissions
		(permission_id, name, service, action, entitytype, owner, description,
		metadata)
	SELECT permission_id, name, service, action, entitytype, owner, description,
		metadata
	FROM new_values nv
	WHERE NOT EXISTS
		(SELECT 1 FROM upsert u WHERE nv.permission_id = u.permission_id)
	`
	metadata, err := json.Marshal(permission.Metadata)
	if err != nil {
		return err
	}
	_, err = d.Exec(sql,
		permission.ID,
		permission.Name,
		permission.Service,
		permission.Action,
		permission.EntityType,
		permission.Owner,
		permission.Description,
		string(metadata),
	)
	return err
}

// Delete deletes the permission from the database
func (permission *Permission) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM permissions WHERE permission_id = $1"
	_, err = d.Exec(sql, permission.ID)
	return err
}

// Load retrieves the permission from the database
func (permission *Permission) Load() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	SELECT permission_id, name, service, action, entitytype, owner, description,
		metadata
	FROM permissions
	WHERE permission_id = $1
	`
	rows, err := d.Query(sql, permission.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()
	if err := permission.fromRows(rows); err != nil {
		return err
	}
	return rows.Err()
}

// fromRows unmarshals a database query result row into the permission object
func (permission *Permission) fromRows(rows *sql.Rows) error {
	var metadata string
	err := rows.Scan(
		&permission.ID,
		&permission.Name,
		&permission.Service,
		&permission.Action,
		&permission.EntityType,
		&permission.Owner,
		&permission.Description,
		&metadata,
	)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(metadata), &permission.Metadata)
}

// Decode unmarshals JSON into the permission object
func (permission *Permission) Decode(data io.Reader) error {
	if err := json.NewDecoder(data).Decode(permission); err != nil {
		return err
	}
	if permission.Metadata == nil {
		permission.Metadata = make(map[string]string)
	} else {
		for key, value := range permission.Metadata {
			if value == "" {
				delete(permission.Metadata, key)
			}
		}
	}
	return nil
}

// NewID generates a new uuid ID
func (permission *Permission) NewID() string {
	permission.ID = uuid.New()
	return permission.ID
}

// NewPermission creates and initializes a new permission object
func NewPermission() *Permission {
	permission := &Permission{
		ID:       uuid.New(),
		Metadata: make(map[string]string),
	}
	return permission
}

// ListPermissions retrieve an array of all permission objects from the database
func ListPermissions() ([]*Permission, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT permission_id, name, service, action, entitytype, owner, description,
		metadata
	FROM permissions
	ORDER BY permission_id asc
	`
	rows, err := d.Query(sql)
	if err != nil {
		return nil, err
	}
	permissions := make([]*Permission, 0, 1)
	for rows.Next() {
		permission := &Permission{}
		if err := permission.fromRows(rows); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}
