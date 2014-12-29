package models_test

import (
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
