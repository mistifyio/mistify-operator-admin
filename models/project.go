package models

import (
	"database/sql"
	"errors"
	"local/mistify-operator-admin/db"

	"code.google.com/p/go-uuid/uuid"
)

type Project struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Users []*User `json:"-"`
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
	WITH new_values (project_id, name) as (
		VALUES ($1::uuid, $2)
	),
	upsert as (
		UPDATE projects p SET
			name = nv.name
		FROM new_values nv
		WHERE p.project_id = nv.project_id
		RETURNING nv.project_id
	)
	INSERT INTO projects
		(project_id, name)
	SELECT project_id, name
	FROM new_values nv
	WHERE NOT EXISTS (SELECT 1 FROM upsert u WHERE nv.project_id = u.project_id)
	`
	_, err = d.Exec(sql,
		project.ID,
		project.Name,
	)
	return err
}

func (project *Project) Apply(update *Project) {
	if update.Name != "" {
		project.Name = update.Name
	}
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
	SELECT name
	FROM projects
	WHERE project_id = $1
	`
	err = d.QueryRow(sql, project.ID).Scan(
		&project.Name,
	)
	return err
}

func (project *Project) LoadUsers() error {
	users, err := UsersByProject(project.ID)
	if err != nil {
		return err
	}
	project.Users = users
	return nil
}

func (project *Project) SetUsers(userIDs []*string) error {
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
	SELECT project_id, name
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
	defer rows.Close()
	projects := make([]*Project, 0, 1)
	for rows.Next() {
		project := &Project{}
		rows.Scan(
			&project.ID,
			&project.Name,
		)
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}
