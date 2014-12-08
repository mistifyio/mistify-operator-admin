package models

import (
	"fmt"

	"local/mistify-operator-admin/db"
)

func UsersByProject(projectID string) ([]*User, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT u.user_id, u.username, u.email
	FROM users u
	JOIN projects_users pu ON u.user_id = pu.user_id
	WHERE pu.project_id = $1
	ORDER BY u.user_id asc
	`
	rows, err := d.Query(sql, projectID)
	if err != nil {
		return nil, err
	}
	return usersFromRows(rows)
}

func ProjectsByUser(userID string) ([]*Project, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT p.project_id, p.name
	FROM projects p
	JOIN projects_users pu ON p.project_id = pu.project_id
	WHERE pu.user_id = $1
	ORDER BY project_id asc
	`
	rows, err := d.Query(sql, userID)
	if err != nil {
		return nil, err
	}
	return projectsFromRows(rows)
}

func AddProjectUser(projectID string, userID string) error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	INSERT INTO projects_users (project_id, user_id)
	SELECT $1, $2
	WHERE NOT EXISTS (SELECT 1 FROM projects_users WHERE project_id = $1 AND user_id = $2)
	`
	_, err = d.Exec(sql, projectID, userID)
	return err
}

func RemoveProjectUser(projectID string, userID string) error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	DELETE FROM projects_users
	WHERE project_id = $1 AND user_id = $2
	`
	_, err = d.Exec(sql, projectID, userID)
	return err
}

func SetProjectUsers(projectID string, userIDs []*string) error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}

	sql := `
	WITH deletes AS (
		DELETE FROM projects_users
		WHERE project_id = $1
	),
	new_values (project_id, user_id) AS (
	`
	if len(userIDs) > 0 {
		sql += " VALUES "
		for i := range userIDs {
			sql += fmt.Sprintf("($1::uuid, $%d::uuid)", i+2)
		}
	}
	sql += `
	)
	INSERT INTO projects_users (project_id, user_id)
	SELECT project_id, user_id
	FROM new_values
	`
	// Variadic can be frustrating. http://stackoverflow.com/a/12990540
	vars := make([]interface{}, len(userIDs)+1)
	vars[0] = projectID
	for i, v := range userIDs {
		vars[i+1] = interface{}(v)
	}
	_, err = d.Query(sql, vars...)
	return err
}

func SetUserProjects(userID string, projectIDs []*string) error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}

	sql := `
	WITH deletes AS (
		DELETE FROM projects_users
		WHERE user_id = $1
	),
	new_values (project_id, user_id) AS (
	`
	if len(projectIDs) > 0 {
		sql += " VALUES "
		for i, _ := range projectIDs {
			sql += fmt.Sprintf("($%d::uuid, $1::uuid)", i+2)
		}
	}
	sql += `
	)
	INSERT INTO projects_users (project_id, user_id)
	SELECT project_id, user_id
	FROM new_values
	`
	// Variadic can be frustrating. http://stackoverflow.com/a/12990540
	vars := make([]interface{}, len(projectIDs)+1)
	vars[0] = userID
	for i, v := range projectIDs {
		vars[i+1] = interface{}(v)
	}
	_, err = d.Query(sql, vars...)
	return err
}
