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
	IPRange struct {
		ID       string            `json:"id"`
		CIDR     *net.IPNet        `json:"cidr"`
		Gateway  net.IP            `json:"gateway"`
		Start    net.IP            `json:"start"`
		End      net.IP            `json:"end"`
		Metadata map[string]string `json:"metadata"`
		Network  *Network          `json:"-"`
	}

	ipRangeData struct {
		ID       string            `json:"id"`
		CIDR     string            `json:"cidr"`
		Gateway  string            `json:"gateway"`
		Start    string            `json:"start"`
		End      string            `json:"end"`
		Metadata map[string]string `json:"metadata"`
	}
)

func (iprange *IPRange) id() string {
	return iprange.ID
}

func (iprange *IPRange) pkeyName() string {
	return "iprange_id"
}

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

func (iprange IPRange) MarshalJSON() ([]byte, error) {
	return json.Marshal(iprange.exportData())
}

func (iprange *IPRange) Validate() error {
	if iprange.ID == "" {
		return errors.New("missing id")
	}
	if uuid.Parse(iprange.ID) == nil {
		return errors.New("invalid id. must be uuid")
	}
	if iprange.CIDR == nil {
		return errors.New("missing cidr")
	}
	if iprange.Gateway == nil {
		return errors.New("missing gateway")
	}
	if iprange.Start == nil {
		return errors.New("missing start ip")
	}
	if iprange.End == nil {
		return errors.New("missing end ip")
	}
	return nil
}

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
		RETURNING i.network_id
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

func (iprange *IPRange) Delete() error {
	d, err := db.Connect(nil)
	if err != nil {
		return err
	}
	sql := "DELETE FROM ipranges WHERE iprange_id = $1"
	_, err = d.Exec(sql, iprange.ID)
	return err
}

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

func (iprange *IPRange) NewID() string {
	iprange.ID = uuid.New()
	return iprange.ID
}

func NewIPRange() *IPRange {
	iprange := &IPRange{
		ID:       uuid.New(),
		Metadata: make(map[string]string),
	}
	return iprange
}

func FetchIPRange(id string) (*IPRange, error) {
	iprange := &IPRange{
		ID: id,
	}
	if err := iprange.Load(); err != nil {
		return nil, err
	}
	return iprange, nil
}

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
