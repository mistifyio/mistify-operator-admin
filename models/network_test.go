package models_test

import (
	"strings"
	"testing"

	"code.google.com/p/go-uuid/uuid"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/models"
)

var networkJSON = `{
	"id": "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d",
	"name": "foobar",
	"metadata": {
		"foo": "bar"
	}
}`

func createNetwork(t *testing.T) *models.Network {
	r := strings.NewReader(networkJSON)
	network := &models.Network{}
	h.Ok(t, network.Decode(r))
	return network
}

func checkNetworkValues(t *testing.T, network *models.Network) {
	h.Equals(t, "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d", network.ID)
	h.Equals(t, "foobar", network.Name)
	h.Equals(t, map[string]string{"foo": "bar"}, network.Metadata)
}

func TestNewNetwork(t *testing.T) {
	network := models.NewNetwork()
	h.Assert(t, uuid.Parse(network.ID) != nil, "missing uuid ID")
	h.Assert(t, network.Metadata != nil, "uninitialized metadata")
}

func TestNetworkNewID(t *testing.T) {
	network := models.NewNetwork()
	id1 := network.ID
	network.NewID()
	h.Assert(t, uuid.Parse(network.ID) != nil, "missing uuid ID")
	h.Assert(t, id1 != network.ID, "New ID was not generated")
}

func TestNetworkDecode(t *testing.T) {
	network := createNetwork(t)
	checkNetworkValues(t, network)
}

func TestNetworkValidate(t *testing.T) {
	network := &models.Network{}
	var err error

	err = network.Validate()
	h.Assert(t, errContains(models.ErrNoID, err), "expected ErrNoID")
	h.Assert(t, errContains(models.ErrBadID, err), "expected ErrBadID")
	h.Assert(t, errContains(models.ErrNoName, err), "expected ErrNoName")
	h.Assert(t, errContains(models.ErrNilMetadata, err), "expected ErrNilMetadata")

	network.ID = "foobar"
	err = network.Validate()
	h.Assert(t, errDoesNotContain(models.ErrNoID, err), "did not expect ErrNoID")
	h.Assert(t, errContains(models.ErrBadID, err), "expected ErrBadID")

	network.NewID()
	h.Assert(t, errDoesNotContain(models.ErrBadID, network.Validate()), "did not expect ErrBadID")

	network.Name = "foobar"
	h.Assert(t, errDoesNotContain(models.ErrNoName, network.Validate()), "did not expect ErrNoName")

	network.Metadata = make(map[string]string)
	err = network.Validate()
	h.Assert(t, errDoesNotContain(models.ErrNilMetadata, err), "did not expect ErrNilMetadata")

	h.Ok(t, err)
}

func TestNetworkSave(t *testing.T) {
	network := createNetwork(t)
	h.Ok(t, network.Save())
}

func TestNetworkDelete(t *testing.T) {
	network := createNetwork(t)
	h.Ok(t, network.Delete())
}

func TestNetworkLoad(t *testing.T) {
	network := createNetwork(t)
	h.Ok(t, network.Save())

	network2 := models.NewNetwork()
	network2.ID = network.ID

	h.Ok(t, network2.Load())
	checkNetworkValues(t, network2)
	h.Ok(t, network2.Delete())
}

func TestFetchNetwork(t *testing.T) {
	network := createNetwork(t)
	h.Ok(t, network.Save())

	network2, err := models.FetchNetwork("ebf3bfd5-9915-4ed1-bcb3-117bb48b155d")
	h.Ok(t, err)
	checkNetworkValues(t, network2)
	h.Ok(t, network2.Delete())
}

func TestListNetworks(t *testing.T) {
	network := createNetwork(t)
	h.Ok(t, network.Save())

	networks, err := models.ListNetworks()
	h.Ok(t, err)
	h.Equals(t, 1, len(networks))
	checkNetworkValues(t, networks[0])
	h.Ok(t, network.Delete())
}

func TestNetworkIPRangeRelations(t *testing.T) {
	// Prep
	network := createNetwork(t)
	h.Ok(t, network.Save())
	iprange := createIPRange(t)
	h.Ok(t, iprange.Save())

	// Add
	h.Ok(t, network.AddIPRange(iprange))

	// Load
	h.Ok(t, network.LoadIPRanges())
	h.Equals(t, 1, len(network.IPRanges))

	// Remove
	h.Ok(t, network.RemoveIPRange(iprange))
	h.Ok(t, network.LoadIPRanges())
	h.Equals(t, 0, len(network.IPRanges))

	// Set
	h.Ok(t, network.SetIPRanges([]*models.IPRange{iprange}))
	h.Ok(t, network.LoadIPRanges())
	h.Equals(t, 1, len(network.IPRanges))

	// Lookup networks by iprange
	networks, err := models.NetworksByIPRange(iprange)
	h.Ok(t, err)
	h.Equals(t, 1, len(networks))

	// Clear
	h.Ok(t, network.SetIPRanges(make([]*models.IPRange, 0)))
	h.Ok(t, network.LoadIPRanges())
	h.Equals(t, 0, len(network.IPRanges))

	// Cleanup
	h.Ok(t, iprange.Delete())
	h.Ok(t, network.Delete())
}
