package core

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Nextdoor/conductor/services/data"
)

func metadataEndpoints() []endpoint {
	return []endpoint{
		newEp("/api/metadata", get, metadataListNamespaces),
		newEp("/api/metadata/{namespace:[^/]+?}", get, metadataListKeys),
		newEp("/api/metadata/{namespace:[^/]+?}/{key:[^/]+?}", get, metadataGetKey),
		newAdminEp("/api/metadata/{namespace:[^/]+?}", post, metadataSet),
		newAdminEp("/api/metadata/{namespace:[^/]+?}", del, metadataDeleteNamespace),
		newAdminEp("/api/metadata/{namespace:[^/]+?}/{key:[^/]+?}", del, metadataDeleteKey),
	}
}

func metadataListNamespaces(r *http.Request) response {
	dataClient := data.NewClient()
	namespaces, err := dataClient.MetadataListNamespaces()
	if err != nil {
		return errorResponse(
			err.Error(),
			http.StatusInternalServerError)
	}
	return dataResponse(namespaces)
}

func metadataListKeys(r *http.Request) response {
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	dataClient := data.NewClient()
	keys, err := dataClient.MetadataListKeys(namespace)
	if err != nil {
		return errorResponse(
			err.Error(),
			http.StatusInternalServerError)
	}
	return dataResponse(keys)
}

func metadataGetKey(r *http.Request) response {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	key := vars["key"]

	dataClient := data.NewClient()
	value, err := dataClient.MetadataGetKey(namespace, key)
	if err != nil {
		if err == data.ErrNoSuchNamespaceOrKey {
			return errorResponse(
				err.Error(),
				http.StatusNotFound)
		}
		return errorResponse(
			err.Error(),
			http.StatusInternalServerError)
	}
	return dataResponse(value)
}

func metadataSet(r *http.Request) response {
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	err := r.ParseForm()
	if err != nil {
		return errorResponse("Error parsing POST form", http.StatusBadRequest)
	}

	newData := make(map[string]string)
	for key, values := range r.PostForm {
		if len(values) != 1 {
			return errorResponse(
				fmt.Sprintf("Bad POST form"),
				http.StatusBadRequest)
		}
		newData[key] = values[0]
	}

	dataClient := data.NewClient()

	err = dataClient.MetadataSet(namespace, newData)
	if err != nil {
		return errorResponse(
			err.Error(),
			http.StatusInternalServerError)
	}
	return emptyResponse()
}

func metadataDeleteNamespace(r *http.Request) response {
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	dataClient := data.NewClient()
	err := dataClient.MetadataDeleteNamespace(namespace)
	if err != nil {
		return errorResponse(
			err.Error(),
			http.StatusInternalServerError)
	}
	return emptyResponse()
}

func metadataDeleteKey(r *http.Request) response {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	key := vars["key"]

	dataClient := data.NewClient()
	err := dataClient.MetadataDeleteKey(namespace, key)
	if err != nil {
		return errorResponse(
			err.Error(),
			http.StatusInternalServerError)
	}
	return emptyResponse()
}
