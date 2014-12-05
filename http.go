package operator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/bakins/net-http-recover"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

type (
	HttpResponse struct {
		http.ResponseWriter
	}

	HttpError struct {
		Message string   `json:"message"`
		Code    int      `json:"code"`
		Stack   []string `json:"stack"`
	}
)

func Run(port uint) error {
	router := mux.NewRouter()
	router.StrictSlash(true)

	commonMiddleware := alice.New(
		func(h http.Handler) http.Handler {
			return handlers.CombinedLoggingHandler(os.Stdout, h)
		},
		handlers.CompressHandler,
		func(h http.Handler) http.Handler {
			return recovery.Handler(os.Stderr, h, true)
		},
	)

	// NOTE: Due to weirdness with PrefixPath and StrictSlash, can't just pass
	// a prefixed subrouter to the register functions and have the base path
	// work cleanly. The register functions need to add a base path handler to
	// the main router before setting subhandlers on either main or subrouter

	// users.RegisterHandlers(r)
	// permissions.RegisterHandlers(r)
	// network.RegisterHandlers(r)
	// hypervisors.RegisterHandlers(r)
	RegisterFlavorRoutes("/flavors", router)
	// generalConfig.RegisterHandlers(r)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        commonMiddleware.Then(router),
		MaxHeaderBytes: 1 << 20,
	}
	return server.ListenAndServe()
}

func (hr *HttpResponse) JSON(code int, obj interface{}) {
	hr.Header().Set("Content-Type", "application/json")
	hr.WriteHeader(code)
	encoder := json.NewEncoder(hr)
	if err := encoder.Encode(obj); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
	}
}

func (hr *HttpResponse) JSONError(code int, err error) {
	httpError := &HttpError{
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

func (hr *HttpResponse) JSONMsg(code int, msg string) {
	msgObj := map[string]string{
		"message": msg,
	}
	hr.JSON(code, msgObj)
}
