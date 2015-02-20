package operator

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-operator-admin/models"
)

// RegisterNetworkRoutes registers the network routes and handlers
func RegisterNetworkRoutes(prefix string, router *mux.Router) {
	RegisterOneRoute(router, RouteInfo{prefix, ListNetworks, []string{"GET"}, "networks.list"})
	RegisterOneRoute(router, RouteInfo{prefix, CreateNetwork, []string{"POST"}, "networks.create"})
	sub := router.PathPrefix(prefix).Subrouter()
	RegisterOneRoute(sub, RouteInfo{"/{networkID}", GetNetwork, []string{"GET"}, "networks.get"})
	RegisterOneRoute(sub, RouteInfo{"/{networkID}", UpdateNetwork, []string{"PATCH"}, "networks.update"})
	RegisterOneRoute(sub, RouteInfo{"/{networkID}", DeleteNetwork, []string{"DELETE"}, "networks.delete"})
	RegisterOneRoute(sub, RouteInfo{"/{networkID}/ipranges", GetNetworkIPRanges, []string{"GET"}, "networks.ipranges.get"})
	RegisterOneRoute(sub, RouteInfo{"/{networkID}/ipranges", SetNetworkIPRanges, []string{"PUT"}, "networks.ipranges.set"})
	RegisterOneRoute(sub, RouteInfo{"/{networkID}/ipranges/{iprangeID}", AddNetworkIPRange, []string{"PUT"}, "networks.ipranges.add"})
	RegisterOneRoute(sub, RouteInfo{"/{networkID}/ipranges/{iprangeID}", RemoveNetworkIPRange, []string{"DELETE"}, "networks.ipranges.remove"})
}

// ListNetworks gets a list of all networks
func ListNetworks(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	networks, err := models.ListNetworks()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, networks)
}

// GetNetwork gets a particular network
func GetNetwork(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	network, ok := getNetworkHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, network)
}

// CreateNetwork creates a new network
func CreateNetwork(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}

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

	if !saveNetworkHelper(hr, network) {
		return
	}
	hr.JSON(http.StatusCreated, network)
}

// UpdateNetwork updates an existing network
func UpdateNetwork(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	network, ok := getNetworkHelper(hr, r)
	if !ok {
		return // Specific response handled by getNetworkHelper
	}

	// Parse Request
	if err := network.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	if !saveNetworkHelper(hr, network) {
		return
	}
	hr.JSON(http.StatusOK, network)
}

// DeleteNetwork deletes an existing network
func DeleteNetwork(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	network, ok := getNetworkHelper(hr, r)
	if !ok {
		return
	}

	if err := network.Delete(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, network)
}

// GetNetworkIPRanges gets a list of ipranges associated with the network
func GetNetworkIPRanges(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	network, ok := getNetworkHelper(hr, r)
	if !ok {
		return
	}
	if err := network.LoadIPRanges(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, network.IPRanges)
}

// SetNetworkIPRanges sets the list of ipranges associated with the network
func SetNetworkIPRanges(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	network, ok := getNetworkHelper(hr, r)
	if !ok {
		return
	}
	var iprangeIDs []string
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&iprangeIDs); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	ipranges := make([]*models.IPRange, len(iprangeIDs))
	for i, v := range iprangeIDs {
		ipranges[i] = &models.IPRange{ID: v}
	}

	if err := network.SetIPRanges(ipranges); err != nil {
		hr.JSONMsg(http.StatusInternalServerError, err.Error())
		return
	}

	hr.JSON(http.StatusOK, network.IPRanges)
}

// AddNetworkIPRange associates an iprange with the network
func AddNetworkIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	network, ok := getNetworkHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	iprangeID, ok := vars["iprangeID"]

	if err := network.AddIPRange(&models.IPRange{ID: iprangeID}); err != nil {
		hr.JSONMsg(http.StatusInternalServerError, err.Error())
		return
	}
	hr.JSON(http.StatusCreated, &struct{}{})
}

// RemoveNetworkIPRange removes an association of an iprange with the network
func RemoveNetworkIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	network, ok := getNetworkHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	iprangeID, ok := vars["iprangeID"]

	if err := network.RemoveIPRange(&models.IPRange{ID: iprangeID}); err != nil {
		hr.JSONMsg(http.StatusInternalServerError, err.Error())
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

// getNetworkHelper gets the network object and handles sending a response in
// case of error
func getNetworkHelper(hr HTTPResponse, r *http.Request) (*models.Network, bool) {
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

// saveNetworkHelper saves the netework object and handles sending a response in
// case of error
func saveNetworkHelper(hr HTTPResponse, network *models.Network) bool {
	if err := network.Validate(); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	if err := network.Save(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
