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
	router.HandleFunc(prefix, ListProjects).Methods("GET")
	router.HandleFunc(prefix, CreateProject).Methods("POST")
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/{projectID}", GetProject).Methods("GET")
	sub.HandleFunc("/{projectID}", UpdateProject).Methods("PATCH")
	sub.HandleFunc("/{projectID}", DeleteProject).Methods("DELETE")
	sub.HandleFunc("/{projectID}/users", GetProjectUsers).Methods("GET")
	sub.HandleFunc("/{projectID}/users", SetProjectUsers).Methods("PUT")
	sub.HandleFunc("/{projectID}/users/{userID}", AddProjectUser).Methods("PUT")
	sub.HandleFunc("/{projectID}/users/{userID}", RemoveProjectUser).Methods("DELETE")
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
