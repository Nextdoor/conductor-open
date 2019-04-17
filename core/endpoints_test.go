// +build data

package core

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoEndpoints(t *testing.T) {
	NewServer([]endpoint{})
}

// Test that HTTP handler gets called.
func TestEndpoint(t *testing.T) {
	// Create a request to hit the handler.
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)
	res := httptest.NewRecorder()

	handler := func(r *http.Request) response {
		return dataResponse("test")
	}

	// Create a server with test handler.
	endpoints := []endpoint{newOpenEp("/test", get, handler)}
	server := NewServer(endpoints)

	server.ServeHTTP(res, req)

	resp := res.Result()
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), `{"result":"test"}`)
}

// Test that auth handler works when not authorized.
func TestAuthEndpointUnauthorized(t *testing.T) {
	// Create a request to hit the handler.
	req, err := http.NewRequest("GET", "/test-auth", nil)
	assert.NoError(t, err)
	res := httptest.NewRecorder()

	handler := func(r *http.Request) response {
		return dataResponse("test")
	}

	// Create a server with test handler.
	endpoints := []endpoint{newEp("/test-auth", get, handler)}
	server := NewServer(endpoints)

	server.ServeHTTP(res, req)

	resp := res.Result()
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), `{"error":"Unauthorized"}`)
}

// Test that auth handler works when authorized.
func TestAuthEndpointAuthorized(t *testing.T) {
	_, testData := setup(t)

	// Create a request to hit the handler.
	req, err := http.NewRequest("GET", "/test-auth", nil)
	req.AddCookie(testData.TokenCookie)
	assert.NoError(t, err)
	res := httptest.NewRecorder()

	handler := func(r *http.Request) response {
		return dataResponse("test")
	}

	// Create a server with test handler.
	endpoints := []endpoint{newEp("/test-auth", get, handler)}
	server := NewServer(endpoints)

	server.ServeHTTP(res, req)

	resp := res.Result()
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), `{"result":"test"}`)
}

func TestEndpointFormatting(t *testing.T) {
	// Create a handler that returns a struct.
	handler := func(r *http.Request) response {
		return dataResponse(struct {
			Key string `json:"key"`
		}{
			"value",
		})
	}

	// Create a server with the test handler.
	endpoints := []endpoint{newOpenEp("/test", get, handler)}
	server := NewServer(endpoints)

	// Send a request for that endpoint.
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	resp := res.Result()
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), `{"result":{"key":"value"}}`)

	// Send a request for that endpoint, specifying json format.
	req, err = http.NewRequest("GET", "/test.json", nil)
	assert.NoError(t, err)
	res = httptest.NewRecorder()

	server.ServeHTTP(res, req)

	resp = res.Result()
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Should be an identical result as above, since json is the default format.
	assert.Contains(t, string(body), `{"result":{"key":"value"}}`)

	// Send a request for that endpoint, specifying pretty format.
	req, err = http.NewRequest("GET", "/test.pretty", nil)
	assert.NoError(t, err)
	res = httptest.NewRecorder()

	server.ServeHTTP(res, req)

	resp = res.Result()
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Test that there are indents and spaces.
	prettyResult := `{
	"result": {
		"key": "value"
	}
}`
	assert.Contains(t, string(body), prettyResult)
}
