package models

import (
	"fmt"
	"strings"

	"github.com/mistifyio/mistify-operator-admin/db"
)

// relatable is an interface that allows for associations between objects
// in a database
type relatable interface {
	id() string
	pkeyName() string
}

// AddRelation creates a new relation between two relatable objects in the
// database
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

// RemoveRelation removes a relation between two relatable objects in the
// database
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

// SetRelations creates and ensures the only relations between a relatable
// object and another relatable type is the provided set
func SetRelations(tableName string, r1 relatable, r2s []relatable) error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}

	// Clear and set relations
	// If we ever need auditing, switch from a txn to a writable CTE
	// that handles deletes/inserts more granularly
	txn, err := d.Begin()
	if err != nil {
		return err
	}

	r1pkey := r1.pkeyName()

	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", tableName, r1pkey)
	if _, err := txn.Exec(deleteSQL, r1.id()); err != nil {
		txn.Rollback()
		return err
	}

	if len(r2s) == 0 {
		txn.Commit()
		return err
	}

	r2pkey := r2s[0].pkeyName()

	placeholders := make([]string, len(r2s))
	values := make([]interface{}, len(r2s)+1)
	values[0] = r1.id()
	for i, r2 := range r2s {
		placeholders[i] = fmt.Sprintf("($1::uuid, $%d::uuid)", i+2)
		values[i+1] = interface{}(r2.id())
	}

	sql := `
	INSERT INTO %s (%s, %s)
	VALUES %s
	`
	sql = fmt.Sprintf(sql,
		tableName, r1pkey, r2pkey,
		strings.Join(placeholders, ","),
	)
	if _, err = txn.Exec(sql, values...); err != nil {
		txn.Rollback()
		return err
	}
	return txn.Commit()
}

// ClearRelations relations removes all relations a relatable object has with another
// relatable object type
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
