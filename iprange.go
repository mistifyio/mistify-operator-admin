package operator

import (
	"database/sql"
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
}

func ListIPRanges(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	ipranges, err := models.ListIPRanges()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, ipranges)
}

func GetIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, iprange)
}

func CreateIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}

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

	ok := saveIPRangeHelper(hr, iprange)
	if !ok {
		return
	}
	hr.JSON(http.StatusCreated, iprange)
}

func UpdateIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return // Specific response handled by getIPRangeHelper
	}

	// Parse Request
	if err := iprange.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	ok = saveIPRangeHelper(hr, iprange)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, iprange)
}

func DeleteIPRange(w http.ResponseWriter, r *http.Request) {
	hr := HttpResponse{w}
	iprange, ok := getIPRangeHelper(hr, r)
	if !ok {
		return
	}

	err := iprange.Delete()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, iprange)
}

func getIPRangeHelper(hr HttpResponse, r *http.Request) (*models.IPRange, bool) {
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

func saveIPRangeHelper(hr HttpResponse, iprange *models.IPRange) bool {
	err := iprange.Validate()
	if err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	err = iprange.Save()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}
