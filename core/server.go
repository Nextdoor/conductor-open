/* Conductor server definition.

 */
package core

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewServer(endpoints []endpoint) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	middlewares := []middleware{
		newPanicRecoveryMiddleware(),
		newAuthMiddleware(),
	}

	for _, ep := range endpoints {
		var handler http.Handler = ep
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i].Wrap(handler)
		}
		ep.Route(router, handler)
	}

	return router
}
