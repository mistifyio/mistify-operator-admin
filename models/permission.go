package models

import (
	"database/sql"
	"encoding/json"
	"io"

	"code.google.com/p/go-uuid/uuid"
	"github.com/hashicorp/go-multierror"
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
		Projects    []*Project        `json:"-"`
	}
)

// id returns the ID, required by the relatable interface
func (permission *Permission) id() string {
	return permission.ID
}

// pkeyName returns the database primary key name, required by the relatable
// interface
func (permission *Permission) pkeyName() string {
	return "permission_id"
}

// Validate ensures the permission properties are set correctly
func (permission *Permission) Validate() error {
	var results *multierror.Error
	if permission.ID == "" {
		results = multierror.Append(results, ErrNoID)
	}
	if uuid.Parse(permission.ID) == nil {
		results = multierror.Append(results, ErrBadID)
	}
	if permission.Service == "" {
		results = multierror.Append(results, ErrNoService)
	}
	if permission.Action == "" {
		results = multierror.Append(results, ErrNoAction)
	}
	if permission.Metadata == nil {
		results = multierror.Append(results, ErrNilMetadata)
	}
	return results.ErrorOrNil()
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
		VALUES ($1::uuid, $2, $3, $4, $5, $6::boolean, $7, $8::json)
	),
	upsert as (
		UPDATE permissions p SET
			permission_id = nv.permission_id,
			name = nv.name,
			service = nv.service,
			action = nv.action,
			entitytype = nv.entitytype,
			owner = nv.owner,
			metadata = nv.metadata
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

// LoadProjects retrieves the projects associated with the permission
func (permission *Permission) LoadProjects() error {
	projects, err := ProjectsByPermission(permission)
	if err != nil {
		return err
	}
	permission.Projects = projects
	return nil
}

// SetProjects creates and ensures the only relations the permission has with
// projects
func (permission *Permission) SetProjects(projects []*Project) error {
	if len(projects) == 0 {
		return ClearRelations("projects_permissions", permission)
	}
	relatables := make([]relatable, len(projects))
	for i, project := range projects {
		relatables[i] = relatable(project)
	}
	if err := SetRelations("projects_permissions", permission, relatables); err != nil {
		return err
	}
	return permission.LoadProjects()
}

// AddProject adds a relation to a project
func (permission *Permission) AddProject(project *Project) error {
	return AddRelation("projects_permissions", permission, project)
}

// RemoveProject removes a relation with a project
func (permission *Permission) RemoveProject(project *Project) error {
	return RemoveRelation("projects_permissions", permission, project)
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

// FetchPermission retrieves a permission object from the database by ID
func FetchPermission(id string) (*Permission, error) {
	permission := &Permission{
		ID: id,
	}
	err := permission.Load()
	if err != nil {
		return nil, err
	}
	return permission, nil
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
	permissions, err := permissionsFromRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

// PermissionsByProject retrieves an array of permission related to a project
func PermissionsByProject(project *Project) ([]*Permission, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT p.permission_id, p.name, p.service, p.action, p.entitytype, p.owner,
		p.description, p.metadata
	FROM permissions p
    JOIN projects_permissions pp ON p.permission_id = pp.permission_id
    WHERE pp.project_id = $1
    ORDER BY permission_id asc
    `
	rows, err := d.Query(sql, project.ID)
	if err != nil {
		return nil, err
	}
	permissions, err := permissionsFromRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

// permissionsFromRows unmarshals multiple query rows into an array of permissions
func permissionsFromRows(rows *sql.Rows) ([]*Permission, error) {
	permissions := make([]*Permission, 0, 1)
	for rows.Next() {
		permission := &Permission{}
		permission.fromRows(rows)
		permissions = append(permissions, permission)
	}
	return permissions, nil
}
