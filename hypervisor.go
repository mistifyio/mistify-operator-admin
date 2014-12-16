package operator

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-operator-admin/models"
)

func RegisterHypervisorRoutes(prefix string, router *mux.Router) {
	router.HandleFunc(prefix, ListHypervisors).Methods("GET")
	router.HandleFunc(prefix, CreateHypervisor).Methods("POST")
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/{hypervisorID}", GetHypervisor).Methods("GET")
	sub.HandleFunc("/{hypervisorID}", UpdateHypervisor).Methods("PATCH")
	sub.HandleFunc("/{hypervisorID}", DeleteHypervisor).Methods("DELETE")
	sub.HandleFunc("/{hypervisorID}/ipranges", GetHypervisorIPRanges).Methods("GET")
	sub.HandleFunc("/{hypervisorID}/ipranges", SetHypervisorIPRanges).Methods("PUT")
	sub.HandleFunc("/{hypervisorID}/ipranges/{iprangeID}", AddHypervisorIPRange).Methods("PUT")
	sub.HandleFunc("/{hypervisorID}/ipranges/{iprangeID}", RemoveHypervisorIPRange).Methods("DELETE")
}

func ListHypervisors(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	hypervisors, err := models.ListHypervisors()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, hypervisors)
}

func GetHypervisor(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	hypervisor, ok := getHypervisorHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, hypervisor)
}

func CreateHypervisor(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}

	// Parse Request
	hypervisor := &models.Hypervisor{}
	if err := hypervisor.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	// Assign an ID
	if hypervisor.ID != "" {
		hr.JSONMsg(http.StatusBadRequest, "id must not be defined")
		return
	}
	hypervisor.NewID()

	if !saveHypervisorHelper(hr, hypervisor) {
		return
	}
	hr.JSON(http.StatusCreated, hypervisor)
}

func UpdateHypervisor(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	hypervisor, ok := getHypervisorHelper(hr, r)
	if !ok {
		return // Specific response handled by getHypervisorHelper
	}

	// Parse Request
	if err := hypervisor.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	if !saveHypervisorHelper(hr, hypervisor) {
		return
	}
	hr.JSON(http.StatusOK, hypervisor)
}

func DeleteHypervisor(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	hypervisor, ok := getHypervisorHelper(hr, r)
	if !ok {
		return
	}

	if err := hypervisor.Delete(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, hypervisor)
}

func GetHypervisorIPRanges(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	hypervisor, ok := getHypervisorHelper(hr, r)
	if !ok {
		return
	}
	if err := hypervisor.LoadIPRanges(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, hypervisor.IPRanges)
}

func SetHypervisorIPRanges(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	hypervisor, ok := getHypervisorHelper(hr, r)
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

	if err := hypervisor.SetIPRanges(ipranges); err != nil {
		hr.JSONMsg(http.StatusInternalServerError, err.Error())
		return
	}

	hr.JSON(http.StatusOK, hypervisor.IPRanges)
}

func AddHypervisorIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	hypervisor, ok := getHypervisorHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	iprangeID, ok := vars["iprangeID"]

	if err := hypervisor.AddIPRange(&models.IPRange{ID: iprangeID}); err != nil {
		hr.JSONMsg(http.StatusInternalServerError, err.Error())
		return
	}
	hr.JSON(http.StatusCreated, &struct{}{})
}

func RemoveHypervisorIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	hypervisor, ok := getHypervisorHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	iprangeID, ok := vars["iprangeID"]

	if err := hypervisor.RemoveIPRange(&models.IPRange{ID: iprangeID}); err != nil {
		hr.JSONMsg(http.StatusInternalServerError, err.Error())
		return
	}
	hr.JSON(http.StatusOK, &struct{}{})
}

func getHypervisorHelper(hr HTTPResponse, r *http.Request) (*models.Hypervisor, bool) {
	vars := mux.Vars(r)
	hypervisorID, ok := vars["hypervisorID"]
	if !ok {
		hr.JSONMsg(http.StatusBadRequest, "missing hypervisor id")
		return nil, false
	}
	if uuid.Parse(hypervisorID) == nil {
		hr.JSONMsg(http.StatusBadRequest, "invalid hypervisor id")
		return nil, false
	}
	hypervisor, err := models.FetchHypervisor(hypervisorID)
	if err != nil {
		if err == sql.ErrNoRows {
			hr.JSONMsg(http.StatusNotFound, "not found")
			return nil, false
		}
		hr.JSONError(http.StatusInternalServerError, err)
		return nil, false
	}
	return hypervisor, true
}

func saveHypervisorHelper(hr HTTPResponse, hypervisor *models.Hypervisor) bool {
	if err := hypervisor.Validate(); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	if err := hypervisor.Save(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
