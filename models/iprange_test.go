package models_test

import (
	"net"
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

func TestIPRangeValidate(t *testing.T) {
	iprange := &models.IPRange{}
	var err error

	err = iprange.Validate()
	h.Assert(t, errContains(models.ErrNoID, err), "expected ErrNoID")
	h.Assert(t, errContains(models.ErrBadID, err), "expected ErrBadID")
	h.Assert(t, errContains(models.ErrNoCIDR, err), "expected ErrNoCIDR")
	h.Assert(t, errContains(models.ErrNoGateway, err), "expected ErrNoGateway")
	h.Assert(t, errContains(models.ErrNoStartIP, err), "expected ErrNoStartIP")
	h.Assert(t, errContains(models.ErrNoEndIP, err), "expected ErrNoEndIP")
	h.Assert(t, errContains(models.ErrNilMetadata, err), "expected ErrNilMetadata")

	iprange.ID = "foobar"
	err = iprange.Validate()
	h.Assert(t, errDoesNotContain(models.ErrNoID, err), "did not expect ErrNoID")
	h.Assert(t, errContains(models.ErrBadID, err), "expected ErrBadID")

	iprange.NewID()
	h.Assert(t, errDoesNotContain(models.ErrBadID, iprange.Validate()), "did not expect ErrBadID")

	_, cidr, err := net.ParseCIDR("192.168.1.0/24")
	h.Ok(t, err)
	iprange.CIDR = cidr
	h.Assert(t, errDoesNotContain(models.ErrNoCIDR, iprange.Validate()), "did not expect ErrNoCidr")

	iprange.Gateway = net.ParseIP("192.168.1.1")
	h.Assert(t, errDoesNotContain(models.ErrNoGateway, iprange.Validate()), "did not expect ErrNoGateway")

	iprange.Start = net.ParseIP("192.168.1.10")
	h.Assert(t, errDoesNotContain(models.ErrNoStartIP, iprange.Validate()), "did not expect ErrNoStartIP")

	iprange.End = net.ParseIP("192.168.1.255")
	h.Assert(t, errDoesNotContain(models.ErrNoEndIP, iprange.Validate()), "did not expect ErrNoEndIP")

	iprange.Metadata = make(map[string]string)
	err = iprange.Validate()
	h.Assert(t, errDoesNotContain(models.ErrNilMetadata, err), "did not expect ErrNilMetadata")

	h.Ok(t, err)
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

func TestIPRangeHypervisorRelations(t *testing.T) {
	// Prep
	iprange := createIPRange(t)
	h.Ok(t, iprange.Save())
	hypervisor := createHypervisor(t)
	h.Ok(t, hypervisor.Save())

	// Add
	h.Ok(t, iprange.AddHypervisor(hypervisor))

	// Load
	h.Ok(t, iprange.LoadHypervisors())
	h.Equals(t, 1, len(iprange.Hypervisors))

	// Remove
	h.Ok(t, iprange.RemoveHypervisor(hypervisor))
	h.Ok(t, iprange.LoadHypervisors())
	h.Equals(t, 0, len(iprange.Hypervisors))

	// Set
	h.Ok(t, iprange.SetHypervisors([]*models.Hypervisor{hypervisor}))
	h.Ok(t, iprange.LoadHypervisors())
	h.Equals(t, 1, len(iprange.Hypervisors))

	// Lookup ipranges by hypervisors
	ipranges, err := models.IPRangesByHypervisor(hypervisor)
	h.Ok(t, err)
	h.Equals(t, 1, len(ipranges))

	// Clear
	h.Ok(t, iprange.SetHypervisors(make([]*models.Hypervisor, 0)))
	h.Ok(t, iprange.LoadHypervisors())
	h.Equals(t, 0, len(iprange.Hypervisors))

	// Cleanup
	h.Ok(t, hypervisor.Delete())
	h.Ok(t, iprange.Delete())
}

func TestIPRangeNetworkRelations(t *testing.T) {
	// Prep
	iprange := createIPRange(t)
	h.Ok(t, iprange.Save())
	network := createNetwork(t)
	h.Ok(t, network.Save())

	// Set
	h.Ok(t, iprange.SetNetwork(network))

	// Load
	h.Ok(t, iprange.LoadNetwork())
	h.Assert(t, iprange.Network != nil, "nil network")

	// Lookup ipranges by networks
	ipranges, err := models.IPRangesByNetwork(network)
	h.Ok(t, err)
	h.Equals(t, 1, len(ipranges))

	// Remove
	h.Ok(t, iprange.RemoveNetwork(network))
	h.Ok(t, iprange.LoadNetwork())
	h.Assert(t, iprange.Network == nil, "non nil network")

	// Cleanup
	h.Ok(t, network.Delete())
	h.Ok(t, iprange.Delete())
}
