package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"local/mistify-operator-admin/db"

	"code.google.com/p/go-uuid/uuid"
)

type Project struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata"`
	Users    []*User           `json:"-"`
}

func (project *Project) Validate() error {
	if project.ID == "" {
		return errors.New("missing id")
	}
	if uuid.Parse(project.ID) == nil {
		return errors.New("invalid id. must be uuid")
	}
	if project.Name == "" {
		return errors.New("missing name")
	}
	if project.Metadata == nil {
		return errors.New("metadata must not be nil")
	}
	return nil
}

func (project *Project) Save() error {
	err := project.Validate()
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

func (project *Project) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM projects WHERE project_id = $1"
	_, err = d.Exec(sql, project.ID)
	return err
}

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

func (project *Project) LoadUsers() error {
	users, err := UsersByProject(project.ID)
	if err != nil {
		return err
	}
	project.Users = users
	return nil
}

func (project *Project) SetUsers(userIDs []string) error {
	err := SetProjectUsers(project.ID, userIDs)
	if err != nil {
		return err
	}
	return project.LoadUsers()
}

func (project *Project) AddUser(userID string) error {
	err := AddProjectUser(project.ID, userID)
	if err != nil {
		return err
	}
	return project.LoadUsers()
}

func (project *Project) RemoveUser(userID string) error {
	err := RemoveProjectUser(project.ID, userID)
	if err != nil {
		return err
	}
	return project.LoadUsers()
}

func (project *Project) NewID() string {
	project.ID = uuid.New()
	return project.ID
}

func NewProject() *Project {
	project := &Project{
		ID: uuid.New(),
	}
	return project
}

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

func projectsFromRows(rows *sql.Rows) ([]*Project, error) {
	projects := make([]*Project, 0, 1)
	for rows.Next() {
		project := &Project{}
		project.fromRows(rows)
		projects = append(projects, project)
	}
	return projects, nil
}
