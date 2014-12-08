package models

import (
	"database/sql"
	"errors"
	"local/mistify-operator-admin/db"
	"net/mail"

	"code.google.com/p/go-uuid/uuid"
)

type User struct {
	ID       string     `json:"id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
	Projects []*Project `json:"-"`
}

func (user *User) Validate() error {
	if user.ID == "" {
		return errors.New("missing id")
	}
	if uuid.Parse(user.ID) == nil {
		return errors.New("invalid id. must be uuid")
	}
	if user.Username == "" {
		return errors.New("missing username")
	}
	if user.Email == "" {
		return errors.New("missing email")
	}
	if _, err := mail.ParseAddress(user.Email); err != nil {
		return err
	}
	return nil
}

func (user *User) Save() error {
	err := user.Validate()
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
	WITH new_values (user_id, username, email) as (
		VALUES ($1::uuid, $2, $3)
	),
	upsert as (
		UPDATE users u SET
			username = nv.username,
			email = nv.email
		FROM new_values nv
		WHERE u.user_id = nv.user_id
		RETURNING nv.user_id
	)
	INSERT INTO users
		(user_id, username, email)
	SELECT user_id, username, email
	FROM new_values nv
	WHERE NOT EXISTS (SELECT 1 FROM upsert u WHERE nv.user_id = u.user_id)
	`
	_, err = d.Exec(sql,
		user.ID,
		user.Username,
		user.Email,
	)
	return err
}

func (user *User) Apply(update *User) {
	if update.Username != "" {
		user.Username = update.Username
	}
	if _, err := mail.ParseAddress(update.Email); err == nil {
		user.Email = update.Email
	}
}

func (user *User) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM users WHERE user_id = $1"
	_, err = d.Exec(sql, user.ID)
	return err
}

func (user *User) Load() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	SELECT username, email
	FROM users
	WHERE user_id = $1
	`
	err = d.QueryRow(sql, user.ID).Scan(&user.Username, &user.Email)
	return err
}

func (user *User) LoadProjects() error {
	projects, err := ProjectsByUser(user.ID)
	if err != nil {
		return err
	}
	user.Projects = projects
	return nil
}

func (user *User) SetProjects(projectIDs []*string) error {
	err := SetUserProjects(user.ID, projectIDs)
	if err != nil {
		return err
	}
	return user.LoadProjects()
}

func (user *User) AddProject(projectID string) error {
	err := AddProjectUser(projectID, user.ID)
	if err != nil {
		return err
	}
	return user.LoadProjects()
}

func (user *User) RemoveProject(projectID string) error {
	err := RemoveProjectUser(projectID, user.ID)
	if err != nil {
		return err
	}
	return user.LoadProjects()
}

func (user *User) NewID() string {
	user.ID = uuid.New()
	return user.ID
}

func NewUser() *User {
	user := &User{
		ID: uuid.New(),
	}
	return user
}

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

func ListUsers() ([]*User, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT user_id, username, email
	FROM users
	ORDER BY user_id asc
	`
	rows, err := d.Query(sql)
	if err != nil {
		return nil, err
	}
	return usersFromRows(rows)
}

func usersFromRows(rows *sql.Rows) ([]*User, error) {
	defer rows.Close()
	users := make([]*User, 0, 1)
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
