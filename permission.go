package operator

import (
	"encoding/json"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-operator-admin/models"
)

// RegisterPermissionRoutes registers the permission routes and handlers
func RegisterPermissionRoutes(prefix string, router *mux.Router) {
	router.HandleFunc(prefix, ListPermissions).Methods("GET")
	router.HandleFunc(prefix, CreatePermission).Methods("POST")
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/{permissionID}", GetPermission).Methods("GET")
	sub.HandleFunc("/{permissionID}", UpdatePermission).Methods("PATCH")
	sub.HandleFunc("/{permissionID}", DeletePermission).Methods("DELETE")
	sub.HandleFunc("/{permissionID}/projects", GetPermissionProjects).Methods("GET")
	sub.HandleFunc("/{permissionID}/projects", SetPermissionProjects).Methods("PUT")
	sub.HandleFunc("/{permissionID}/projects/{projectID}", AddPermissionProject).Methods("PUT")
	sub.HandleFunc("/{permissionID}/projects/{projectID}", RemovePermissionProject).Methods("DELETE")
}

// ListPermissions gets a list of all permissions
func ListPermissions(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	permissions, err := models.ListPermissions()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, permissions)
}

// GetPermission gets a particular permission
func GetPermission(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	permission, ok := getPermissionHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, permission)
}

// CreatePermission creates a new permission
func CreatePermission(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}

	// Parse Request
	permission := &models.Permission{}
	if err := permission.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	// Assign an ID
	if permission.ID != "" {
		hr.JSONMsg(http.StatusBadRequest, "id must not be defined")
		return
	}
	permission.NewID()

	if !savePermissionHelper(hr, permission) {
		return
	}
	hr.JSON(http.StatusCreated, permission)
}

// UpdatePermission updates an existing permission
func UpdatePermission(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	permission, ok := getPermissionHelper(hr, r)
	if !ok {
		return // Specific response handled by getPermissionHelper
	}

	// Parse Request
	if err := permission.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	if !savePermissionHelper(hr, permission) {
		return
	}
	hr.JSON(http.StatusOK, permission)
}

// DeletePermission permission deletes an existing permission
func DeletePermission(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	permission, ok := getPermissionHelper(hr, r)
	if !ok {
		return
	}

	if err := permission.Delete(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, permission)
}

// GetPermissionProjects gets a list of projects associated with the permission
func GetPermissionProjects(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	permission, ok := getPermissionHelper(hr, r)
	if !ok {
		return
	}
	if err := permission.LoadProjects(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, permission.Projects)
}

// SetPermissionProjects sets the list of projects associated with the permission
func SetPermissionProjects(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	permission, ok := getPermissionHelper(hr, r)
	if !ok {
		return
	}

	var projectIDs []string
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&projectIDs); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	projects := make([]*models.Project, len(projectIDs))
	for i, v := range projectIDs {
		projects[i] = &models.Project{ID: v}
	}

	if err := permission.SetProjects(projects); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, permission.Projects)
}

// AddPermissionProject associates a project with the permission
func AddPermissionProject(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	permission, ok := getPermissionHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	projectID := vars["projectID"]

	if err := permission.AddProject(&models.Project{ID: projectID}); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

// RemovePermissionProject removes an association of a project with the permission
func RemovePermissionProject(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	permission, ok := getPermissionHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	projectID := vars["projectID"]

	if err := permission.RemoveProject(&models.Project{ID: projectID}); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

// getPermissionHelper gets the permission object and handles sending a response in case of
// error
func getPermissionHelper(hr HTTPResponse, r *http.Request) (*models.Permission, bool) {
	vars := mux.Vars(r)
	permissionID, ok := vars["permissionID"]
	if !ok {
		hr.JSONMsg(http.StatusBadRequest, "missing permission id")
		return nil, false
	}
	if uuid.Parse(permissionID) == nil {
		hr.JSONMsg(http.StatusBadRequest, "invalid permission id")
		return nil, false
	}
	permission, err := models.FetchPermission(permissionID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			hr.JSONMsg(http.StatusNotFound, "not found")
			return nil, false
		}
		hr.JSONError(http.StatusInternalServerError, err)
		return nil, false
	}
	return permission, true
}

// savePermissionHelper saves the permission object and handles sending a response in case
// of error
func savePermissionHelper(hr HTTPResponse, permission *models.Permission) bool {
	if err := permission.Validate(); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	if err := permission.Save(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
