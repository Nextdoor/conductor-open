package core

import "net/http"

func userEndpoints() []endpoint {
	return []endpoint{
		newEp("/api/user", get, currentUser),
	}
}

func currentUser(r *http.Request) response {
	return dataResponse(r.Context().Value("user"))
}
