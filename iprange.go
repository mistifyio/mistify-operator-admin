package operator

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-operator-admin/models"
)

// RegisterIPRangeRoutes registers the iprange routes and handlers
func RegisterIPRangeRoutes(prefix string, router *mux.Router) {
	RegisterOneRoute(router, RouteInfo{prefix, ListIPRanges, []string{"GET"}, "ipranges.list"})
	RegisterOneRoute(router, RouteInfo{prefix, CreateIPRange, []string{"POST"}, "ipranges.create"})
	sub := router.PathPrefix(prefix).Subrouter()
	RegisterOneRoute(sub, RouteInfo{"/{iprangeID}", GetIPRange, []string{"GET"}, "ipranges.get"})
	RegisterOneRoute(sub, RouteInfo{"/{iprangeID}", UpdateIPRange, []string{"PATCH"}, "ipranges.update"})
	RegisterOneRoute(sub, RouteInfo{"/{iprangeID}", DeleteIPRange, []string{"DELETE"}, "ipranges.delete"})
	RegisterOneRoute(sub, RouteInfo{"/{iprangeID}/hypervisors", GetIPRangeHypervisors, []string{"GET"}, "ipranges.hypervisors.get"})
	RegisterOneRoute(sub, RouteInfo{"/{iprangeID}/hypervisors", SetIPRangeHypervisors, []string{"PUT"}, "ipranges.hypervisors.set"})
	RegisterOneRoute(sub, RouteInfo{"/{iprangeID}/hypervisors/{hypervisorID}", AddIPRangeHypervisor, []string{"PUT"}, "ipranges.hypervisors.add"})
	RegisterOneRoute(sub, RouteInfo{"/{iprangeID}/hypervisors/{hypervisorID}", RemoveIPRangeHypervisor, []string{"DELETE"}, "ipranges.hypervisors.remove"})
	RegisterOneRoute(sub, RouteInfo{"/{iprangeID}/network", GetIPRangeNetwork, []string{"GET"}, "ipranges.hypervisors.getnetwork"})
	RegisterOneRoute(sub, RouteInfo{"/{iprangeID}/network/{networkID}", SetIPRangeNetwork, []string{"PUT"}, "ipranges.hypervisors.setnetwork"})
	RegisterOneRoute(sub, RouteInfo{"/{iprangeID}/network/{networkID}", RemoveIPRangeNetwork, []string{"DELETE"}, "ipranges.hypervisors.removenetwork"})
}

// ListIPRanges gets a list of all ipranges
func ListIPRanges(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	ipranges, err := models.ListIPRanges()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, ipranges)
}

// GetIPRange gets a particular iprange
func GetIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, iprange)
}

// CreateIPRange creates a new iprange
func CreateIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}

	// Parse Request
	iprange := &models.IPRange{}
	if err := iprange.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	// Assign an ID
	if iprange.ID != "" {
		hr.JSONMsg(http.StatusBadRequest, "id must not be defined")
		return
	}
	iprange.NewID()

	if !saveIPRangeHelper(hr, iprange) {
		return
	}
	hr.JSON(http.StatusCreated, iprange)
}

// UpdateIPRange updates an existing iprange
func UpdateIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return // Specific response handled by getIPRangeHelper
	}

	// Parse Request
	if err := iprange.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	if !saveIPRangeHelper(hr, iprange) {
		return
	}
	hr.JSON(http.StatusOK, iprange)
}

// DeleteIPRange deletes an existing iprange
func DeleteIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}

	if err := iprange.Delete(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, iprange)
}

// GetIPRangeHypervisors gets a list of hypervisors associated with the iprange
func GetIPRangeHypervisors(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}
	if err := iprange.LoadHypervisors(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, iprange.Hypervisors)
}

// SetIPRangeHypervisors sets the list of hypervisors associated with the
// iprange
func SetIPRangeHypervisors(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}
	var hypervisorIDs []string
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&hypervisorIDs); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	hypervisors := make([]*models.Hypervisor, len(hypervisorIDs))
	for i, v := range hypervisorIDs {
		hypervisors[i] = &models.Hypervisor{ID: v}
	}

	if err := iprange.SetHypervisors(hypervisors); err != nil {
		hr.JSONMsg(http.StatusInternalServerError, err.Error())
		return
	}

	hr.JSON(http.StatusOK, iprange.Hypervisors)
}

// AddIPRangeHypervisor associates a hypervisor with the iprange
func AddIPRangeHypervisor(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	hypervisorID, ok := vars["hypervisorID"]

	if err := iprange.AddHypervisor(&models.Hypervisor{ID: hypervisorID}); err != nil {
		hr.JSONMsg(http.StatusInternalServerError, err.Error())
		return
	}
	hr.JSON(http.StatusCreated, &struct{}{})
}

// RemoveIPRangeHypervisor removes an association of a hypervisor with the
// iprange
func RemoveIPRangeHypervisor(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	hypervisorID, ok := vars["hypervisorID"]

	if err := iprange.RemoveHypervisor(&models.Hypervisor{ID: hypervisorID}); err != nil {
		hr.JSONMsg(http.StatusInternalServerError, err.Error())
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

// GetIPRangeNetwork retrieves the network associated with the iprange
func GetIPRangeNetwork(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}
	if err := iprange.LoadNetwork(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, iprange.Network)
}

// SetIPRangeNetwork sets the network associated with the iprange
func SetIPRangeNetwork(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	networkID, ok := vars["networkID"]

	if err := iprange.SetNetwork(&models.Network{ID: networkID}); err != nil {
		hr.JSONMsg(http.StatusInternalServerError, err.Error())
		return
	}
	hr.JSON(http.StatusCreated, &struct{}{})
}

// RemoveIPRangeNetwork unsets the network associated with the iprange
func RemoveIPRangeNetwork(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	networkID, ok := vars["networkID"]

	if err := iprange.RemoveNetwork(&models.Network{ID: networkID}); err != nil {
		hr.JSONMsg(http.StatusInternalServerError, err.Error())
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

// getIPRangeHelper gets the iprange object and handles sending a response in
// case of error
func getIPRangeHelper(hr HTTPResponse, r *http.Request) (*models.IPRange, bool) {
	vars := mux.Vars(r)
	iprangeID, ok := vars["iprangeID"]
	if !ok {
		hr.JSONMsg(http.StatusBadRequest, "missing iprange id")
		return nil, false
	}
	if uuid.Parse(iprangeID) == nil {
		hr.JSONMsg(http.StatusBadRequest, "invalid iprange id")
		return nil, false
	}
	iprange, err := models.FetchIPRange(iprangeID)
	if err != nil {
		if err == sql.ErrNoRows {
			hr.JSONMsg(http.StatusNotFound, "not found")
			return nil, false
		}
		hr.JSONError(http.StatusInternalServerError, err)
		return nil, false
	}
	return iprange, true
}

// saveIPRangeHelper saves the hypervisor object and handles sending a response
// in case of error
func saveIPRangeHelper(hr HTTPResponse, iprange *models.IPRange) bool {
	if err := iprange.Validate(); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	if err := iprange.Save(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
