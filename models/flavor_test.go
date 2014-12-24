package models_test

import (
	"strings"
	"testing"

	"code.google.com/p/go-uuid/uuid"
	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/models"
)

var flavorJSON = `{
	"id": "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d",
	"name": "fooName",
	"cpu": 5,
	"memory": 10,
	"disk": 15,
	"metadata": {
		"foo": "bar"
	}
}`

func createFlavor(t *testing.T) *models.Flavor {
	r := strings.NewReader(flavorJSON)
	flavor := &models.Flavor{}
	h.Ok(t, flavor.Decode(r))
	return flavor
}

func checkFlavorValues(t *testing.T, flavor *models.Flavor) {
	h.Equals(t, "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d", flavor.ID)
	h.Equals(t, "fooName", flavor.Name)
	h.Equals(t, 5, flavor.CPU)
	h.Equals(t, 10, flavor.Memory)
	h.Equals(t, 15, flavor.Disk)
	h.Equals(t, map[string]string{"foo": "bar"}, flavor.Metadata)
}

func TestNewFlavor(t *testing.T) {
	flavor := models.NewFlavor()
	h.Assert(t, uuid.Parse(flavor.ID) != nil, "missing uuid ID")
	h.Assert(t, flavor.Metadata != nil, "uninitialized metadata")
}

func TestFlavorNewID(t *testing.T) {
	flavor := models.NewFlavor()
	id1 := flavor.ID
	flavor.NewID()
	h.Assert(t, uuid.Parse(flavor.ID) != nil, "missing uuid ID")
	h.Assert(t, id1 != flavor.ID, "New ID was not generated")
}

func TestFlavorDecode(t *testing.T) {
	flavor := createFlavor(t)
	checkFlavorValues(t, flavor)
}

func TestFlavorValidate(t *testing.T) {
	// TODO: Validation test
}

func TestFlavorSave(t *testing.T) {
	flavor := createFlavor(t)
	h.Ok(t, flavor.Save())
}

func TestFlavorDelete(t *testing.T) {
	flavor := createFlavor(t)
	h.Ok(t, flavor.Delete())
}

func TestFlavorLoad(t *testing.T) {
	flavor := createFlavor(t)
	h.Ok(t, flavor.Save())

	flavor2 := models.NewFlavor()
	flavor2.ID = flavor.ID

	h.Ok(t, flavor2.Load())
	checkFlavorValues(t, flavor2)
	h.Ok(t, flavor2.Delete())
}

func TestFetchFlavor(t *testing.T) {
	flavor := createFlavor(t)
	h.Ok(t, flavor.Save())

	flavor2, err := models.FetchFlavor("ebf3bfd5-9915-4ed1-bcb3-117bb48b155d")
	h.Ok(t, err)
	checkFlavorValues(t, flavor2)
	h.Ok(t, flavor2.Delete())
}

func TestListFlavors(t *testing.T) {
	flavor := createFlavor(t)
	h.Ok(t, flavor.Save())

	flavors, err := models.ListFlavors()
	h.Ok(t, err)
	h.Equals(t, 1, len(flavors))
	checkFlavorValues(t, flavors[0])

	h.Ok(t, flavor.Delete())
}
