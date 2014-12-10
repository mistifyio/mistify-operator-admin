package operator

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"local/mistify-operator-admin/models"
)

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

func ListProjects(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	projects, err := models.ListProjects()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, projects)
}

func GetProject(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, project)
}

func CreateProject(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}

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

	ok := saveProjectHelper(hr, project)
	if !ok {
		return
	}
	hr.JSON(http.StatusCreated, project)
}

func UpdateProject(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return // Specific response handled by getProjectHelper
	}

	if err := project.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}
	ok = saveProjectHelper(hr, project)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, project)
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}

	err := project.Delete()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, project)
}

func GetProjectUsers(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}
	err := project.LoadUsers()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, project.Users)
}

func SetProjectUsers(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}

	var userIDs []*string
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&userIDs)
	if err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	err = project.SetUsers(userIDs)
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, project.Users)
}

func AddProjectUser(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	userID := vars["userID"]

	err := project.AddUser(userID)
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, project.Users)
}

func RemoveProjectUser(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	project, ok := getProjectHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	userID := vars["userID"]

	err := project.RemoveUser(userID)
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, project.Users)
}

func getProjectHelper(hr HttpResponse, r *http.Request) (*models.Project, bool) {
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

func saveProjectHelper(hr HttpResponse, project *models.Project) bool {
	err := project.Validate()
	if err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	err = project.Save()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
