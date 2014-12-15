package models

import (
	"fmt"

	"github.com/mistifyio/mistify-operator-admin/db"
)

type relatable interface {
	id() string
	pkeyName() string
}

func AddRelation(tableName string, r1 relatable, r2 relatable) error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	INSERT INTO %s (%s, %s)
	SELECT $1, $2
	WHERE NOT EXISTS (SELECT 1 FROM %s WHERE %s = $1 AND %s = $2)
	`
	sql = fmt.Sprintf(sql,
		tableName, r1.pkeyName(), r2.pkeyName(),
		tableName, r1.pkeyName(), r2.pkeyName(),
	)
	_, err = d.Exec(sql, r1.id(), r2.id())
	return err
}

func RemoveRelation(tableName string, r1 relatable, r2 relatable) error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	DELETE FROM %s
	WHERE %s = $1 AND %s = $2
	`
	sql = fmt.Sprintf(sql,
		tableName,
		r1.pkeyName(), r2.pkeyName(),
	)
	_, err = d.Exec(sql, r1.id(), r2.id())
	return err
}

func SetRelations(tableName string, r1 relatable, r2s []relatable) error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}

	if len(r2s) == 0 {
		return nil
	}

	values := ""
	vars := make([]interface{}, len(r2s)+1)
	vars[0] = r1.id()
	values += "VALUES "
	for i, r2 := range r2s {
		values += fmt.Sprintf("($1::uuid, $%d::uuid)", i+2)
		vars[i+1] = interface{}(r2.id())
	}

	r1pkey := r1.pkeyName()
	r2pkey := r2s[0].pkeyName()

	sql := `
	WITH deletes AS (
		DELETE FROM %s
		WHERE %s = $1
	),
	new_values (%s, %s) AS (
		%s
	)
	INSERT INTO %s (%s, %s)
	SELECT %s, %s
	FROM new_values
	`
	sql = fmt.Sprintf(sql,
		tableName,
		r1pkey,
		r1pkey, r2pkey,
		values,
		tableName, r1pkey, r2pkey,
		r1pkey, r2pkey,
	)
	_, err = d.Exec(sql, vars...)
	return err
}

func ClearRelations(tableName string, r1 relatable) error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	DELETE FROM %s
	WHERE %s = $1
	`
	sql = fmt.Sprintf(sql,
		tableName,
		r1.pkeyName(),
	)
	_, err = d.Exec(sql, r1.id())
	return err
}
