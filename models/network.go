package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"

	"code.google.com/p/go-uuid/uuid"
	"github.com/mistifyio/mistify-operator-admin/db"
)

type Network struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata"`
}

func (network *Network) id() string {
	return network.ID
}

func (network *Network) pkeyName() string {
	return "network_id"
}

func (network *Network) Validate() error {
	if network.ID == "" {
		return errors.New("missing id")
	}
	if uuid.Parse(network.ID) == nil {
		return errors.New("invalid id. must be uuid")
	}
	if network.Name == "" {
		return errors.New("missing name")
	}
	if network.Metadata == nil {
		return errors.New("metadata must not be nil")
	}
	return nil
}

func (network *Network) Save() error {
	err := network.Validate()
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
	WITH new_values (network_id, name, metadata) as (
		VALUES ($1::uuid, $2, $3::json)
	),
	upsert as (
		UPDATE networks n SET
			name = nv.name,
			metadata = nv.metadata
		FROM new_values nv
		WHERE n.network_id = nv.network_id
		RETURNING nv.network_id
	)
	INSERT INTO networks
		(network_id, name, metadata)
	SELECT network_id, name, metadata
	FROM new_values nv
	WHERE NOT EXISTS (SELECT 1 FROM upsert u WHERE nv.network_id = u.network_id)
	`
	metadata, err := json.Marshal(network.Metadata)
	if err != nil {
		return err
	}
	_, err = d.Exec(sql,
		network.ID,
		network.Name,
		string(metadata),
	)
	return err
}

func (network *Network) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM networks WHERE network_id = $1"
	_, err = d.Exec(sql, network.ID)
	return err
}

func (network *Network) Load() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	SELECT network_id, name, metadata
	FROM networks
	WHERE network_id = $1
	`
	rows, err := d.Query(sql, network.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()
	if err := network.fromRows(rows); err != nil {
		return err
	}
	return rows.Err()
}

func (network *Network) fromRows(rows *sql.Rows) error {
	var metadata string
	err := rows.Scan(
		&network.ID,
		&network.Name,
		&metadata,
	)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(metadata), &network.Metadata)
}

func (network *Network) Decode(data io.Reader) error {
	if err := json.NewDecoder(data).Decode(network); err != nil {
		return err
	}
	if network.Metadata == nil {
		network.Metadata = make(map[string]string)
	} else {
		for key, value := range network.Metadata {
			if value == "" {
				delete(network.Metadata, key)
			}
		}
	}
	return nil
}

func (network *Network) NewID() string {
	network.ID = uuid.New()
	return network.ID
}

func NewNetwork() *Network {
	network := &Network{
		ID:       uuid.New(),
		Metadata: make(map[string]string),
	}
	return network
}

func FetchNetwork(id string) (*Network, error) {
	network := &Network{
		ID: id,
	}
	if err := network.Load(); err != nil {
		return nil, err
	}
	return network, nil
}

func ListNetworks() ([]*Network, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT network_id, name, metadata
	FROM networks
	ORDER BY network_id
	`
	rows, err := d.Query(sql)
	if err != nil {
		return nil, err
	}
	networks := make([]*Network, 0, 1)
	for rows.Next() {
		network := &Network{}
		if err := network.fromRows(rows); err != nil {
			return nil, err
		}
		networks = append(networks, network)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return networks, nil
}
