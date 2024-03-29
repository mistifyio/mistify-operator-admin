package operator

import (
	"encoding/json"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-operator-admin/models"
)

// RegisterProjectRoutes registers the project routes and handlers
func RegisterProjectRoutes(prefix string, router *mux.Router) {
	RegisterOneRoute(router, RouteInfo{prefix, ListProjects, []string{"GET"}, "projects.list"})
	RegisterOneRoute(router, RouteInfo{prefix, CreateProject, []string{"POST"}, "projects.create"})
	sub := router.PathPrefix(prefix).Subrouter()
	RegisterOneRoute(sub, RouteInfo{"/{projectID}", GetProject, []string{"GET"}, "projects.get"})
	RegisterOneRoute(sub, RouteInfo{"/{projectID}", UpdateProject, []string{"PATCH"}, "projects.update"})
	RegisterOneRoute(sub, RouteInfo{"/{projectID}", DeleteProject, []string{"DELETE"}, "projects.delete"})
	RegisterOneRoute(sub, RouteInfo{"/{projectID}/users", GetProjectUsers, []string{"GET"}, "projects.users.get"})
	RegisterOneRoute(sub, RouteInfo{"/{projectID}/users", SetProjectUsers, []string{"PUT"}, "projects.users.set"})
	RegisterOneRoute(sub, RouteInfo{"/{projectID}/users/{userID}", AddProjectUser, []string{"PUT"}, "projects.users.add"})
	RegisterOneRoute(sub, RouteInfo{"/{projectID}/users/{userID}", RemoveProjectUser, []string{"DELETE"}, "projects.users.remove"})
	RegisterOneRoute(sub, RouteInfo{"/{projectID}/permissions", GetProjectPermissions, []string{"GET"}, "projects.permissions.get"})
	RegisterOneRoute(sub, RouteInfo{"/{projectID}/permissions", SetProjectPermissions, []string{"PUT"}, "projects.permissions.set"})
	RegisterOneRoute(sub, RouteInfo{"/{projectID}/permissions/{permissionID}", AddProjectPermission, []string{"PUT"}, "projects.permissions.add"})
	RegisterOneRoute(sub, RouteInfo{"/{projectID}/permissions/{permissionID}", RemoveProjectPermission, []string{"DELETE"}, "projects.permissions.remove"})
}

// ListProjects gets a list of all projects
func ListProjects(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	projects, err := models.ListProjects()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, projects)
}

// GetProject gets a particular project
func GetProject(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, project)
}

// CreateProject creates a new project
func CreateProject(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}

	// Parse Request
	project := &models.Project{}
	if err := project.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	// Assign an ID
	if project.ID != "" {
		hr.JSONMsg(http.StatusBadRequest, "id must not be defined")
		return
	}
	project.NewID()

	if !saveProjectHelper(hr, project) {
		return
	}
	hr.JSON(http.StatusCreated, project)
}

// UpdateProject updates an existing project
func UpdateProject(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return // Specific response handled by getProjectHelper
	}

	if err := project.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}
	if !saveProjectHelper(hr, project) {
		return
	}
	hr.JSON(http.StatusOK, project)
}

// DeleteProject deletes an existing project
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}

	if err := project.Delete(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, project)
}

// GetProjectUsers gets a list of users associated with the project
func GetProjectUsers(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}
	if err := project.LoadUsers(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, project.Users)
}

// SetProjectUsers sets teh list of users associated with the project
func SetProjectUsers(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}

	var userIDs []string
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userIDs); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	users := make([]*models.User, len(userIDs))
	for i, v := range userIDs {
		users[i] = &models.User{ID: v}
	}

	if err := project.SetUsers(users); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, project.Users)
}

// AddProjectUser associates a user with the project
func AddProjectUser(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	userID := vars["userID"]

	if err := project.AddUser(&models.User{ID: userID}); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

// RemoveProjectUser disassociates a user with the project
func RemoveProjectUser(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	userID := vars["userID"]

	if err := project.RemoveUser(&models.User{ID: userID}); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

// GetProjectPermissions gets a list of permissions associated with the project
func GetProjectPermissions(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}
	if err := project.LoadPermissions(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, project.Permissions)
}

// SetProjectPermissions sets teh list of permissions associated with the project
func SetProjectPermissions(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}

	var permissionIDs []string
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&permissionIDs); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	permissions := make([]*models.Permission, len(permissionIDs))
	for i, v := range permissionIDs {
		permissions[i] = &models.Permission{ID: v}
	}

	if err := project.SetPermissions(permissions); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, project.Permissions)
}

// AddProjectPermission associates a permission with the project
func AddProjectPermission(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	permissionID := vars["permissionID"]

	if err := project.AddPermission(&models.Permission{ID: permissionID}); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

// RemoveProjectPermission disassociates a permission with the project
func RemoveProjectPermission(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	permissionID := vars["permissionID"]

	if err := project.RemovePermission(&models.Permission{ID: permissionID}); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

// getProjectHelper gets the project object and handles sending a response in
// case of error
func getProjectHelper(hr HTTPResponse, r *http.Request) (*models.Project, bool) {
	vars := mux.Vars(r)
	projectID, ok := vars["projectID"]
	if !ok {
		hr.JSONMsg(http.StatusBadRequest, "missing project id")
		return nil, false
	}
	if uuid.Parse(projectID) == nil {
		hr.JSONMsg(http.StatusBadRequest, "invalid project id")
		return nil, false
	}
	project, err := models.FetchProject(projectID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			hr.JSONMsg(http.StatusNotFound, "not found")
			return nil, false
		}
		hr.JSONError(http.StatusInternalServerError, err)
		return nil, false
	}
	return project, true
}

// saveProjectHelper saves the project object and handles sending a response in
// case of error
func saveProjectHelper(hr HTTPResponse, project *models.Project) bool {
	if err := project.Validate(); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	if err := project.Save(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
