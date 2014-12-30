package models

import (
	"database/sql"
	"encoding/json"
	"io"
	"net"

	"code.google.com/p/go-uuid/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/mistifyio/mistify-operator-admin/db"
)

type (
	// IPRange describes a segment of IP addresses
	IPRange struct {
		ID          string            `json:"id"`
		CIDR        *net.IPNet        `json:"cidr"`
		Gateway     net.IP            `json:"gateway"`
		Start       net.IP            `json:"start"`
		End         net.IP            `json:"end"`
		Metadata    map[string]string `json:"metadata"`
		Network     *Network          `json:"-"`
		Hypervisors []*Hypervisor     `json:"-"`
	}

	// ipRangeData is a middle-man for JSON and database (un)marshalling
	ipRangeData struct {
		ID       string            `json:"id"`
		CIDR     string            `json:"cidr"`
		Gateway  string            `json:"gateway"`
		Start    string            `json:"start"`
		End      string            `json:"end"`
		Metadata map[string]string `json:"metadata"`
	}
)

// id returns the id, required by the relatable interface
func (iprange *IPRange) id() string {
	return iprange.ID
}

// pkeyName returns the database primary key name, required by the relatable
// interface
func (iprange *IPRange) pkeyName() string {
	return "iprange_id"
}

// importData unmarshals the middle-man structure into an iprange object
func (iprange *IPRange) importData(data *ipRangeData) error {
	_, cidr, err := net.ParseCIDR(data.CIDR)
	if err != nil {
		return err
	}
	iprange.ID = data.ID
	iprange.CIDR = cidr
	iprange.Gateway = net.ParseIP(data.Gateway)
	iprange.Start = net.ParseIP(data.Start)
	iprange.End = net.ParseIP(data.End)
	iprange.Metadata = data.Metadata
	return nil
}

// exportData marshals the iprange object into the middle-man structure
func (iprange *IPRange) exportData() *ipRangeData {
	return &ipRangeData{
		ID:       iprange.ID,
		CIDR:     fmtString(iprange.CIDR),
		Gateway:  fmtString(iprange.Gateway),
		Start:    fmtString(iprange.Start),
		End:      fmtString(iprange.End),
		Metadata: iprange.Metadata,
	}
}

// UnmarshalJSON unmarshals JSON into an iprange object
func (iprange *IPRange) UnmarshalJSON(b []byte) error {
	data := &ipRangeData{}
	if err := json.Unmarshal(b, data); err != nil {
		return err
	}
	if err := iprange.importData(data); err != nil {
		return err
	}
	return nil
}

// MarshalJSON marshals an iprange object into JSON
func (iprange IPRange) MarshalJSON() ([]byte, error) {
	return json.Marshal(iprange.exportData())
}

// Validate ensures the iprange proerties are set correctly
func (iprange *IPRange) Validate() error {
	var results *multierror.Error
	if iprange.ID == "" {
		results = multierror.Append(results, ErrNoID)
	}
	if uuid.Parse(iprange.ID) == nil {
		results = multierror.Append(results, ErrBadID)
	}
	if iprange.CIDR == nil {
		results = multierror.Append(results, ErrNoCIDR)
	}
	if iprange.Gateway == nil {
		results = multierror.Append(results, ErrNoGateway)
	}
	if iprange.Start == nil {
		results = multierror.Append(results, ErrNoStartIP)
	}
	if iprange.End == nil {
		results = multierror.Append(results, ErrNoEndIP)
	}
	if iprange.Metadata == nil {
		results = multierror.Append(results, ErrNilMetadata)
	}
	return results.ErrorOrNil()
}

// Save persists an iprange to the database
func (iprange *IPRange) Save() error {
	if err := iprange.Validate(); err != nil {
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
	WITH new_values (iprange_id, cidr, gateway, start_ip, end_ip, metadata) as (
		VALUES ($1::uuid, $2::cidr, $3::inet, $4::inet, $5::inet, $6::json)
	),
	upsert as (
		UPDATE ipranges i SET
			cidr = nv.cidr,
			gateway = nv.gateway,
			start_ip = nv.start_ip,
			end_ip = nv.end_ip,
			metadata = nv.metadata
		FROM new_values nv
		WHERE i.iprange_id = nv.iprange_id
		RETURNING i.iprange_id
	)
	INSERT INTO ipranges
		(iprange_id, cidr, gateway, start_ip, end_ip, metadata)
	SELECT iprange_id, cidr, gateway, start_ip, end_ip, metadata
	FROM new_values nv
	WHERE NOT EXISTS (SELECT 1 FROM upsert u WHERE nv.iprange_id = u.iprange_id)
    `
	data := iprange.exportData()
	metadata, err := json.Marshal(data.Metadata)
	if err != nil {
		return err
	}
	_, err = d.Exec(sql,
		data.ID,
		data.CIDR,
		data.Gateway,
		data.Start,
		data.End,
		string(metadata),
	)
	return err
}

// Delete removes an iprange from the database
func (iprange *IPRange) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM ipranges WHERE iprange_id = $1"
	_, err = d.Exec(sql, iprange.ID)
	return err
}

// Load retrieves an iprange from the database
func (iprange *IPRange) Load() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := `
	SELECT iprange_id, cidr, gateway, start_ip, end_ip, metadata
	FROM ipranges
	WHERE iprange_id = $1
	`
	rows, err := d.Query(sql, iprange.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()
	if err := iprange.fromRows(rows); err != nil {
		return err
	}
	return rows.Err()
}

// fromRows unmarshals a database query result row into the iprange object
func (iprange *IPRange) fromRows(rows *sql.Rows) error {
	var metadata string
	data := &ipRangeData{}
	err := rows.Scan(
		&data.ID,
		&data.CIDR,
		&data.Gateway,
		&data.Start,
		&data.End,
		&metadata,
	)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(metadata), &data.Metadata); err != nil {
		return err
	}
	return iprange.importData(data)
}

// Decode unmarshals JSON into the flavor object
func (iprange *IPRange) Decode(data io.Reader) error {
	if err := json.NewDecoder(data).Decode(iprange); err != nil {
		return err
	}
	if iprange.Metadata == nil {
		iprange.Metadata = make(map[string]string)
	} else {
		for key, value := range iprange.Metadata {
			if value == "" {
				delete(iprange.Metadata, key)
			}
		}
	}
	return nil
}

// LoadHypervisors retrieves the hypervisors associated with the iprange from
// the database
func (iprange *IPRange) LoadHypervisors() error {
	hypervisors, err := HypervisorsByIPRange(iprange)
	if err != nil {
		return err
	}
	iprange.Hypervisors = hypervisors
	return nil
}

// AddHypervisor adds a relation to a hypervisor
func (iprange *IPRange) AddHypervisor(hypervisor *Hypervisor) error {
	return AddRelation("hypervisors_ipranges", iprange, hypervisor)
}

// RemoveHypervisor removes a relation from a hypervisor
func (iprange *IPRange) RemoveHypervisor(hypervisor *Hypervisor) error {
	return RemoveRelation("hypervisors_ipranges", iprange, hypervisor)
}

// SetHypervisors creates and ensures the only relations the iprange has with
// hypervisors
func (iprange *IPRange) SetHypervisors(hypervisors []*Hypervisor) error {
	if len(hypervisors) == 0 {
		return ClearRelations("hypervisors_ipranges", iprange)
	}
	relatables := make([]relatable, len(hypervisors))
	for i, hypervisor := range hypervisors {
		relatables[i] = relatable(hypervisor)
	}
	if err := SetRelations("hypervisors_ipranges", iprange, relatables); err != nil {
		return err
	}
	return iprange.LoadHypervisors()
}

// LoadNetwork retrieves the network related with the iprange from the
// database
func (iprange *IPRange) LoadNetwork() error {
	networks, err := NetworksByIPRange(iprange)
	if err != nil {
		return err
	}
	if len(networks) > 0 {
		iprange.Network = networks[0]
	} else {
		iprange.Network = nil
	}
	return nil
}

// SetNetwork sets the related network
func (iprange *IPRange) SetNetwork(network *Network) error {
	// Only one can be set at a time
	relatables := make([]relatable, 1)
	relatables[0] = relatable(network)
	return SetRelations("iprange_networks", iprange, relatables)
}

// RemoveNetwork clears the network relation
func (iprange *IPRange) RemoveNetwork(network *Network) error {
	return RemoveRelation("iprange_networks", iprange, network)
}

// NewID generates a new uuid ID
func (iprange *IPRange) NewID() string {
	iprange.ID = uuid.New()
	return iprange.ID
}

// NewIPRange creates and initializes a new iprange object
func NewIPRange() *IPRange {
	iprange := &IPRange{
		ID:       uuid.New(),
		Metadata: make(map[string]string),
	}
	return iprange
}

// FetchIPRange retrieves an iprange object from the database by ID
func FetchIPRange(id string) (*IPRange, error) {
	iprange := &IPRange{
		ID: id,
	}
	if err := iprange.Load(); err != nil {
		return nil, err
	}
	return iprange, nil
}

// ListIPRanges retrieves an array of all iprange objects from the database
func ListIPRanges() ([]*IPRange, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT iprange_id, cidr, gateway, start_ip, end_ip, metadata
	FROM ipranges
	ORDER BY iprange_id
	`
	rows, err := d.Query(sql)
	if err != nil {
		return nil, err
	}
	ipranges := make([]*IPRange, 0, 1)
	for rows.Next() {
		iprange := &IPRange{}
		if err := iprange.fromRows(rows); err != nil {
			return nil, err
		}
		ipranges = append(ipranges, iprange)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return ipranges, nil
}

// IPRangesByHypervisor retrieves an array of iprange objects associated with a
// hypervisor from the database
func IPRangesByHypervisor(hypervisor *Hypervisor) ([]*IPRange, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT i.iprange_id, i.cidr, i.gateway, i.start_ip, i.end_ip, i.metadata
	FROM ipranges i
	JOIN hypervisors_ipranges hi ON i.iprange_id = hi.iprange_id
	WHERE hi.hypervisor_id = $1
	ORDER BY i.iprange_id asc
	`
	rows, err := d.Query(sql, hypervisor.ID)
	if err != nil {
		return nil, err
	}
	ipranges, err := iprangesFromRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ipranges, nil
}

// IPRangesByNetwork retrieves an array of iprange objects associated with a
// network from the database
func IPRangesByNetwork(network *Network) ([]*IPRange, error) {
	d, err := db.Connect(nil)
	if err != nil {
		return nil, err
	}
	sql := `
	SELECT i.iprange_id, i.cidr, i.gateway, i.start_ip, i.end_ip, i.metadata
	FROM ipranges i
	JOIN iprange_networks i_n ON i.iprange_id = i_n.iprange_id
	WHERE i_n.network_id = $1
	ORDER BY i.iprange_id asc
	`
	rows, err := d.Query(sql, network.ID)
	if err != nil {
		return nil, err
	}
	ipranges, err := iprangesFromRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ipranges, nil
}

// iprangesFromRows unmarshals multiple query rows into an array of ipranges
func iprangesFromRows(rows *sql.Rows) ([]*IPRange, error) {
	ipranges := make([]*IPRange, 0, 1)
	for rows.Next() {
		iprange := &IPRange{}
		if err := iprange.fromRows(rows); err != nil {
			return nil, err
		}
		ipranges = append(ipranges, iprange)
	}
	return ipranges, nil
}
