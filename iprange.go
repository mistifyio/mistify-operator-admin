package operator

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-operator-admin/models"
)

func RegisterIPRangeRoutes(prefix string, router *mux.Router) {
	router.HandleFunc(prefix, ListIPRanges).Methods("GET")
	router.HandleFunc(prefix, CreateIPRange).Methods("POST")
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/{iprangeID}", GetIPRange).Methods("GET")
	sub.HandleFunc("/{iprangeID}", UpdateIPRange).Methods("PATCH")
	sub.HandleFunc("/{iprangeID}", DeleteIPRange).Methods("DELETE")
	sub.HandleFunc("/{iprangeID}/hypervisors", GetIPRangeHypervisors).Methods("GET")
	sub.HandleFunc("/{iprangeID}/hypervisors", SetIPRangeHypervisors).Methods("PUT")
	sub.HandleFunc("/{iprangeID}/hypervisors/{hypervisorID}", AddIPRangeHypervisor).Methods("PUT")
	sub.HandleFunc("/{iprangeID}/hypervisors/{hypervisorID}", RemoveIPRangeHypervisor).Methods("DELETE")
	sub.HandleFunc("/{iprangeID}/network", GetIPRangeNetwork).Methods("GET")
	sub.HandleFunc("/{iprangeID}/network/{networkID}", SetIPRangeNetwork).Methods("PUT")
	sub.HandleFunc("/{iprangeID}/network/{networkID}", RemoveIPRangeNetwork).Methods("DELETE")
}

func ListIPRanges(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	ipranges, err := models.ListIPRanges()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, ipranges)
}

func GetIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, iprange)
}

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
