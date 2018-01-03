package core

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test that the panic is recovered.
func TestPanickingEndpoint(t *testing.T) {
	// Create a request to hit the handler.
	req, err := http.NewRequest("GET", "/test-panic", nil)
	assert.NoError(t, err)
	res := httptest.NewRecorder()

	handler := func(r *http.Request) response {
		panic("test-panic")
		return dataResponse("didn't panic")
	}

	// Create a server with test handler.
	endpoints := []endpoint{newOpenEp("/test-panic", get, handler)}
	server := NewServer(endpoints)
	assert.NoError(t, err)

	server.ServeHTTP(res, req)

	resp := res.Result()
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), `"error":"Panic: test-panic. Stack trace: `)
}
