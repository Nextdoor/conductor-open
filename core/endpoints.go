package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Nextdoor/conductor/shared/datadog"
)

func Endpoints() []endpoint {
	var endpoints []endpoint
	endpoints = append(endpoints, codeEndpoints()...)
	endpoints = append(endpoints, searchEndpoints()...)
	endpoints = append(endpoints, coreEndpoints()...)
	endpoints = append(endpoints, jobEndpoints()...)
	endpoints = append(endpoints, metadataEndpoints()...)
	endpoints = append(endpoints, phaseEndpoints()...)
	endpoints = append(endpoints, ticketEndpoints()...)
	endpoints = append(endpoints, trainEndpoints()...)
	endpoints = append(endpoints, userEndpoints()...)
	return endpoints
}

type httpMethod int

const (
	get httpMethod = iota
	post
	del
)

func (method httpMethod) String() string {
	switch method {
	case get:
		return "GET"
	case post:
		return "POST"
	case del:
		return "DELETE"
	default:
		panic(fmt.Errorf("Unknown httpMethod %s", string(method)))
	}
}

type middleware interface {
	Wrap(next http.Handler) http.Handler
}

type handlerFunc func(*http.Request) response

// Creates an endpoint that requires authentication.
// If the method is "get", it will require at least "viewer" permissions.
// Otherwise, it will require at least "user" permissions.
func newEp(path string, method httpMethod,
	handler handlerFunc) endpoint {
	return endpoint{
		uri:         path,
		method:      method,
		needsViewer: true,
		needsUser:   method != get,
		needsAdmin:  false,
		handler:     handler,
	}
}

// Creates an endpoint that requires admin authentication.
func newAdminEp(path string, method httpMethod,
	handler handlerFunc) endpoint {
	return endpoint{
		uri:         path,
		method:      method,
		needsViewer: true,
		needsUser:   true,
		needsAdmin:  true,
		handler:     handler,
	}
}

// Creates an endpoint that doesn't require authentication.
func newOpenEp(path string, method httpMethod,
	handler handlerFunc) endpoint {
	return endpoint{
		uri:         path,
		method:      method,
		needsViewer: false,
		needsUser:   false,
		needsAdmin:  false,
		handler:     handler,
	}
}

type endpoint struct {
	http.Handler

	uri         string
	method      httpMethod
	needsViewer bool
	needsUser   bool
	needsAdmin  bool
	handler     handlerFunc
}

func (e endpoint) NeedsAuth() bool {
	return e.needsViewer || e.needsUser || e.needsAdmin
}

func (e endpoint) Route(r *mux.Router, handler http.Handler) {
	r.NewRoute().
		//  Support different response formats.
		Path(fmt.Sprintf(`%s{format:(\.(json|pretty))?}`, e.uri)).
		Methods(e.method.String()).
		Handler(handler)
}

func (e endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.handler(r).Write(w, r)
}

type response struct {
	Result       interface{} `json:"result,omitempty"`
	Error        interface{} `json:"error,omitempty"`
	Code         int         `json:"-"`
	RedirectPath string      `json:"-"`
}

func (resp response) Write(w http.ResponseWriter, r *http.Request) {
	if resp.RedirectPath != "" {
		http.Redirect(w, r, resp.RedirectPath, resp.Code)
		return
	}

	vars := mux.Vars(r)
	format, formatSpecified := vars["format"]
	if !formatSpecified || len(format) == 0 {
		format = "json"
	} else {
		// Remove the period.
		format = format[1:]
	}

	var indent string
	switch format {
	case "json":
		indent = ""
	case "pretty":
		indent = "\t"
	}

	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", indent)
	err := encoder.Encode(resp)
	if err != nil {
		logMsg := fmt.Sprintf("Could not marshal response (%+v): %v", r, err)
		datadog.Error("%s", logMsg)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, logMsg)
		return
	}

	logResponse(r, resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.Code)
	fmt.Fprintln(w, buffer.String())
}

func logResponse(r *http.Request, resp response) {
	if resp.Code != http.StatusOK && resp.Code >= http.StatusInternalServerError {
		// Log on any server error code.
		if resp.Error != nil {
			datadog.Error("URL: %s. Responding with http error, code %d. Error message: %s",
				r.RequestURI, resp.Code, resp.Error)
		} else {
			datadog.Error("URL: %s. Responding with empty http error, code %d.",
				r.RequestURI, resp.Code)
		}
	}
}

func resultResponse(result interface{}, code int) response {
	return response{
		Result: result,
		Code:   code,
	}
}

func errorResponse(error interface{}, code int) response {
	return response{
		Error: error,
		Code:  code,
	}
}

func dataResponse(result interface{}) response {
	return resultResponse(result, http.StatusOK)
}

func emptyResponse() response {
	return dataResponse(nil)
}
