package models

import (
	"database/sql"
	"encoding/json"
	"io"

	"code.google.com/p/go-uuid/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/mistifyio/mistify-operator-admin/db"
)

// Project describes a set of users and is what is given  ownership of resources
type Project struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Metadata    map[string]string `json:"metadata"`
	Users       []*User           `json:"-"`
	Permissions []*Permission     `json:"-"`
}

// id returns the id, required by the relatable interface
func (project *Project) id() string {
	return project.ID
}

// pkeyName returns the database primary key name, required by the relatable
// interface
func (project *Project) pkeyName() string {
	return "project_id"
}

// Validate ensures the project properties are set correctly
func (project *Project) Validate() error {
	var results *multierror.Error
	if project.ID == "" {
		results = multierror.Append(results, ErrNoID)
	}
	if uuid.Parse(project.ID) == nil {
		results = multierror.Append(results, ErrBadID)
	}
	if project.Name == "" {
		results = multierror.Append(results, ErrNoName)
	}
	if project.Metadata == nil {
		results = multierror.Append(results, ErrNilMetadata)
	}
	return results.ErrorOrNil()
}

// Save persists a project to the database
func (project *Project) Save() error {
	if err := project.Validate(); err != nil {
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
	WITH new_values (project_id, name, metadata) as (
		VALUES ($1::uuid, $2, $3::json)
	),
	upsert as (
		UPDATE projects p SET
			name = nv.name,
			metadata = nv.metadata
		FROM new_values nv
		WHERE p.project_id = nv.project_id
		RETURNING nv.project_id
	)
	INSERT INTO projects
		(project_id, name, metadata)
	SELECT project_id, name, metadata
	FROM new_values nv
	WHERE NOT EXISTS (SELECT 1 FROM upsert u WHERE nv.project_id = u.project_id)
	`
	metadata, err := json.Marshal(project.Metadata)
	if err != nil {
		return err
	}
	_, err = d.Exec(sql,
		project.ID,
		project.Name,
		string(metadata),
	)
	return err
}

// Delete removes a project from the database
func (project *Project) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM projects WHERE project_id = $1"
	_, err = d.Exec(sql, project.ID)
	return err
}

// Load retrieves a project from the database
func (project *Project) Load() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	SELECT project_id, name, metadata
	FROM projects
	WHERE project_id = $1
	`
	rows, err := d.Query(sql, project.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()
	if err := project.fromRows(rows); err != nil {
		return err
	}
	return rows.Err()
}

// fromRows unmarshals a database query result row into the project object
func (project *Project) fromRows(rows *sql.Rows) error {
	var metadata string
	err := rows.Scan(
		&project.ID,
		&project.Name,
		&metadata,
	)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(metadata), &project.Metadata)
}

// Decode unmarshals JSON into the project object
func (project *Project) Decode(data io.Reader) error {
	if err := json.NewDecoder(data).Decode(project); err != nil {
		return err
	}
	if project.Metadata == nil {
		project.Metadata = make(map[string]string)
	} else {
		for key, value := range project.Metadata {
			if value == "" {
				delete(project.Metadata, key)
			}
		}
	}
	return nil
}

// LoadUsers retrieves the users related to the project from the database
func (project *Project) LoadUsers() error {
	users, err := UsersByProject(project)
	if err != nil {
		return err
	}
	project.Users = users
	return nil
}

// SetUsers creates and ensures the only relations teh project has with users
func (project *Project) SetUsers(users []*User) error {
	if len(users) == 0 {
		return ClearRelations("projects_users", project)
	}
	relatables := make([]relatable, len(users))
	for i, user := range users {
		relatables[i] = relatable(user)
	}
	if err := SetRelations("projects_users", project, relatables); err != nil {
		return err
	}
	return project.LoadUsers()
}

// AddUser adds a relation to a user
func (project *Project) AddUser(user *User) error {
	return AddRelation("projects_users", project, user)
}

// RemoveUser removes a relation with a user
func (project *Project) RemoveUser(user *User) error {
	return RemoveRelation("projects_users", project, user)
}

// LoadPermissions retrieves the permissions related to the project from the database
func (project *Project) LoadPermissions() error {
	permissions, err := PermissionsByProject(project)
	if err != nil {
		return err
	}
	project.Permissions = permissions
	return nil
}

// SetPermissions creates and ensures the only relations teh project has with permissions
func (project *Project) SetPermissions(permissions []*Permission) error {
	if len(permissions) == 0 {
		return ClearRelations("projects_permissions", project)
	}
	relatables := make([]relatable, len(permissions))
	for i, permission := range permissions {
		relatables[i] = relatable(permission)
	}
	if err := SetRelations("projects_permissions", project, relatables); err != nil {
		return err
	}
	return project.LoadPermissions()
}

// AddPermission adds a relation to a permission
func (project *Project) AddPermission(permission *Permission) error {
	return AddRelation("projects_permissions", project, permission)
}

// RemovePermission removes a relation with a permission
func (project *Project) RemovePermission(permission *Permission) error {
	return RemoveRelation("projects_permissions", project, permission)
}

// NewID generates a new uuid ID
func (project *Project) NewID() string {
	project.ID = uuid.New()
	return project.ID
}

// NewProject creates and initializes a new project object
func NewProject() *Project {
	project := &Project{
		ID:       uuid.New(),
		Metadata: make(map[string]string),
	}
	return project
}

// FetchProject retrieves a project object from the database by ID
func FetchProject(id string) (*Project, error) {
	project := &Project{
		ID: id,
	}
	err := project.Load()
	if err != nil {
		return nil, err
	}
	return project, nil
}

// ListProjects retrieves an array of all projects from the database
func ListProjects() ([]*Project, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT project_id, name, metadata
	FROM projects
	ORDER BY project_id asc
	`
	rows, err := d.Query(sql)
	if err != nil {
		return nil, err
	}
	return projectsFromRows(rows)
}

// ProjectsByUser retrieves an array of projects related to a user
func ProjectsByUser(user *User) ([]*Project, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT p.project_id, p.name, p.metadata
	FROM projects p
	JOIN projects_users pu ON p.project_id = pu.project_id
	WHERE pu.user_id = $1
	ORDER BY project_id asc
	`
	rows, err := d.Query(sql, user.ID)
	if err != nil {
		return nil, err
	}
	projects, err := projectsFromRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}

// ProjectsByPermission retrieves an array of projects related to a permission
func ProjectsByPermission(permission *Permission) ([]*Project, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT p.project_id, p.name, p.metadata
	FROM projects p
	JOIN projects_permissions pp ON p.project_id = pp.project_id
	WHERE pp.permission_id = $1
	ORDER BY project_id asc
	`
	rows, err := d.Query(sql, permission.ID)
	if err != nil {
		return nil, err
	}
	projects, err := projectsFromRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}

// projectsFromRows unmarshals multiple query rows into an array of projects
func projectsFromRows(rows *sql.Rows) ([]*Project, error) {
	projects := make([]*Project, 0, 1)
	for rows.Next() {
		project := &Project{}
		project.fromRows(rows)
		projects = append(projects, project)
	}
	return projects, nil
}
