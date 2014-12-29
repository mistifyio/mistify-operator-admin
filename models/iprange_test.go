package models_test

import (
	"strings"
	"testing"

	"code.google.com/p/go-uuid/uuid"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/models"
)

var iprangeJSON = `{
	"id": "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d",
	"cidr": "192.168.1.0/24",
	"gateway": "192.168.1.1",
	"start": "192.168.1.10",
	"end": "192.168.1.255",
	"metadata": {
		"foo": "bar"
	}
}`

func createIPRange(t *testing.T) *models.IPRange {
	r := strings.NewReader(iprangeJSON)
	iprange := &models.IPRange{}
	h.Ok(t, iprange.Decode(r))
	return iprange
}

func checkIPRangeValues(t *testing.T, iprange *models.IPRange) {
	h.Equals(t, "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d", iprange.ID)
	h.Equals(t, "192.168.1.0/24", iprange.CIDR.String())
	h.Equals(t, "192.168.1.1", iprange.Gateway.String())
	h.Equals(t, "192.168.1.10", iprange.Start.String())
	h.Equals(t, "192.168.1.255", iprange.End.String())
	h.Equals(t, map[string]string{"foo": "bar"}, iprange.Metadata)
}

func TestNewIPRange(t *testing.T) {
	iprange := models.NewIPRange()
	h.Assert(t, uuid.Parse(iprange.ID) != nil, "missing uuid ID")
	h.Assert(t, iprange.Metadata != nil, "uninitialized metadata")
}

func TestIPRangeNewID(t *testing.T) {
	iprange := models.NewIPRange()
	id1 := iprange.ID
	iprange.NewID()
	h.Assert(t, uuid.Parse(iprange.ID) != nil, "missing uuid ID")
	h.Assert(t, id1 != iprange.ID, "New ID was not generated")
}

func TestIPRangeUnmarshalJSON(t *testing.T) {
	iprange := &models.IPRange{}
	h.Ok(t, iprange.UnmarshalJSON([]byte(iprangeJSON)))
	checkIPRangeValues(t, iprange)
}

func TestIPRangeDecode(t *testing.T) {
	iprange := createIPRange(t)
	checkIPRangeValues(t, iprange)
}

func TestIPRangeMarshalJSON(t *testing.T) {
	iprange := createIPRange(t)
	_, err := iprange.MarshalJSON()
	h.Ok(t, err)
}

func TestIPRangeSave(t *testing.T) {
	iprange := createIPRange(t)
	h.Ok(t, iprange.Save())
}

func TestIPRangeDelete(t *testing.T) {
	iprange := createIPRange(t)
	h.Ok(t, iprange.Delete())
}

func TestIPRangeLoad(t *testing.T) {
	iprange := createIPRange(t)
	h.Ok(t, iprange.Save())

	iprange2 := models.NewIPRange()
	iprange2.ID = iprange.ID

	h.Ok(t, iprange2.Load())
	checkIPRangeValues(t, iprange2)
	h.Ok(t, iprange2.Delete())
}

func TestFetchIPRange(t *testing.T) {
	iprange := createIPRange(t)
	h.Ok(t, iprange.Save())

	iprange2, err := models.FetchIPRange("ebf3bfd5-9915-4ed1-bcb3-117bb48b155d")
	h.Ok(t, err)
	checkIPRangeValues(t, iprange2)
	h.Ok(t, iprange2.Delete())
}

func TestListIPRanges(t *testing.T) {
	iprange := createIPRange(t)
	h.Ok(t, iprange.Save())

	ipranges, err := models.ListIPRanges()
	h.Ok(t, err)
	h.Equals(t, 1, len(ipranges))
	checkIPRangeValues(t, ipranges[0])
	h.Ok(t, iprange.Delete())
}
