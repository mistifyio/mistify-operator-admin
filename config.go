package operator

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-operator-admin/models"
)

func RegisterConfigRoutes(prefix string, router *mux.Router) {
	router.HandleFunc(prefix, GetConfig).Methods("GET")
	router.HandleFunc(prefix, SetConfig).Methods("PUT")
	router.HandleFunc(prefix, UpdateConfig).Methods("PATCH")
	router.HandleFunc(prefix, UpdateConfig).Methods("PATCH")
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/{namespace}", GetConfigNamespace).Methods("GET")
	sub.HandleFunc("/{namespace}", SetConfigNamespace).Methods("PUT")
	sub.HandleFunc("/{namespace}", DeleteConfigNamespace).Methods("DELETE")
	sub.HandleFunc("/{namespace}/{key}", DeleteConfigKey).Methods("DELETE")
}

func GetConfig(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	config, ok := getConfigHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, config.Get())
}

func SetConfig(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}

	config := models.NewConfig()
	if err := config.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	if !saveConfigHelper(hr, config) {
		return
	}
	hr.JSON(http.StatusOK, config.Get())
}

func UpdateConfig(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	config, ok := getConfigHelper(hr, r)
	if !ok {
		return
	}
	newConfig := models.NewConfig()
	if err := newConfig.Decode(r.Body); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}
	config.Merge(newConfig)
	if !saveConfigHelper(hr, config) {
		return
	}
	hr.JSON(http.StatusOK, config.Get())
}

func GetConfigNamespace(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	config, ok := getConfigHelper(hr, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	hr.JSON(http.StatusOK, config.GetNamespace(vars["namespace"]))
}

func SetConfigNamespace(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	config, ok := getConfigHelper(hr, r)
	if !ok {
		return
	}

	var ns map[string]string
	if err := json.NewDecoder(r.Body).Decode(&ns); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return
	}

	vars := mux.Vars(r)
	config.SetNamespace(vars["namespace"], ns)

	if !saveConfigHelper(hr, config) {
		return
	}
	hr.JSON(http.StatusOK, config.Get())
}

func DeleteConfigNamespace(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	config, ok := getConfigHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	config.DeleteNamespace(vars["namespace"])

	if !saveConfigHelper(hr, config) {
		return
	}
	hr.JSON(http.StatusOK, config.Get())

}

func DeleteConfigKey(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	config, ok := getConfigHelper(hr, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	config.DeleteValue(vars["namespace"], vars["key"])

	if !saveConfigHelper(hr, config) {
		return
	}
	hr.JSON(http.StatusOK, config.GetNamespace(vars["namespace"]))

}

func getConfigHelper(hr HTTPResponse, r *http.Request) (*models.Config, bool) {
	config := models.NewConfig()
	err := config.Load()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return nil, false
	}
	return config, true
}

func saveConfigHelper(hr HTTPResponse, config *models.Config) bool {
	if err := config.Validate(); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	if err := config.Save(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}