package models_test

import (
	"net"
	"strings"
	"testing"

	"code.google.com/p/go-uuid/uuid"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/models"
)

var hypervisorJSON = `{
	"id": "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d",
	"mac": "01:23:45:67:89:ab",
	"ip": "192.168.1.20",
	"metadata": {
		"foo": "bar"
	}
}`

func createHypervisor(t *testing.T) *models.Hypervisor {
	r := strings.NewReader(hypervisorJSON)
	hypervisor := &models.Hypervisor{}
	h.Ok(t, hypervisor.Decode(r))
	return hypervisor
}

func checkHypervisorValues(t *testing.T, hypervisor *models.Hypervisor) {
	h.Equals(t, "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d", hypervisor.ID)
	h.Equals(t, "01:23:45:67:89:ab", hypervisor.MAC.String())
	h.Equals(t, "192.168.1.20", hypervisor.IP.String())
	h.Equals(t, map[string]string{"foo": "bar"}, hypervisor.Metadata)
}

func TestNewHypervisor(t *testing.T) {
	hypervisor := models.NewHypervisor()
	h.Assert(t, uuid.Parse(hypervisor.ID) != nil, "missing uuid ID")
	h.Assert(t, hypervisor.Metadata != nil, "uninitialized metadata")
}

func TestHypervisorNewID(t *testing.T) {
	hypervisor := models.NewHypervisor()
	id1 := hypervisor.ID
	hypervisor.NewID()
	h.Assert(t, uuid.Parse(hypervisor.ID) != nil, "missing uuid ID")
	h.Assert(t, id1 != hypervisor.ID, "New ID was not generated")
}

func TestHypervisorUnmarshalJSON(t *testing.T) {
	hypervisor := &models.Hypervisor{}
	h.Ok(t, hypervisor.UnmarshalJSON([]byte(hypervisorJSON)))
	checkHypervisorValues(t, hypervisor)
}

func TestHypervisorDecode(t *testing.T) {
	hypervisor := createHypervisor(t)
	checkHypervisorValues(t, hypervisor)
}

func TestHypervisorValidate(t *testing.T) {
	hypervisor := &models.Hypervisor{}
	var err error

	err = hypervisor.Validate()
	h.Assert(t, errContains(models.ErrNoID, err), "expected ErrNoID")
	h.Assert(t, errContains(models.ErrBadID, err), "expected ErrBadID")
	h.Assert(t, errContains(models.ErrNoMAC, err), "expected ErrNoMAC")
	h.Assert(t, errContains(models.ErrNoIP, err), "expected ErrNoIP")
	h.Assert(t, errContains(models.ErrNilMetadata, err), "expected ErrNilMetadata")

	hypervisor.ID = "foobar"
	err = hypervisor.Validate()
	h.Assert(t, errDoesNotContain(models.ErrNoID, err), "did not expect ErrNoID")
	h.Assert(t, errContains(models.ErrBadID, err), "expected ErrBadID")
	hypervisor.NewID()
	h.Assert(t, errDoesNotContain(models.ErrBadID, hypervisor.Validate()), "did not expect ErrBadID")

	hypervisor.MAC, err = net.ParseMAC("01:23:45:67:89:ab")
	h.Assert(t, errDoesNotContain(models.ErrNoMAC, hypervisor.Validate()), "did not expect ErrNoMac")

	hypervisor.IP = net.ParseIP("192.168.1.1")
	h.Assert(t, errDoesNotContain(models.ErrNoIP, hypervisor.Validate()), "did not expect ErrNoIP")

	hypervisor.Metadata = make(map[string]string)
	err = hypervisor.Validate()
	h.Assert(t, errDoesNotContain(models.ErrNilMetadata, err), "did not expect ErrNilMetadata")

	h.Ok(t, err)
}

func TestHypervisorMarshalJSON(t *testing.T) {
	hypervisor := createHypervisor(t)
	_, err := hypervisor.MarshalJSON()
	h.Ok(t, err)
}

func TestHypervisorSave(t *testing.T) {
	hypervisor := createHypervisor(t)
	h.Ok(t, hypervisor.Save())
}

func TestHypervisorDelete(t *testing.T) {
	hypervisor := createHypervisor(t)
	h.Ok(t, hypervisor.Delete())
}

func TestHypervisorLoad(t *testing.T) {
	hypervisor := createHypervisor(t)
	h.Ok(t, hypervisor.Save())

	hypervisor2 := models.NewHypervisor()
	hypervisor2.ID = hypervisor.ID

	h.Ok(t, hypervisor2.Load())
	checkHypervisorValues(t, hypervisor2)
	h.Ok(t, hypervisor2.Delete())
}

func TestFetchHypervisor(t *testing.T) {
	hypervisor := createHypervisor(t)
	h.Ok(t, hypervisor.Save())

	hypervisor2, err := models.FetchHypervisor("ebf3bfd5-9915-4ed1-bcb3-117bb48b155d")
	h.Ok(t, err)
	checkHypervisorValues(t, hypervisor2)
	h.Ok(t, hypervisor2.Delete())
}

func TestListHypervisors(t *testing.T) {
	hypervisor := createHypervisor(t)
	h.Ok(t, hypervisor.Save())

	hypervisors, err := models.ListHypervisors()
	h.Ok(t, err)
	h.Equals(t, 1, len(hypervisors))
	checkHypervisorValues(t, hypervisors[0])
	h.Ok(t, hypervisor.Delete())
}

func TestHypervisorIPRangeRelations(t *testing.T) {
	// Prep
	hypervisor := createHypervisor(t)
	h.Ok(t, hypervisor.Save())
	iprange := createIPRange(t)
	h.Ok(t, iprange.Save())

	// Add
	h.Ok(t, hypervisor.AddIPRange(iprange))

	// Load
	h.Ok(t, hypervisor.LoadIPRanges())
	h.Equals(t, 1, len(hypervisor.IPRanges))

	// Remove
	h.Ok(t, hypervisor.RemoveIPRange(iprange))
	h.Ok(t, hypervisor.LoadIPRanges())
	h.Equals(t, 0, len(hypervisor.IPRanges))

	// Set
	h.Ok(t, hypervisor.SetIPRanges([]*models.IPRange{iprange}))
	h.Ok(t, hypervisor.LoadIPRanges())
	h.Equals(t, 1, len(hypervisor.IPRanges))

	// Lookup hypervisors by iprange
	hypervisors, err := models.HypervisorsByIPRange(iprange)
	h.Ok(t, err)
	h.Equals(t, 1, len(hypervisors))

	// Clear
	h.Ok(t, hypervisor.SetIPRanges(make([]*models.IPRange, 0)))
	h.Ok(t, hypervisor.LoadIPRanges())
	h.Equals(t, 0, len(hypervisor.IPRanges))

	// Cleanup
	h.Ok(t, iprange.Delete())
	h.Ok(t, hypervisor.Delete())
}
