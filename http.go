// Package operator is the primary package of the Operator Admin API
package operator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/bakins/go-metrics-middleware"
	"github.com/bakins/net-http-recover"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/mistifyio/mistify-operator-admin/metrics"
)

type (
	// HTTPResponse is a wrapper for http.ResponseWriter which provides access
	// to several convenience methods
	HTTPResponse struct {
		http.ResponseWriter
	}

	// HTTPError contains information for http error responses
	HTTPError struct {
		Message string   `json:"message"`
		Code    int      `json:"code"`
		Stack   []string `json:"stack"`
	}
)

// Run starts the server
func Run(port uint) error {
	router := mux.NewRouter()
	router.StrictSlash(true)

	m, err := metrics.GetObject(nil)
	if err != nil {
		log.Fatal(err)
	}
	mw := mmw.New(m)

	// Common middleware applied to every request
	commonMiddleware := alice.New(
		func(h http.Handler) http.Handler {
			return handlers.CombinedLoggingHandler(os.Stdout, h)
		},
		handlers.CompressHandler,
		func(h http.Handler) http.Handler {
			return recovery.Handler(os.Stderr, h, true)
		},
		mw.HandlerWrapper("OPERATOR-API"),
	)

	// NOTE: Due to weirdness with PrefixPath and StrictSlash, can't just pass
	// a prefixed subrouter to the register functions and have the base path
	// work cleanly. The register functions need to add a base path handler to
	// the main router before setting subhandlers on either main or subrouter

	// Register the various routes
	RegisterPermissionRoutes("/permissions", router)
	RegisterNetworkRoutes("/networks", router)
	RegisterIPRangeRoutes("/ipranges", router)
	RegisterHypervisorRoutes("/hypervisors", router)
	RegisterProjectRoutes("/projects", router)
	RegisterUserRoutes("/users", router)
	RegisterFlavorRoutes("/flavors", router)
	RegisterConfigRoutes("/config", router)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        commonMiddleware.Then(router),
		MaxHeaderBytes: 1 << 20,
	}
	return server.ListenAndServe()
}

// JSON writes appropriate headers and JSON body to the http response
func (hr *HTTPResponse) JSON(code int, obj interface{}) {
	hr.Header().Set("Content-Type", "application/json")
	hr.WriteHeader(code)
	encoder := json.NewEncoder(hr)
	if err := encoder.Encode(obj); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
	}
}

// JSONError prepares an HTTPError with a stack trace and writes it with
// HTTPResponse.JSON
func (hr *HTTPResponse) JSONError(code int, err error) {
	httpError := &HTTPError{
		Message: err.Error(),
		Code:    code,
		Stack:   make([]string, 0, 4),
	}
	for i := 1; ; i++ { //
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		httpError.Stack = append(httpError.Stack, fmt.Sprintf("%s:%d (0x%x)", file, line, pc))
	}
	hr.JSON(code, httpError)
}

// JSONMsg is a convenience method to write a JSON response with just a message
// string
func (hr *HTTPResponse) JSONMsg(code int, msg string) {
	msgObj := map[string]string{
		"message": msg,
	}
	hr.JSON(code, msgObj)
}
