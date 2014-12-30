package models_test

import (
	"strings"
	"testing"

	"code.google.com/p/go-uuid/uuid"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/models"
)

var permissionJSON = `{
	"id": "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d",
	"Name": "foobar",
	"Service": "baz",
	"Action": "qwerty",
	"EntityType": "asdf",
	"Owner": true,
	"Description": "stuff",
	"metadata": {
		"foo": "bar"
	}
}`

func createPermission(t *testing.T) *models.Permission {
	r := strings.NewReader(permissionJSON)
	permission := &models.Permission{}
	h.Ok(t, permission.Decode(r))
	return permission
}

func checkPermissionValues(t *testing.T, permission *models.Permission) {
	h.Equals(t, "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d", permission.ID)
	h.Equals(t, "foobar", permission.Name)
	h.Equals(t, "baz", permission.Service)
	h.Equals(t, "qwerty", permission.Action)
	h.Equals(t, "asdf", permission.EntityType)
	h.Equals(t, true, permission.Owner)
	h.Equals(t, "stuff", permission.Description)
	h.Equals(t, map[string]string{"foo": "bar"}, permission.Metadata)
}

func TestNewPermission(t *testing.T) {
	permission := models.NewPermission()
	h.Assert(t, uuid.Parse(permission.ID) != nil, "missing uuid ID")
	h.Assert(t, permission.Metadata != nil, "uninitialized metadata")
}

func TestPermissionNewID(t *testing.T) {
	permission := models.NewPermission()
	id1 := permission.ID
	permission.NewID()
	h.Assert(t, uuid.Parse(permission.ID) != nil, "missing uuid ID")
	h.Assert(t, id1 != permission.ID, "New ID was not generated")
}

func TestPermissionDecode(t *testing.T) {
	permission := createPermission(t)
	checkPermissionValues(t, permission)
}

func TestPermissionValidate(t *testing.T) {
	permission := &models.Permission{}
	var err error

	err = permission.Validate()
	h.Assert(t, errContains(models.ErrNoID, err), "expected ErrNoID")
	h.Assert(t, errContains(models.ErrBadID, err), "expected ErrBadID")
	h.Assert(t, errContains(models.ErrNoService, err), "expected ErrNoService")
	h.Assert(t, errContains(models.ErrNoAction, err), "expected ErrNoAction")
	h.Assert(t, errContains(models.ErrNilMetadata, err), "expected ErrNilMetadata")

	permission.ID = "foobar"
	err = permission.Validate()
	h.Assert(t, errDoesNotContain(models.ErrNoID, err), "did not expect ErrNoID")
	h.Assert(t, errContains(models.ErrBadID, err), "expected ErrBadID")

	permission.NewID()
	h.Assert(t, errDoesNotContain(models.ErrBadID, permission.Validate()), "did not expect ErrBadID")

	permission.Service = "foobar"
	h.Assert(t, errDoesNotContain(models.ErrNoService, permission.Validate()), "did not expect ErrNoService")

	permission.Action = "foobar"
	h.Assert(t, errDoesNotContain(models.ErrNoAction, permission.Validate()), "did not expect ErrNoAction")

	permission.Metadata = make(map[string]string)
	err = permission.Validate()
	h.Assert(t, errDoesNotContain(models.ErrNilMetadata, err), "did not expect ErrNilMetadata")

	h.Ok(t, err)
}

func TestPermissionSave(t *testing.T) {
	permission := createPermission(t)
	h.Ok(t, permission.Save())
}

func TestPermissionDelete(t *testing.T) {
	permission := createPermission(t)
	h.Ok(t, permission.Delete())
}

func TestPermissionLoad(t *testing.T) {
	permission := createPermission(t)
	h.Ok(t, permission.Save())

	permission2 := models.NewPermission()
	permission2.ID = permission.ID

	h.Ok(t, permission2.Load())
	checkPermissionValues(t, permission2)
	h.Ok(t, permission2.Delete())
}

func TestFetchPermission(t *testing.T) {
	permission := createPermission(t)
	h.Ok(t, permission.Save())

	permission2, err := models.FetchPermission("ebf3bfd5-9915-4ed1-bcb3-117bb48b155d")
	h.Ok(t, err)
	checkPermissionValues(t, permission2)
	h.Ok(t, permission2.Delete())
}

func TestListPermissions(t *testing.T) {
	permission := createPermission(t)
	h.Ok(t, permission.Save())

	permissions, err := models.ListPermissions()
	h.Ok(t, err)
	h.Equals(t, 1, len(permissions))
	checkPermissionValues(t, permissions[0])
	h.Ok(t, permission.Delete())
}

func TestPermissionProjectRelations(t *testing.T) {
	// Prep
	permission := createPermission(t)
	h.Ok(t, permission.Save())
	project := createProject(t)
	h.Ok(t, project.Save())

	// Add
	h.Ok(t, permission.AddProject(project))

	// Load
	h.Ok(t, permission.LoadProjects())
	h.Equals(t, 1, len(permission.Projects))

	// Remove
	h.Ok(t, permission.RemoveProject(project))
	h.Ok(t, permission.LoadProjects())
	h.Equals(t, 0, len(permission.Projects))

	// Set
	h.Ok(t, permission.SetProjects([]*models.Project{project}))
	h.Ok(t, permission.LoadProjects())
	h.Equals(t, 1, len(permission.Projects))

	// Lookup permissions by project
	permissions, err := models.PermissionsByProject(project)
	h.Ok(t, err)
	h.Equals(t, 1, len(permissions))

	// Clear
	h.Ok(t, permission.SetProjects(make([]*models.Project, 0)))
	h.Ok(t, permission.LoadProjects())
	h.Equals(t, 0, len(permission.Projects))

	// Cleanup
	h.Ok(t, project.Delete())
	h.Ok(t, permission.Delete())
}
