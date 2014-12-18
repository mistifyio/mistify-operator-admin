package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"

	"code.google.com/p/go-uuid/uuid"
	"github.com/mistifyio/mistify-operator-admin/db"
)

type (
	// Hypervisor describes a machine where guests will be running
	Hypervisor struct {
		ID       string            `json:"id"`
		MAC      net.HardwareAddr  `json:"mac"`
		IP       net.IP            `json:"ip"`
		Metadata map[string]string `json:"metadata"`
		IPRanges []*IPRange        `json:"-"`
	}

	// hypervisorData is a middle-man for JSON and database (un)marshalling
	hypervisorData struct {
		ID       string            `json:"id"`
		MAC      string            `json:"mac"`
		IP       string            `json:"ip"`
		Metadata map[string]string `json:"metadata"`
	}
)

// id returns the id, required by the relatable interface
func (hypervisor *Hypervisor) id() string {
	return hypervisor.ID
}

// pkeyName returns the database primary key name, required by the relatable
// interface
func (hypervisor *Hypervisor) pkeyName() string {
	return "hypervisor_id"
}

// importData unmarshals the middle-man structure into a hypervisor object
func (hypervisor *Hypervisor) importData(data *hypervisorData) error {
	mac, err := net.ParseMAC(data.MAC)
	if err != nil {
		return err
	}
	hypervisor.ID = data.ID
	hypervisor.MAC = mac
	hypervisor.IP = net.ParseIP(data.IP)
	hypervisor.Metadata = data.Metadata
	return nil
}

// exportData marshals the hypervisor object into the middle-man structure
func (hypervisor *Hypervisor) exportData() *hypervisorData {
	return &hypervisorData{
		ID:       hypervisor.ID,
		MAC:      fmtString(hypervisor.MAC),
		IP:       fmtString(hypervisor.IP),
		Metadata: hypervisor.Metadata,
	}
}

// UnmarshalJSON unmarshals JSON into a hypervisor
func (hypervisor *Hypervisor) UnmarshalJSON(b []byte) error {
	data := &hypervisorData{}
	if err := json.Unmarshal(b, data); err != nil {
		return err
	}
	if err := hypervisor.importData(data); err != nil {
		return err
	}
	return nil
}

// MarshalJSON marshals a hypervisor into JSON
func (hypervisor Hypervisor) MarshalJSON() ([]byte, error) {
	return json.Marshal(hypervisor.exportData())
}

// Validate ensures the hypervisor properties are set correctly
func (hypervisor *Hypervisor) Validate() error {
	if hypervisor.ID == "" {
		return errors.New("missing id")
	}
	if uuid.Parse(hypervisor.ID) == nil {
		return errors.New("invalid id. must be uuid")
	}
	if hypervisor.MAC == nil {
		return errors.New("missing mac")
	}
	if hypervisor.IP == nil {
		return errors.New("missing ip")
	}
	if hypervisor.Metadata == nil {
		return errors.New("missing metadata")
	}
	return nil
}

// Save persists a hypervisor to the database
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
	WITH new_values (hypervisor_id, mac, ip, metadata) as (
		VALUES ($1::uuid, $2::macaddr, $3::inet, $4::json)
	),
	upsert as (
		UPDATE hypervisors h SET
			mac = nv.mac,
			ip = nv.ip,
			metadata = nv.metadata
		FROM new_values nv
		WHERE h.hypervisor_id = nv.hypervisor_id
		RETURNING h.hypervisor_id
	)
	INSERT INTO hypervisors
		(hypervisor_id, mac, ip, metadata)
	SELECT hypervisor_id, mac, ip, metadata
	FROM new_values nv
	WHERE NOT EXISTS (SELECT 1 FROM upsert u WHERE nv.hypervisor_id = u.hypervisor_id)
    `
	data := hypervisor.exportData()
	metadata, err := json.Marshal(data.Metadata)
	if err != nil {
		return err
	}
	fmt.Println(data)
	_, err = d.Exec(sql,
		data.ID,
		data.MAC,
		data.IP,
		string(metadata),
	)
	return err
}

// Delete removes a hypervisor from the database
func (hypervisor *Hypervisor) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM hypervisors WHERE hypervisor_id = $1"
	_, err = d.Exec(sql, hypervisor.ID)
	return err
}

// Load retrieves a hypervisor from the database
func (hypervisor *Hypervisor) Load() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	SELECT hypervisor_id, mac, ip, metadata
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

// fromRows unmarshals a database query result row into the hypervisor object
func (hypervisor *Hypervisor) fromRows(rows *sql.Rows) error {
	var metadata string
	data := &hypervisorData{}
	err := rows.Scan(
		&data.ID,
		&data.MAC,
		&data.IP,
		&metadata,
	)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(metadata), &data.Metadata); err != nil {
		return err
	}
	return hypervisor.importData(data)
}

// Decode unmarshals JSON into the flavor object
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

// LoadIPRanges retrieves all of the ipranges related to the hypervisor
func (hypervisor *Hypervisor) LoadIPRanges() error {
	ipranges, err := IPRangesByHypervisor(hypervisor)
	if err != nil {
		return err
	}
	hypervisor.IPRanges = ipranges
	return nil
}

// AddIPRange adds a relation to an iprange
func (hypervisor *Hypervisor) AddIPRange(iprange *IPRange) error {
	return AddRelation("hypervisors_ipranges", hypervisor, iprange)
}

// RemoveIPRange removes a relation with an iprange
func (hypervisor *Hypervisor) RemoveIPRange(iprange *IPRange) error {
	return RemoveRelation("hypervisors_ipranges", hypervisor, iprange)
}

// SetIPRanges creates and ensures the only relations the hypervisor has with
// ipranges
func (hypervisor *Hypervisor) SetIPRanges(ipranges []*IPRange) error {
	if len(ipranges) == 0 {
		return ClearRelations("hypervisors_ipranges", hypervisor)
	}
	relatables := make([]relatable, len(ipranges))
	for i, iprange := range ipranges {
		relatables[i] = relatable(iprange)
	}
	if err := SetRelations("hypervisors_ipranges", hypervisor, relatables); err != nil {
		return err
	}
	return hypervisor.LoadIPRanges()
}

// NewID generates a new uuid ID
func (hypervisor *Hypervisor) NewID() string {
	hypervisor.ID = uuid.New()
	return hypervisor.ID
}

// NewHypervisor creates and initializes a new hypervisor object
func NewHypervisor() *Hypervisor {
	hypervisor := &Hypervisor{
		ID:       uuid.New(),
		Metadata: make(map[string]string),
	}
	return hypervisor
}

// FetchHypervisor retrieves a hypervisor object from the database by ID
func FetchHypervisor(id string) (*Hypervisor, error) {
	hypervisor := &Hypervisor{
		ID: id,
	}
	if err := hypervisor.Load(); err != nil {
		return nil, err
	}
	return hypervisor, nil
}

// ListHypervisors retrieves an array of all hypervisor objects from the
// database
func ListHypervisors() ([]*Hypervisor, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT hypervisor_id, mac, ip, metadata
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

// HypervisorsByIPRange retrieves an array of hypervisors associated with an
// iprange from the database
func HypervisorsByIPRange(iprange *IPRange) ([]*Hypervisor, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT h.hypervisor_id, h.mac, h.ip, h.metadata
	FROM hypervisors h
	JOIN hypervisors_ipranges hi ON h.hypervisor_id = hi.hypervisor_id
	WHERE hi.iprange_id = $1
	ORDER BY h.hypervisor_id asc
	`
	rows, err := d.Query(sql, iprange.ID)
	if err != nil {
		return nil, err
	}
	hypervisors, err := hypervisorsFromRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hypervisors, nil
}

// hypervisorsFromRows unmarhsals multiple query rows into an array of hypervisors
func hypervisorsFromRows(rows *sql.Rows) ([]*Hypervisor, error) {
	hypervisors := make([]*Hypervisor, 0, 1)
	for rows.Next() {
		hypervisor := &Hypervisor{}
		if err := hypervisor.fromRows(rows); err != nil {
			return nil, err
		}
		hypervisors = append(hypervisors, hypervisor)
	}
	return hypervisors, nil
}
