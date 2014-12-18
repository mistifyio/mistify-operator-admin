package operator

import (
	"encoding/json"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-operator-admin/models"
)

// RegisterUserRoutes registers the user routes and handlers
func RegisterUserRoutes(prefix string, router *mux.Router) {
	router.HandleFunc(prefix, ListUsers).Methods("GET")
	router.HandleFunc(prefix, CreateUser).Methods("POST")
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/{userID}", GetUser).Methods("GET")
	sub.HandleFunc("/{userID}", UpdateUser).Methods("PATCH")
	sub.HandleFunc("/{userID}", DeleteUser).Methods("DELETE")
	sub.HandleFunc("/{userID}/projects", GetUserProjects).Methods("GET")
	sub.HandleFunc("/{userID}/projects", SetUserProjects).Methods("PUT")
	sub.HandleFunc("/{userID}/projects/{projectID}", AddUserProject).Methods("PUT")
	sub.HandleFunc("/{userID}/projects/{projectID}", RemoveUserProject).Methods("DELETE")
}

// ListUsers gets a list of all users
func ListUsers(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	users, err := models.ListUsers()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, users)
}

// GetUser gets a particular user
func GetUser(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, user)
}

// CreateUser creates a new user
func CreateUser(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}

	// Parse Request
	user := &models.User{}
	if err := user.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	// Assign an ID
	if user.ID != "" {
		hr.JSONMsg(http.StatusBadRequest, "id must not be defined")
		return
	}
	user.NewID()

	if !saveUserHelper(hr, user) {
		return
	}
	hr.JSON(http.StatusCreated, user)
}

// UpdateUser updates an existing user
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return // Specific response handled by getUserHelper
	}

	// Parse Request
	if err := user.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	if !saveUserHelper(hr, user) {
		return
	}
	hr.JSON(http.StatusOK, user)
}

// DeleteUser user deletes an existing user
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return
	}

	if err := user.Delete(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, user)
}

// GetUserProjects gets a list of projects associated with the user
func GetUserProjects(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return
	}
	if err := user.LoadProjects(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, user.Projects)
}

// SetUserProjects sets the list of projects associated with the user
func SetUserProjects(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	user, ok := getUserHelper(hr, r)
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

	if err := user.SetProjects(projects); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, user.Projects)
}

// AddUserProject associates a project with the user
func AddUserProject(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	projectID := vars["projectID"]

	if err := user.AddProject(&models.Project{ID: projectID}); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

// RemoveUserProject removes an association of a project with the user
func RemoveUserProject(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	projectID := vars["projectID"]

	if err := user.RemoveProject(&models.Project{ID: projectID}); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

// getUserHelper gets the user object and handles sending a response in case of
// error
func getUserHelper(hr HTTPResponse, r *http.Request) (*models.User, bool) {
	vars := mux.Vars(r)
	userID, ok := vars["userID"]
	if !ok {
		hr.JSONMsg(http.StatusBadRequest, "missing user id")
		return nil, false
	}
	if uuid.Parse(userID) == nil {
		hr.JSONMsg(http.StatusBadRequest, "invalid user id")
		return nil, false
	}
	user, err := models.FetchUser(userID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			hr.JSONMsg(http.StatusNotFound, "not found")
			return nil, false
		}
		hr.JSONError(http.StatusInternalServerError, err)
		return nil, false
	}
	return user, true
}

// saveUserHelper saves the user object and handles sending a response in case
// of error
func saveUserHelper(hr HTTPResponse, user *models.User) bool {
	if err := user.Validate(); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	if err := user.Save(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
