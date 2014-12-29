package models_test

import (
	"strings"
	"testing"

	"code.google.com/p/go-uuid/uuid"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-operator-admin/models"
)

var projectJSON = `{
	"id": "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d",
	"name": "foobar",
	"metadata": {
		"foo": "bar"
	}
}`

func createProject(t *testing.T) *models.Project {
	r := strings.NewReader(projectJSON)
	project := &models.Project{}
	h.Ok(t, project.Decode(r))
	return project
}

func checkProjectValues(t *testing.T, project *models.Project) {
	h.Equals(t, "ebf3bfd5-9915-4ed1-bcb3-117bb48b155d", project.ID)
	h.Equals(t, "foobar", project.Name)
	h.Equals(t, map[string]string{"foo": "bar"}, project.Metadata)
}

func TestNewProject(t *testing.T) {
	project := models.NewProject()
	h.Assert(t, uuid.Parse(project.ID) != nil, "missing uuid ID")
	h.Assert(t, project.Metadata != nil, "uninitialized metadata")
}

func TestProjectNewID(t *testing.T) {
	project := models.NewProject()
	id1 := project.ID
	project.NewID()
	h.Assert(t, uuid.Parse(project.ID) != nil, "missing uuid ID")
	h.Assert(t, id1 != project.ID, "New ID was not generated")
}

func TestProjectDecode(t *testing.T) {
	project := createProject(t)
	checkProjectValues(t, project)
}

func TestProjectSave(t *testing.T) {
	project := createProject(t)
	h.Ok(t, project.Save())
}

func TestProjectDelete(t *testing.T) {
	project := createProject(t)
	h.Ok(t, project.Delete())
}

func TestProjectLoad(t *testing.T) {
	project := createProject(t)
	h.Ok(t, project.Save())

	project2 := models.NewProject()
	project2.ID = project.ID

	h.Ok(t, project2.Load())
	checkProjectValues(t, project2)
	h.Ok(t, project2.Delete())
}

func TestFetchProject(t *testing.T) {
	project := createProject(t)
	h.Ok(t, project.Save())

	project2, err := models.FetchProject("ebf3bfd5-9915-4ed1-bcb3-117bb48b155d")
	h.Ok(t, err)
	checkProjectValues(t, project2)
	h.Ok(t, project2.Delete())
}

func TestListProjects(t *testing.T) {
	project := createProject(t)
	h.Ok(t, project.Save())

	projects, err := models.ListProjects()
	h.Ok(t, err)
	h.Equals(t, 1, len(projects))
	checkProjectValues(t, projects[0])
	h.Ok(t, project.Delete())
}
