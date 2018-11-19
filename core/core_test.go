// +build data

package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/shared/settings"
	"github.com/Nextdoor/conductor/shared/types"
)

func TestGetConfig(t *testing.T) {
	server, testData := setup(t)

	dataClient := data.NewClient()

	err := dataClient.SetMode(types.Schedule)
	assert.NoError(t, err)
	err = dataClient.SetOptions(&types.DefaultOptions)
	assert.NoError(t, err)

	config, err := dataClient.Config()
	assert.NoError(t, err)

	expectedBytes, err := json.Marshal(config)
	assert.NoError(t, err)
	expectedConfig := string(expectedBytes)

	// Get config from endpoint and compare against config from dataClient.
	path := "/api/config"
	req, err := http.NewRequest("GET", path, nil)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	assert.NoError(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Contains(t, res.Body.String(), expectedConfig)

	options := types.DefaultOptions
	options.CloseTime[0].StartTime.Hour = 0
	err = dataClient.SetOptions(&options)
	assert.NoError(t, err)

	// Change options and update config.
	config, err = dataClient.Config()
	assert.NoError(t, err)

	expectedBytes, err = json.Marshal(config)
	assert.NoError(t, err)
	expectedConfig = string(expectedBytes)

	// Get config from endpoint and compare against new config from dataClient.
	server.ServeHTTP(res, req)
	assert.Contains(t, res.Body.String(), expectedConfig)
}

func TestGetAndSetMode(t *testing.T) {
	server, testData := setup(t)

	dataClient := data.NewClient()

	// Set mode to schedule and get.
	dataClient.SetMode(types.Schedule)
	mode, err := dataClient.Mode()
	assert.NoError(t, err)
	assert.Equal(t, types.Schedule, mode)

	// Get mode from endpoint and ensure schedule.
	path := "/api/mode"
	req, err := http.NewRequest("GET", path, nil)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	assert.NoError(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Contains(t, res.Body.String(),
		fmt.Sprintf(`{"result":"%s"}`, types.Schedule.String()))

	// Set mode to manual and get.
	err = dataClient.SetMode(types.Manual)
	mode, err = dataClient.Mode()
	assert.NoError(t, err)
	assert.Equal(t, types.Manual, mode)

	// Get mode from endpoint and ensure manual.
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Contains(t, res.Body.String(),
		fmt.Sprintf(`{"result":"%s"}`, types.Manual.String()))

	// Set mode to manual by endpoint, ensure failure on non-admin.
	settings.CustomizeAdminEmails([]string{}) // Remove admin users.
	form := url.Values{"mode": []string{"manual"}}
	req, err = http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	assert.NoError(t, err)
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Contains(t, res.Body.String(), AdminPermissionMessage)

	// Set current user to an admin user.
	settings.CustomizeAdminEmails([]string{testData.User.Email})

	// Set mode to manual by endpoint.
	form = url.Values{"mode": []string{"manual"}}
	req, err = http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	assert.NoError(t, err)
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Contains(t, res.Body.String(), `{"result":"manual"}`)

	// Get mode and ensure manual.
	mode, err = dataClient.Mode()
	assert.NoError(t, err)
	assert.Equal(t, types.Manual, mode)

	// Set mode to schedule by endpoint.
	form = url.Values{"mode": []string{"schedule"}}
	req, err = http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	assert.NoError(t, err)
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Contains(t, res.Body.String(), `{"result":"schedule"}`)

	// Get mode and ensure schedule.
	mode, err = dataClient.Mode()
	assert.NoError(t, err)
	assert.Equal(t, types.Schedule, mode)
}

func TestGetAndSetOptions(t *testing.T) {
	server, testData := setup(t)

	dataClient := data.NewClient()

	// Set options to default and get.
	dataClient.SetOptions(&types.DefaultOptions)
	options, err := dataClient.Options()
	assert.NoError(t, err)
	assert.Equal(t, &types.DefaultOptions, options)

	// Get options from endpoint and ensure they are the default options set above.
	path := "/api/options"
	req, err := http.NewRequest("GET", path, nil)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	assert.NoError(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Contains(t, res.Body.String(), options.String())

	// Change options and get.
	newOptions := types.DefaultOptions
	newOptions.CloseTime = append(newOptions.CloseTime,
		types.RepeatingTimeInterval{
			Every:     []time.Weekday{time.Sunday},
			StartTime: types.Clock{Hour: 0, Minute: 1},
			EndTime:   types.Clock{Hour: 2, Minute: 3},
		},
	)

	err = dataClient.SetOptions(&newOptions)
	options, err = dataClient.Options()
	assert.NoError(t, err)
	assert.Equal(t, &newOptions, options)

	// Get options and ensure they're the same as the new options.
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Contains(t, res.Body.String(), newOptions.String())

	// Update newOptions.
	newOptions.CloseTime[0].StartTime.Hour = 20

	// Set options by endpoint, ensure failure on non-admin.
	settings.CustomizeAdminEmails([]string{}) // Remove admin users.
	form := url.Values{"options": []string{newOptions.String()}}
	req, err = http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	assert.NoError(t, err)
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Contains(t, res.Body.String(), AdminPermissionMessage)

	// Set current user to an admin user.
	settings.CustomizeAdminEmails([]string{testData.User.Email})

	// Change options and ensure it worked.
	req, err = http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	assert.NoError(t, err)
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Contains(t, res.Body.String(), newOptions.String())

	// Get mode and ensure new options.
	options, err = dataClient.Options()
	assert.NoError(t, err)
	assert.Equal(t, options, &newOptions)

	// Ensure validation errors if options are invalid.
	invalidOptionStrings := []string{
		`invalid`,
		`{"close_time":null}`,
		`{"close_time!":null}`,
		`{"close_time":[]}`,
		`{"close_time":[{"every":[0]}]}`,
		`{"close_time":[{"every":[0],"start_time":{"hour": 0,"minute":0}}]}`,
		`{"close_time":[{"every":[0],"start_time":{"hour": 0,"minute":0},"end_time":{"hour":0}}]}`,
	}
	for _, invalidOptionString := range invalidOptionStrings {
		form = url.Values{"options": []string{invalidOptionString}}
		req, err = http.NewRequest("POST", path, strings.NewReader(form.Encode()))
		req.Header.Add("X-Conductor-User", testData.User.Name)
		req.Header.Add("X-Conductor-Email", testData.User.Email)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		assert.NoError(t, err)
		res = httptest.NewRecorder()
		server.ServeHTTP(res, req)
		assert.Contains(t, res.Body.String(), "Options validation error")
	}
}
