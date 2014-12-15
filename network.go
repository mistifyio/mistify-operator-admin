package operator

import (
	"database/sql"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-operator-admin/models"
)

func RegisterNetworkRoutes(prefix string, router *mux.Router) {
	router.HandleFunc(prefix, ListNetworks).Methods("GET")
	router.HandleFunc(prefix, CreateNetwork).Methods("POST")
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/{networkID}", GetNetwork).Methods("GET")
	sub.HandleFunc("/{networkID}", UpdateNetwork).Methods("PATCH")
	sub.HandleFunc("/{networkID}", DeleteNetwork).Methods("DELETE")
}

func ListNetworks(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	networks, err := models.ListNetworks()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, networks)
}

func GetNetwork(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	network, ok := getNetworkHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, network)
}

func CreateNetwork(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}

	// Parse Request
	network := &models.Network{}
	if err := network.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	// Assign an ID
	if network.ID != "" {
		hr.JSONMsg(http.StatusBadRequest, "id must not be defined")
		return
	}
	network.NewID()

	ok := saveNetworkHelper(hr, network)
	if !ok {
		return
	}
	hr.JSON(http.StatusCreated, network)
}

func UpdateNetwork(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	network, ok := getNetworkHelper(hr, r)
	if !ok {
		return // Specific response handled by getNetworkHelper
	}

	// Parse Request
	if err := network.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	ok = saveNetworkHelper(hr, network)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, network)
}

func DeleteNetwork(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	network, ok := getNetworkHelper(hr, r)
	if !ok {
		return
	}

	err := network.Delete()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, network)
}

func getNetworkHelper(hr HttpResponse, r *http.Request) (*models.Network, bool) {
	vars := mux.Vars(r)
	networkID, ok := vars["networkID"]
	if !ok {
		hr.JSONMsg(http.StatusBadRequest, "missing network id")
		return nil, false
	}
	if uuid.Parse(networkID) == nil {
		hr.JSONMsg(http.StatusBadRequest, "invalid network id")
		return nil, false
	}
	network, err := models.FetchNetwork(networkID)
	if err != nil {
		if err == sql.ErrNoRows {
			hr.JSONMsg(http.StatusNotFound, "not found")
			return nil, false
		}
		hr.JSONError(http.StatusInternalServerError, err)
		return nil, false
	}
	return network, true
}

func saveNetworkHelper(hr HttpResponse, network *models.Network) bool {
	err := network.Validate()
	if err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	err = network.Save()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
