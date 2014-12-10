package operator

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"local/mistify-operator-admin/models"

	"code.google.com/p/go-uuid/uuid"
)

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

func ListUsers(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	users, err := models.ListUsers()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, users)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, user)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}

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

	ok := saveUserHelper(hr, user)
	if !ok {
		return
	}
	hr.JSON(http.StatusCreated, user)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return // Specific response handled by getUserHelper
	}

	// Parse Request
	if err := user.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	ok = saveUserHelper(hr, user)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, user)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return
	}

	err := user.Delete()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
	}
	hr.JSON(http.StatusOK, user)
}

func GetUserProjects(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return
	}
	err := user.LoadProjects()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, user.Projects)
}

func SetUserProjects(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return
	}

	var projectIDs []*string
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&projectIDs)
	if err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	err = user.SetProjects(projectIDs)
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, user.Projects)
}

func AddUserProject(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	projectID := vars["projectID"]

	err := user.AddProject(projectID)
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, user.Projects)
}

func RemoveUserProject(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	user, ok := getUserHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	projectID := vars["projectID"]

	err := user.RemoveProject(projectID)
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, user.Projects)
}

func getUserHelper(hr HttpResponse, r *http.Request) (*models.User, bool) {
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

func saveUserHelper(hr HttpResponse, user *models.User) bool {
	err := user.Validate()
	if err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	err = user.Save()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
