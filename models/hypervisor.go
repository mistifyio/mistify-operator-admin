package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"local/mistify-operator-admin/db"
	"net"

	"code.google.com/p/go-uuid/uuid"
)

type (
	Hypervisor struct {
		ID       string            `json:"id"`
		Mac      net.HardwareAddr  `json:"mac"`
		IPv6     net.IP            `json:"ipv6"`
		Metadata map[string]string `json:"metadata"`
		IPRanges []*IPRange        `json:"-"`
	}
)

func (hypervisor *Hypervisor) Validate() error {
	if hypervisor.ID == "" {
		return errors.New("missing id")
	}
	if uuid.Parse(hypervisor.ID) == nil {
		return errors.New("invalid id. must be uuid")
	}
	if hypervisor.Mac == nil {
		return errors.New("missing mac")
	}
	if hypervisor.IPv6 == nil {
		return errors.New("missing ipv6")
	}
	if hypervisor.Metadata == nil {
		return errors.New("missing metadata")
	}
	return nil
}

func (hypervisor *Hypervisor) Save() error {
	if err := hypervisor.Validate(); err != nil {
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
	WITH new_values (hypervisor_id, mac, ipv6, metadata) as (
		VALUES ($1::uuid, $2::macaddr, $3::inet $4::json)
	),
	upsert as (
		UPDATE hypervisors h SET
			mac = nv.mac,
			ipv6 = nv.ipv6,
			metadata = nv.metadata
		FROM new_values nv
		WHERE h.hypervisor_id = nv.hypervisor_id
		RETURNING h.hypervisor_id
	)
	INSERT INTO hypervisors
		(hypervisor_id, mac, ipv6, metadata)
	SELECT iprange_isor_id, mac, ipv6, metadata
	FROM new_values nv
	WHERE NOT EXISTS (SELECT 1 FROM upsert u WHERE nv.hypervisor_id = u.hypervisor_id)
    `
	metadata, err := json.Marshal(hypervisor.Metadata)
	if err != nil {
		return err
	}
	_, err = d.Exec(sql,
		hypervisor.ID,
		hypervisor.Mac,
		hypervisor.IPv6,
		string(metadata),
	)
	return err
}

func (hypervisor *Hypervisor) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM hypervisors WHERE hypervisor_id = $1"
	_, err = d.Exec(sql, hypervisor.ID)
	return err
}

func (hypervisor *Hypervisor) Load() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	SELECT hypervisor_id, mac, ipv6, metadata
	FROM hypervisors
	WHERE hypervisor_id = $1
	`
	rows, err := d.Query(sql, hypervisor.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()
	if err := hypervisor.fromRows(rows); err != nil {
		return err
	}
	return rows.Err()
}

func (hypervisor *Hypervisor) fromRows(rows *sql.Rows) error {
	var metadata string
	err := rows.Scan(
		&hypervisor.ID,
		&hypervisor.Mac,
		&hypervisor.IPv6,
		&metadata,
	)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(metadata), &hypervisor.Metadata); err != nil {
		return err
	}
	return nil
}

func (hypervisor *Hypervisor) Decode(data io.Reader) error {
	if err := json.NewDecoder(data).Decode(hypervisor); err != nil {
		return err
	}
	if hypervisor.Metadata == nil {
		hypervisor.Metadata = make(map[string]string)
	} else {
		for key, value := range hypervisor.Metadata {
			if value == "" {
				delete(hypervisor.Metadata, key)
			}
		}
	}
	return nil
}

func (hypervisor *Hypervisor) NewID() string {
	hypervisor.ID = uuid.New()
	return hypervisor.ID
}

func NewHypervisor() *Hypervisor {
	hypervisor := &Hypervisor{
		ID:       uuid.New(),
		Metadata: make(map[string]string),
	}
	return hypervisor
}

func FetchHypervisor(id string) (*Hypervisor, error) {
	hypervisor := &Hypervisor{
		ID: id,
	}
	if err := hypervisor.Load(); err != nil {
		return nil, err
	}
	return hypervisor, nil
}

func ListHypervisors() ([]*Hypervisor, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT hypervisor_id, mac, ipv6, metadata
	FROM hypervisors
	ORDER BY hypervisor_id
	`
	rows, err := d.Query(sql)
	if err != nil {
		return nil, err
	}
	hypervisors := make([]*Hypervisor, 0, 1)
	for rows.Next() {
		hypervisor := &Hypervisor{}
		if err := hypervisor.fromRows(rows); err != nil {
			return nil, err
		}
		hypervisors = append(hypervisors, hypervisor)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return hypervisors, nil
}
