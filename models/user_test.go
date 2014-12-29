package models_test

import (
	"strings"
	"testing"

	"code.google.com/p/go-uuid/uuid"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/models"
)

var userJSON = `{
	"id": "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d",
	"username": "foobar",
	"email": "foo@bar.com",
	"metadata": {
		"foo": "bar"
	}
}`

func createUser(t *testing.T) *models.User {
	r := strings.NewReader(userJSON)
	user := &models.User{}
	h.Ok(t, user.Decode(r))
	return user
}

func checkUserValues(t *testing.T, user *models.User) {
	h.Equals(t, "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d", user.ID)
	h.Equals(t, "foobar", user.Username)
	h.Equals(t, "foo@bar.com", user.Email)
	h.Equals(t, map[string]string{"foo": "bar"}, user.Metadata)
}

func TestNewUser(t *testing.T) {
	user := models.NewUser()
	h.Assert(t, uuid.Parse(user.ID) != nil, "missing uuid ID")
	h.Assert(t, user.Metadata != nil, "uninitialized metadata")
}

func TestUserNewID(t *testing.T) {
	user := models.NewUser()
	id1 := user.ID
	user.NewID()
	h.Assert(t, uuid.Parse(user.ID) != nil, "missing uuid ID")
	h.Assert(t, id1 != user.ID, "New ID was not generated")
}

func TestUserDecode(t *testing.T) {
	user := createUser(t)
	checkUserValues(t, user)
}

func TestUserSave(t *testing.T) {
	user := createUser(t)
	h.Ok(t, user.Save())
}

func TestUserDelete(t *testing.T) {
	user := createUser(t)
	h.Ok(t, user.Delete())
}

func TestUserLoad(t *testing.T) {
	user := createUser(t)
	h.Ok(t, user.Save())

	user2 := models.NewUser()
	user2.ID = user.ID

	h.Ok(t, user2.Load())
	checkUserValues(t, user2)
	h.Ok(t, user2.Delete())
}

func TestFetchUser(t *testing.T) {
	user := createUser(t)
	h.Ok(t, user.Save())

	user2, err := models.FetchUser("ebf3bfd5-9915-4ed1-bcb3-117bb48b155d")
	h.Ok(t, err)
	checkUserValues(t, user2)
	h.Ok(t, user2.Delete())
}

func TestListUsers(t *testing.T) {
	user := createUser(t)
	h.Ok(t, user.Save())

	users, err := models.ListUsers()
	h.Ok(t, err)
	h.Equals(t, 1, len(users))
	checkUserValues(t, users[0])
	h.Ok(t, user.Delete())
}
