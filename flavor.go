package operator

import (
	"database/sql"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-operator-admin/models"
)

// RegisterFlavorRoutes registers the flavor routes and handlers
func RegisterFlavorRoutes(prefix string, router *mux.Router) {
	router.HandleFunc(prefix, ListFlavors).Methods("GET")
	router.HandleFunc(prefix, CreateFlavor).Methods("POST")
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/{flavorID}", GetFlavor).Methods("GET")
	sub.HandleFunc("/{flavorID}", UpdateFlavor).Methods("PATCH")
	sub.HandleFunc("/{flavorID}", DeleteFlavor).Methods("DELETE")
}

// ListFlavors get a list of all flavors
func ListFlavors(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	flavors, err := models.ListFlavors()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, flavors)
}

// GetFlavor gets a paritcular flavor
func GetFlavor(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	flavor, ok := getFlavorHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, flavor)
}

// CreateFlavor creates a new flavor
func CreateFlavor(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}

	// Parse Request
	flavor := &models.Flavor{}
	if err := flavor.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	// Assign an ID
	if flavor.ID != "" {
		hr.JSONMsg(http.StatusBadRequest, "id must not be defined")
		return
	}
	flavor.NewID()

	if !saveFlavorHelper(hr, flavor) {
		return
	}
	hr.JSON(http.StatusCreated, flavor)
}

// UpdateFlavor updates an existing flavor
func UpdateFlavor(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	flavor, ok := getFlavorHelper(hr, r)
	if !ok {
		return // Specific response handled by getFlavorHelper
	}

	// Parse Request
	if err := flavor.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	if !saveFlavorHelper(hr, flavor) {
		return
	}
	hr.JSON(http.StatusOK, flavor)
}

// DeleteFlavor deletes an existing flavor
func DeleteFlavor(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	flavor, ok := getFlavorHelper(hr, r)
	if !ok {
		return
	}

	if err := flavor.Delete(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, flavor)
}

// getFlavorHelper gets the flavor object and handles sending a response in case
// of error
func getFlavorHelper(hr HTTPResponse, r *http.Request) (*models.Flavor, bool) {
	vars := mux.Vars(r)
	flavorID, ok := vars["flavorID"]
	if !ok {
		hr.JSONMsg(http.StatusBadRequest, "missing flavor id")
		return nil, false
	}
	if uuid.Parse(flavorID) == nil {
		hr.JSONMsg(http.StatusBadRequest, "invalid flavor id")
		return nil, false
	}
	flavor, err := models.FetchFlavor(flavorID)
	if err != nil {
		if err == sql.ErrNoRows {
			hr.JSONMsg(http.StatusNotFound, "not found")
			return nil, false
		}
		hr.JSONError(http.StatusInternalServerError, err)
		return nil, false
	}
	return flavor, true
}

// saveFlavorHelper saves the flavor object and handles sending a response in
// case of error
func saveFlavorHelper(hr HTTPResponse, flavor *models.Flavor) bool {
	if err := flavor.Validate(); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	if err := flavor.Save(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
