package operator

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-operator-admin/models"
)

// RegisterConfigRoutes registers the config routes and handlers
func RegisterConfigRoutes(prefix string, router *mux.Router) {
	RegisterOneRoute(router, RouteInfo{prefix, GetConfig, []string{"GET"}, "config.get"})
	RegisterOneRoute(router, RouteInfo{prefix, SetConfig, []string{"PUT"}, "config.set"})
	RegisterOneRoute(router, RouteInfo{prefix, UpdateConfig, []string{"PATCH"}, "config.update"})
	sub := router.PathPrefix(prefix).Subrouter()
	RegisterOneRoute(sub, RouteInfo{"/{namespace}", GetConfigNamespace, []string{"GET"}, "config.namespace.get"})
	RegisterOneRoute(sub, RouteInfo{"/{namespace}", SetConfigNamespace, []string{"PUT"}, "config.namespace.set"})
	RegisterOneRoute(sub, RouteInfo{"/{namespace}", DeleteConfigNamespace, []string{"DELETE"}, "config.namespace.delete"})
	RegisterOneRoute(sub, RouteInfo{"/{namespace}/{key}", DeleteConfigKey, []string{"DELETE"}, "config.namespace.deletekey"})
}

// GetConfig gets the config
func GetConfig(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	config, ok := getConfigHelper(hr, r)
	if !ok {
		return
	}
	hr.JSON(http.StatusOK, config.Get())
}

// SetConfig sets the config
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

// UpdateConfig updates a portion of the config
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

// GetConfigNamespace gets a particular namespace of the config
func GetConfigNamespace(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	config, ok := getConfigHelper(hr, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	hr.JSON(http.StatusOK, config.GetNamespace(vars["namespace"]))
}

// SetConfigNamespace sets the config for a particular namespace
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

// DeleteConfigNamespace removes a particular namespace from the config
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

// DeleteConfigKey deletes a particular key from the config
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

// getConfigHelper gets the config object and handles sending a response in case
// of error
func getConfigHelper(hr HTTPResponse, r *http.Request) (*models.Config, bool) {
	config := models.NewConfig()
	err := config.Load()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return nil, false
	}
	return config, true
}

// saveConfigHelper handles saving the config object and sending a response in
// case of error
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
