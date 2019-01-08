// +build data

package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/shared/settings"
)

const (
	namespace1 = "test-namespace-1"
	namespace2 = "test-namespace-2"
	key1       = "key-1"
	key2       = "key-2"
	value1     = "value-1"
	value2     = "value-1"
)

var (
	testServer *mux.Router
	testData   *TestData
	dataClient data.Client
)

/* Helper methods that call the API endpoints and return the results */

func listNamespaces() string {
	req, _ := http.NewRequest("GET", "/api/metadata", nil)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	res := httptest.NewRecorder()
	testServer.ServeHTTP(res, req)
	return res.Body.String()
}

func listKeys(namespace string) string {
	path := fmt.Sprintf("/api/metadata/%s", namespace)
	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	res := httptest.NewRecorder()
	testServer.ServeHTTP(res, req)
	return res.Body.String()
}

func getKey(namespace, key string) string {
	path := fmt.Sprintf("/api/metadata/%s/%s", namespace, key)
	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	res := httptest.NewRecorder()
	testServer.ServeHTTP(res, req)
	return res.Body.String()
}

func deleteNamespace(namespace string) string {
	path := fmt.Sprintf("/api/metadata/%s", namespace)
	req, _ := http.NewRequest("DELETE", path, nil)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	res := httptest.NewRecorder()
	testServer.ServeHTTP(res, req)
	return res.Body.String()
}
func deleteKey(namespace, key string) string {
	path := fmt.Sprintf("/api/metadata/%s/%s", namespace, key)
	req, _ := http.NewRequest("DELETE", path, nil)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	res := httptest.NewRecorder()
	testServer.ServeHTTP(res, req)
	return res.Body.String()
}

func set(namespace string, newData map[string]string) string {
	path := fmt.Sprintf("/api/metadata/%s", namespace)

	formMap := make(map[string][]string)
	for key, value := range newData {
		formMap[key] = []string{value}
	}
	form := url.Values(formMap)

	req, _ := http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	res := httptest.NewRecorder()
	testServer.ServeHTTP(res, req)
	return res.Body.String()
}

func cleanupNamespaces(t *testing.T) {
	dataClient := data.NewClient()
	for _, namespace := range []string{namespace1, namespace2} {
		err := dataClient.MetadataDeleteNamespace(namespace)
		assert.NoError(t, err)
	}
}

/* Tests start here. Their order matters. */

func TestSetup(t *testing.T) {
	testServer, testData = setup(t) // Initialize all test data for this suite.
	dataClient = data.NewClient()
}

func TestMetadataListNoNamespaces(t *testing.T) {
	// No keys or namespaces.
	assert.JSONEq(t, `{"result":[]}`, listNamespaces())
}

func TestMetadataListOneNamespace(t *testing.T) {
	// One key in a namespace.
	err := dataClient.MetadataSet(namespace1,
		map[string]string{key1: value1})
	assert.NoError(t, err)

	assert.JSONEq(t,
		fmt.Sprintf(`{"result":["%s"]}`, namespace1),
		listNamespaces())
}

func TestMetadataListTwoNamespaces(t *testing.T) {
	// One key in a different namespace.
	err := dataClient.MetadataSet(namespace2,
		map[string]string{key1: value1})
	assert.NoError(t, err)

	response := listNamespaces()
	// We don't care about namespace order, clients deal with any desired sorting.
	assert.Contains(t, response, namespace1)
	assert.Contains(t, response, namespace2)
}

func TestMetadataListNoKeys(t *testing.T) {
	// No keys or namespaces.
	cleanupNamespaces(t)
	assert.JSONEq(t, `{"result":[]}`, listKeys(namespace1))
}

func TestMetadataListOneKey(t *testing.T) {
	// One key in a namespace.
	err := dataClient.MetadataSet(namespace1,
		map[string]string{key1: value1})
	assert.NoError(t, err)

	assert.JSONEq(t,
		fmt.Sprintf(`{"result":["%s"]}`, key1),
		listKeys(namespace1))
}

func TestMetadataListDifferentKey(t *testing.T) {
	// Different key in a different namespace.
	err := dataClient.MetadataSet(namespace2,
		map[string]string{key2: value2})
	assert.NoError(t, err)

	assert.JSONEq(t,
		fmt.Sprintf(`{"result":["%s"]}`, key2),
		listKeys(namespace2))
}

func TestMetadataListTwoKeys(t *testing.T) {
	// Try with two keys in a namespace.
	err := dataClient.MetadataSet(namespace1,
		map[string]string{key2: value2})
	assert.NoError(t, err)

	response := listKeys(namespace1)
	// We don't care about key order, clients deal with any desired sorting.
	assert.Contains(t, response, key1)
	assert.Contains(t, response, key2)
}

func TestMetadataGetKeyNoNamespace(t *testing.T) {
	cleanupNamespaces(t)
	// Try to get a key from a non-existant namespace.
	assert.JSONEq(t, `{"error":"No such namespace or key"}`, getKey(namespace1, key1))
}

func TestMetadataGetKeySingle(t *testing.T) {
	// Try with one key in a namespace.
	err := dataClient.MetadataSet(namespace1,
		map[string]string{key1: value1})
	assert.NoError(t, err)

	assert.JSONEq(t,
		fmt.Sprintf(`{"result":"%s"}`, value1),
		getKey(namespace1, key1))
}

func TestMetadataGetKeySingleDifferent(t *testing.T) {
	// Different key in a different namespace.
	err := dataClient.MetadataSet(namespace2,
		map[string]string{key2: value2})
	assert.NoError(t, err)

	assert.JSONEq(t,
		fmt.Sprintf(`{"result":"%s"}`, value2),
		getKey(namespace2, key2))
}

func TestMetadataGetKeyTwo(t *testing.T) {
	// Two keys in a namespace.
	err := dataClient.MetadataSet(namespace1,
		map[string]string{key2: value2})
	assert.NoError(t, err)

	// Get both keys from the namespace.
	assert.Contains(t, getKey(namespace1, key1), value1)
	assert.Contains(t, getKey(namespace1, key2), value2)

	// Try to get a non-existant key from an existing namespace.
	assert.JSONEq(t, `{"error":"No such namespace or key"}`, getKey(namespace2, key1))
}

func TestMetadataSetAdminRequired(t *testing.T) {
	// Admin permissions should be required.
	settings.CustomizeAdminEmails([]string{}) // Remove admin users.
	assert.Contains(t, set(namespace1, nil), AdminPermissionMessage)
	settings.CustomizeAdminEmails([]string{testData.User.Email}) // Add admin user.
}

func TestMetadataSetNewNamespace(t *testing.T) {
	cleanupNamespaces(t)
	// Setting in a non-existant namespace should create it.
	assert.JSONEq(t, `{}`, set(namespace1, map[string]string{key1: value1}))
	keys, err := dataClient.MetadataListKeys(namespace1)
	assert.NoError(t, err)
	assert.Equal(t, []string{key1}, keys)
}

func TestMetadataSetDifferentKey(t *testing.T) {
	// Set another key in the new namespace.
	assert.JSONEq(t, `{}`, set(namespace1, map[string]string{key2: value1}))
	keys, err := dataClient.MetadataListKeys(namespace1)
	assert.NoError(t, err)
	assert.Equal(t, []string{key1, key2}, keys)
}

func TestMetadataSetOverride(t *testing.T) {
	// Override an existing value.
	assert.JSONEq(t, `{}`, set(namespace1, map[string]string{key1: value2}))
	value, err := dataClient.MetadataGetKey(namespace1, key1)
	assert.NoError(t, err)
	assert.Equal(t, value2, value)
}

func TestMetadataSetTwoKeys(t *testing.T) {
	// Set two keys at once in a different namespace.
	assert.JSONEq(t, `{}`, set(namespace2, map[string]string{key1: value1, key2: value1}))
	keys, err := dataClient.MetadataListKeys(namespace2)
	assert.NoError(t, err)
	assert.Equal(t, []string{key1, key2}, keys)
}

func TestMetadataSetIsolation(t *testing.T) {
	// The other namespace shouldn't have been affected.
	value, err := dataClient.MetadataGetKey(namespace1, key1)
	assert.NoError(t, err)
	assert.Equal(t, value2, value)
}

func TestMetadataDeleteNamespaceAdminRequired(t *testing.T) {
	// Admin permissions should be required.
	settings.CustomizeAdminEmails([]string{}) // Remove admin users.
	assert.Contains(t, deleteNamespace(namespace1), AdminPermissionMessage)
	settings.CustomizeAdminEmails([]string{testData.User.Email}) // Add admin user.
}

func TestMetadataDeleteNoNamespace(t *testing.T) {
	cleanupNamespaces(t)
	// Try to delete a non-existant namespace.
	assert.JSONEq(t, `{}`, deleteNamespace(namespace1))
}

func TestMetadataDeleteNamespaceSingleKey(t *testing.T) {
	// Delete namespace with one key.
	err := dataClient.MetadataSet(namespace1,
		map[string]string{key1: value1})
	assert.NoError(t, err)
	keys, err := dataClient.MetadataListNamespaces()
	assert.NoError(t, err)
	assert.Equal(t, []string{namespace1}, keys)

	// Delete the namespace and check that it was deleted.
	assert.JSONEq(t, `{}`, deleteNamespace(namespace1))
	keys, err = dataClient.MetadataListNamespaces()
	assert.NoError(t, err)
	assert.Equal(t, []string{}, keys)
}

func TestMetadataDeleteNamespaceDifferent(t *testing.T) {
	// Different key in a different namespace.
	err := dataClient.MetadataSet(namespace2,
		map[string]string{key2: value2})
	assert.NoError(t, err)

	// Delete a non-existant namespace.
	assert.JSONEq(t, `{}`, deleteNamespace(namespace1))

	// Check that the namespace still exists.
	keys, err := dataClient.MetadataListNamespaces()
	assert.NoError(t, err)
	assert.Equal(t, []string{namespace2}, keys)

	// Delete the namespace and check that it was deleted.
	assert.JSONEq(t, `{}`, deleteNamespace(namespace2))
	keys, err = dataClient.MetadataListNamespaces()
	assert.NoError(t, err)
	assert.Equal(t, []string{}, keys)
}

func TestMetadataDeleteNamespacesWithTwoKeys(t *testing.T) {
	// Try with two namespaces with two keys each.
	err := dataClient.MetadataSet(namespace1,
		map[string]string{key1: value1, key2: value2})
	assert.NoError(t, err)
	err = dataClient.MetadataSet(namespace2,
		map[string]string{key1: value1, key2: value2})
	assert.NoError(t, err)
	keys, err := dataClient.MetadataListNamespaces()
	assert.NoError(t, err)
	assert.Equal(t, []string{namespace1, namespace2}, keys)

	// Delete one namespace and check.
	assert.JSONEq(t, `{}`, deleteNamespace(namespace1))
	keys, err = dataClient.MetadataListNamespaces()
	assert.NoError(t, err)
	assert.Equal(t, []string{namespace2}, keys)

	// Delete the other namespace and check.
	assert.JSONEq(t, `{}`, deleteNamespace(namespace2))
	keys, err = dataClient.MetadataListNamespaces()
	assert.NoError(t, err)
	assert.Equal(t, []string{}, keys)
}

func TestMetadataDeleteKeyAdminRequired(t *testing.T) {
	// Admin permissions should be required.
	settings.CustomizeAdminEmails([]string{}) // Remove admin users.
	assert.Contains(t, deleteKey(namespace1, key1), AdminPermissionMessage)
	settings.CustomizeAdminEmails([]string{testData.User.Email}) // Add admin user.
}

func TestMetadataDeleteKeyNoNamespace(t *testing.T) {
	cleanupNamespaces(t)
	// Try to delete a key from a non-existant namespace.
	assert.JSONEq(t, `{}`, deleteKey(namespace1, key1))
}

func TestMetadataDeleteKey(t *testing.T) {
	// One key in a namespace.
	err := dataClient.MetadataSet(namespace1,
		map[string]string{key1: value1})
	assert.NoError(t, err)
	keys, err := dataClient.MetadataListKeys(namespace1)
	assert.NoError(t, err)
	assert.Equal(t, []string{key1}, keys)

	// Delete the key and check that it was deleted.
	assert.JSONEq(t, `{}`, deleteKey(namespace1, key1))
	keys, err = dataClient.MetadataListKeys(namespace1)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, keys)
}

func TestMetadataDeleteKeyDifferent(t *testing.T) {
	// Different key in a different namespace.
	err := dataClient.MetadataSet(namespace2,
		map[string]string{key2: value2})
	assert.NoError(t, err)

	// Delete a non-existant key in an existing namespace.
	assert.JSONEq(t, `{}`, deleteKey(namespace2, key1))

	// Delete a the key from the wrong, non-existant namespace.
	assert.JSONEq(t, `{}`, deleteKey(namespace1, key2))

	// Check that the key still exists.
	keys, err := dataClient.MetadataListKeys(namespace2)
	assert.NoError(t, err)
	assert.Equal(t, []string{key2}, keys)

	// Delete the key and check that it was deleted.
	assert.JSONEq(t, `{}`, deleteKey(namespace2, key2))
	keys, err = dataClient.MetadataListKeys(namespace2)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, keys)
}

func TestMetadataDeleteKeyTwoInNamespace(t *testing.T) {
	// Two keys in a namespace.
	err := dataClient.MetadataSet(namespace1,
		map[string]string{key1: value1, key2: value2})
	assert.NoError(t, err)
	keys, err := dataClient.MetadataListKeys(namespace1)
	assert.NoError(t, err)
	assert.Equal(t, []string{key1, key2}, keys)

	// Delete one key and check.
	assert.JSONEq(t, `{}`, deleteKey(namespace1, key1))
	keys, err = dataClient.MetadataListKeys(namespace1)
	assert.NoError(t, err)
	assert.Equal(t, []string{key2}, keys)

	// Delete the other key and check.
	assert.JSONEq(t, `{}`, deleteKey(namespace1, key2))
	keys, err = dataClient.MetadataListKeys(namespace1)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, keys)

	// Put them back, and delete them in the other order.
	err = dataClient.MetadataSet(namespace1,
		map[string]string{key1: value1, key2: value2})
	assert.NoError(t, err)

	// Delete one key and check.
	assert.JSONEq(t, `{}`, deleteKey(namespace1, key2))
	keys, err = dataClient.MetadataListKeys(namespace1)
	assert.NoError(t, err)
	assert.Equal(t, []string{key1}, keys)

	// Delete the other key and check.
	assert.JSONEq(t, `{}`, deleteKey(namespace1, key1))
	keys, err = dataClient.MetadataListKeys(namespace1)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, keys)
}

func TestMetadataDeleteKeyIsolated(t *testing.T) {
	// The same key in two namespaces.
	err := dataClient.MetadataSet(namespace1,
		map[string]string{key1: value1})
	assert.NoError(t, err)
	err = dataClient.MetadataSet(namespace2,
		map[string]string{key1: value1})
	assert.NoError(t, err)

	// Delete from the first namespace and check that the key is still in the other.
	assert.JSONEq(t, `{}`, deleteKey(namespace1, key1))
	keys, err := dataClient.MetadataListKeys(namespace2)
	assert.NoError(t, err)
	assert.Equal(t, []string{key1}, keys)
}
