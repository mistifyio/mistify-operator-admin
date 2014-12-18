package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"

	"code.google.com/p/go-uuid/uuid"
	"github.com/mistifyio/mistify-operator-admin/db"
)

// Network describes a set of ipranges
type Network struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata"`
	IPRanges []*IPRange        `json:"-"`
}

// id returns the id, required by the relatable interface
func (network *Network) id() string {
	return network.ID
}

// pkeyName returns the database primary key name, required by the relatable
// interface
func (network *Network) pkeyName() string {
	return "network_id"
}

// Validate ensures the network properties are set correctly
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

// Save persists a network to the database
func (network *Network) Save() error {
	if err := network.Validate(); err != nil {
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

// Delete removes a network from the database
func (network *Network) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM networks WHERE network_id = $1"
	_, err = d.Exec(sql, network.ID)
	return err
}

// Load retrieves a network from the database
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

// fromRows unmarshals a database query result row into the network object
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

// Decode unmarshals JSON into the network object
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

// LoadIPRanges retrieves the ipranges associated with the network from the
// database
func (network *Network) LoadIPRanges() error {
	ipranges, err := IPRangesByNetwork(network)
	if err != nil {
		return err
	}
	network.IPRanges = ipranges
	return nil
}

// AddIPRange adds a relation to an iprange
func (network *Network) AddIPRange(iprange *IPRange) error {
	return AddRelation("iprange_networks", network, iprange)
}

// RemoveIPRange removes a relation with an iprange
func (network *Network) RemoveIPRange(iprange *IPRange) error {
	return RemoveRelation("iprange_networks", network, iprange)
}

// SetIPRanges creates and ensures the only relations the network has with
// ipranges
func (network *Network) SetIPRanges(ipranges []*IPRange) error {
	if len(ipranges) == 0 {
		return ClearRelations("iprange_networks", network)
	}
	relatables := make([]relatable, len(ipranges))
	for i, iprange := range ipranges {
		relatables[i] = relatable(iprange)
	}
	if err := SetRelations("iprange_networks", network, relatables); err != nil {
		return err
	}
	return network.LoadIPRanges()
}

// NewID generates a new uuid ID
func (network *Network) NewID() string {
	network.ID = uuid.New()
	return network.ID
}

// NewNetwork creates and initializes a new network object
func NewNetwork() *Network {
	network := &Network{
		ID:       uuid.New(),
		Metadata: make(map[string]string),
	}
	return network
}

// FetchNetwork retrieves a network object from the database by ID
func FetchNetwork(id string) (*Network, error) {
	network := &Network{
		ID: id,
	}
	if err := network.Load(); err != nil {
		return nil, err
	}
	return network, nil
}

// ListNetworks retrieves an array of all network objects from the database
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
	networks, err := networksFromRows(rows)
	if err != nil {
		return nil, err
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return networks, nil
}

// NetworksByIPRange retrieves an array of all network objects associated with
// an iprange from the database
func NetworksByIPRange(iprange *IPRange) ([]*Network, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT n.network_id, n.name, n.metadata
	FROM networks n
	JOIN iprange_networks i_n ON n.network_id = i_n.network_id
	WHERE i_n.iprange_id = $1
	ORDER BY n.network_id asc
	`
	rows, err := d.Query(sql, iprange.ID)
	if err != nil {
		return nil, err
	}
	networks, err := networksFromRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return networks, nil
}

// networksFromRows unmarshals multiple query rows into an array of networks
func networksFromRows(rows *sql.Rows) ([]*Network, error) {
	networks := make([]*Network, 0, 1)
	for rows.Next() {
		network := &Network{}
		if err := network.fromRows(rows); err != nil {
			return nil, err
		}
		networks = append(networks, network)
	}
	return networks, nil
}
