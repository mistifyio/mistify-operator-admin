package operator

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"code.google.com/p/go-uuid/uuid"

	"local/mistify-operator-admin/models"

	"github.com/gorilla/mux"
)

func RegisterFlavorRoutes(prefix string, router *mux.Router) {
	router.HandleFunc(prefix, ListFlavors).Methods("GET")
	router.HandleFunc(prefix, CreateFlavor).Methods("POST")
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/{flavorID}", GetFlavor).Methods("GET")
	sub.HandleFunc("/{flavorID}", UpdateFlavor).Methods("PATCH")
	sub.HandleFunc("/{flavorID}", DeleteFlavor).Methods("DELETE")
}

func ListFlavors(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	flavors, err := models.ListFlavors()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, flavors)
}

func GetFlavor(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	flavor, ok := getFlavorHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, flavor)
}

func CreateFlavor(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}

	// Parse Request
	flavor := &models.Flavor{}
	if err := flavor.FromJSON(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	// Assign an ID
	if flavor.ID != "" {
		hr.JSONMsg(http.StatusBadRequest, "id must not be defined")
		return
	}
	flavor.NewID()

	ok := saveFlavorHelper(hr, flavor)
	if !ok {
		return
	}
	hr.JSON(http.StatusCreated, flavor)
}

func UpdateFlavor(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	flavor, ok := getFlavorHelper(hr, r)
	if !ok {
		return // Specific response handled by getFlavorHelper
	}

	// Parse Request
	flavor := &models.Flavor{}
	if err := flavor.FromJSON(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	flavor.Apply(&flavorUpdates)
	ok = saveFlavorHelper(hr, flavor)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, flavor)
}

func DeleteFlavor(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	flavor, ok := getFlavorHelper(hr, r)
	if !ok {
		return
	}

	err := flavor.Delete()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, flavor)
}

func getFlavorHelper(hr HttpResponse, r *http.Request) (*models.Flavor, bool) {
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

func saveFlavorHelper(hr HttpResponse, flavor *models.Flavor) bool {
	err := flavor.Validate()
	if err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	err = flavor.Save()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
