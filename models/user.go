package models

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/mail"

	"code.google.com/p/go-uuid/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/mistifyio/mistify-operator-admin/db"
)

// User is an entity that can interact with mistify
type User struct {
	ID       string            `json:"id"`
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Metadata map[string]string `json:"metadata"`
	Projects []*Project        `json:"-"`
}

// id returns the ID, required by the relatable interface
func (user *User) id() string {
	return user.ID
}

// pkeyName returns the database primary key name, required by the relatable
// interface
func (user *User) pkeyName() string {
	return "user_id"
}

// Validate ensures the user properties are set correctly
func (user *User) Validate() error {
	var results *multierror.Error
	if user.ID == "" {
		results = multierror.Append(results, ErrNoID)
	}
	if uuid.Parse(user.ID) == nil {
		results = multierror.Append(results, ErrBadID)
	}
	if user.Username == "" {
		results = multierror.Append(results, ErrNoUsername)
	}
	if user.Email == "" {
		results = multierror.Append(results, ErrNoEmail)
	}
	if _, err := mail.ParseAddress(user.Email); err != nil {
		results = multierror.Append(results, err)
	}
	if user.Metadata == nil {
		results = multierror.Append(results, ErrNilMetadata)
	}
	return results.ErrorOrNil()
}

// Save persists the user to the database
func (user *User) Save() error {
	if err := user.Validate(); err != nil {
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
	WITH new_values (user_id, username, email, metadata) as (
		VALUES ($1::uuid, $2, $3, $4::json)
	),
	upsert as (
		UPDATE users u SET
			username = nv.username,
			email = nv.email,
			metadata = nv.metadata
		FROM new_values nv
		WHERE u.user_id = nv.user_id
		RETURNING nv.user_id
	)
	INSERT INTO users
		(user_id, username, email, metadata)
	SELECT user_id, username, email, metadata
	FROM new_values nv
	WHERE NOT EXISTS (SELECT 1 FROM upsert u WHERE nv.user_id = u.user_id)
	`
	metadata, err := json.Marshal(user.Metadata)
	if err != nil {
		return err
	}
	_, err = d.Exec(sql,
		user.ID,
		user.Username,
		user.Email,
		string(metadata),
	)
	return err
}

// Delete removes the user from the database
func (user *User) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM users WHERE user_id = $1"
	_, err = d.Exec(sql, user.ID)
	return err
}

// Load retrieves the user from the database
func (user *User) Load() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	SELECT user_id, username, email, metadata
	FROM users
	WHERE user_id = $1
	`
	rows, err := d.Query(sql, user.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()
	if err := user.fromRows(rows); err != nil {
		return nil
	}
	return rows.Err()
}

// fromRows unmarshals a database query result row into a user object
func (user *User) fromRows(rows *sql.Rows) error {
	var metadata string
	err := rows.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&metadata,
	)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(metadata), &user.Metadata)
}

// Decode unmarshals JSON into the user object
func (user *User) Decode(data io.Reader) error {
	if err := json.NewDecoder(data).Decode(user); err != nil {
		return err
	}
	if user.Metadata == nil {
		user.Metadata = make(map[string]string)
	} else {
		for key, value := range user.Metadata {
			if value == "" {
				delete(user.Metadata, key)
			}
		}
	}
	return nil
}

// LoadProjects retrieves the projects associated with the user
func (user *User) LoadProjects() error {
	projects, err := ProjectsByUser(user)
	if err != nil {
		return err
	}
	user.Projects = projects
	return nil
}

// SetProjects creates and ensures the only relations the user has with
// projects
func (user *User) SetProjects(projects []*Project) error {
	if len(projects) == 0 {
		return ClearRelations("projects_users", user)
	}
	relatables := make([]relatable, len(projects))
	for i, project := range projects {
		relatables[i] = relatable(project)
	}
	if err := SetRelations("projects_users", user, relatables); err != nil {
		return err
	}
	return user.LoadProjects()
}

// AddProject adds a relation to a project
func (user *User) AddProject(project *Project) error {
	return AddRelation("projects_users", user, project)
}

// RemoveProject removes a relation with a project
func (user *User) RemoveProject(project *Project) error {
	return RemoveRelation("projects_users", user, project)
}

// NewID generates a new uuid ID
func (user *User) NewID() string {
	user.ID = uuid.New()
	return user.ID
}

// NewUser creates and initializes a new user object
func NewUser() *User {
	user := &User{
		ID:       uuid.New(),
		Metadata: make(map[string]string),
	}
	return user
}

// FetchUser retrieves a user from the database by ID
func FetchUser(id string) (*User, error) {
	user := &User{
		ID: id,
	}
	err := user.Load()
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ListUsers retrieves an array of all users from the database
func ListUsers() ([]*User, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT user_id, username, email, metadata
	FROM users
	ORDER BY user_id asc
	`
	rows, err := d.Query(sql)
	if err != nil {
		return nil, err
	}
	users, err := usersFromRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// UsersByProject retrieves an array of users associated with a project
func UsersByProject(project *Project) ([]*User, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT u.user_id, u.username, u.email, u.metadata
	FROM users u
	JOIN projects_users pu ON u.user_id = pu.user_id
	WHERE pu.project_id = $1
	ORDER BY u.user_id asc
	`
	rows, err := d.Query(sql, project.ID)
	if err != nil {
		return nil, err
	}
	users, err := usersFromRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// usersFromRows unmarshals multiple query rows into an array of users
func usersFromRows(rows *sql.Rows) ([]*User, error) {
	users := make([]*User, 0, 1)
	for rows.Next() {
		user := &User{}
		if err := user.fromRows(rows); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
