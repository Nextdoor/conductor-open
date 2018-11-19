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
	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/shared/types"
)

func TestJobGet(t *testing.T) {
	server, testData := setup(t)

	dataClient := data.NewClient()

	targetPhase := testData.Train.ActivePhases.Delivery
	path := fmt.Sprintf("/api/train/%d/phase/%d/job", testData.Train.ID, targetPhase.ID)
	req, err := http.NewRequest("GET", path, nil)
	assert.NoError(t, err)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.JSONEq(t, `{"result":[]}`, res.Body.String())

	jobName := "test_job"
	job, err := dataClient.CreateJob(targetPhase, jobName)
	assert.NoError(t, err)
	err = dataClient.StartJob(job, "http://example.com/myjob")
	assert.NoError(t, err)

	path = fmt.Sprintf("/api/train/%d/phase/%d/job", testData.Train.ID, targetPhase.ID)
	req, err = http.NewRequest("GET", path, nil)
	assert.NoError(t, err)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)

	jsonResponse := response{}
	err = json.Unmarshal(res.Body.Bytes(), &jsonResponse)
	assert.NoError(t, err)
	assert.Nil(t, jsonResponse.Error)

	b, err := json.Marshal(jsonResponse.Result)
	assert.NoError(t, err)

	receivedJobs := []types.Job{}
	err = json.Unmarshal(b, &receivedJobs)
	assert.NoError(t, err)
	assert.Len(t, receivedJobs, 1, "The length of results array is not 1")

	assert.WithinDuration(t, *receivedJobs[0].StartedAt.Get(), *job.StartedAt.Get(), time.Second*2)
	assert.Equal(t, receivedJobs[0].URL, job.URL)
	assert.Equal(t, receivedJobs[0].Name, jobName)
	assert.Nil(t, receivedJobs[0].CompletedAt.Get())
}

func TestJobCreate(t *testing.T) {
	server, testData := setup(t)

	dataClient := data.NewClient()

	jobName := "test_job_1"
	types.CustomizeJobs(types.Delivery, []string{"test_job_1"})

	targetPhase := testData.Train.ActivePhases.Delivery
	path := fmt.Sprintf("/api/train/%d/phase/%d/job", testData.Train.ID, targetPhase.ID)
	form := url.Values{
		"name": []string{jobName},
		"url":  []string{"http://example.com/1"},
	}
	// Create new job via POST request.
	req, err := http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	jsonResponse := response{}
	err = json.Unmarshal(res.Body.Bytes(), &jsonResponse)
	assert.NoError(t, err)
	assert.Nil(t, jsonResponse.Error)
	targetPhase, err = dataClient.Phase(targetPhase.ID, testData.Train)
	assert.NoError(t, err)

	jobs := targetPhase.Jobs
	assert.Equal(t, jobs[0].Name, jobName)

	// Create same job via POST request - it should succeed, replacing the job.
	req, err = http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)

	jsonResponse = response{}
	err = json.Unmarshal(res.Body.Bytes(), &jsonResponse)
	assert.NoError(t, err)
	assert.Nil(t, jsonResponse.Error)
	targetPhase, err = dataClient.Phase(targetPhase.ID, testData.Train)
	assert.NoError(t, err)

	jobs = targetPhase.Jobs
	assert.Equal(t, jobs[0].Name, jobName)
}

func TestNoDeployWhenBlocked(t *testing.T) {
	server, testData := setup(t)

	dataClient := data.NewClient()

	jobName := "test_job_1"
	types.CustomizeJobs(types.Deploy, []string{"test_job_1"})

	targetPhase := testData.Train.ActivePhases.Deploy
	path := fmt.Sprintf("/api/train/%d/phase/%d/job", testData.Train.ID, targetPhase.ID)
	form := url.Values{
		"name": []string{jobName},
		"url":  []string{"http://example.com/1"},
	}

	// Block train.
	blockReason := "reason"
	err := dataClient.BlockTrain(testData.Train, &blockReason)
	assert.NoError(t, err)
	assert.Equal(t, &blockReason, testData.Train.BlockedReason)

	// Create new job via POST request.
	req, err := http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	jsonResponse := response{}
	err = json.Unmarshal(res.Body.Bytes(), &jsonResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Cannot start a deploy job for a blocked train.", jsonResponse.Error)
	targetPhase, err = dataClient.Phase(targetPhase.ID, testData.Train)
	assert.NoError(t, err)

	jobs := targetPhase.Jobs
	assert.Len(t, jobs, 0)
}

func TestNoDeployWhenCancelled(t *testing.T) {
	server, testData := setup(t)

	dataClient := data.NewClient()

	jobName := "test_job_1"
	types.CustomizeJobs(types.Deploy, []string{"test_job_1"})

	targetPhase := testData.Train.ActivePhases.Deploy
	path := fmt.Sprintf("/api/train/%d/phase/%d/job", testData.Train.ID, targetPhase.ID)
	form := url.Values{
		"name": []string{jobName},
		"url":  []string{"http://example.com/1"},
	}

	// Cancel train.
	err := dataClient.CancelTrain(testData.Train)
	assert.NoError(t, err)

	// Create new job via POST request.
	req, err := http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	jsonResponse := response{}
	err = json.Unmarshal(res.Body.Bytes(), &jsonResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Cannot start a deploy job for a cancelled train.", jsonResponse.Error)
	targetPhase, err = dataClient.Phase(targetPhase.ID, testData.Train)
	assert.NoError(t, err)

	jobs := targetPhase.Jobs
	assert.Len(t, jobs, 0)
}

func TestJobComplete(t *testing.T) {
	server, testData := setup(t)

	phase.GetService()
	dataClient := data.NewClient()

	jobName := "test_job_1"
	types.CustomizeJobs(types.Delivery, []string{"test_job_1"})

	targetPhase := testData.Train.ActivePhases.Delivery
	// Create job which we'll complete via POST request.

	job, err := dataClient.CreateJob(targetPhase, jobName)
	assert.NoError(t, err)
	err = dataClient.StartJob(job, "http://example.com/myjob")
	assert.NoError(t, err)

	path := fmt.Sprintf("/api/train/%d/phase/%d/job/%s", testData.Train.ID, targetPhase.ID, job.Name)

	form := url.Values{
		"result": []string{"1"},
	}
	// Complete the job
	req, err := http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	jsonResponse := response{}
	assert.Contains(t, res.Body.String(), `{}`)

	targetPhase, err = dataClient.Phase(targetPhase.ID, testData.Train)
	assert.NoError(t, err)

	jobs := targetPhase.Jobs
	assert.Equal(t, jobs[0].Result, types.JobResult(1))
	assert.NotNil(t, jobs[0].CompletedAt.Get())

	// Complete the job again - it should fail.
	req, err = http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Add("X-Conductor-User", testData.User.Name)
	req.Header.Add("X-Conductor-Email", testData.User.Email)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)

	jsonResponse = response{}
	err = json.Unmarshal(res.Body.Bytes(), &jsonResponse)
	assert.NoError(t, err)
	assert.Equal(t,
		fmt.Sprintf("Job with name %s has already been completed for Train %d, Phase %d",
			jobName, testData.Train.ID, targetPhase.ID),
		jsonResponse.Error)

	targetPhase, err = dataClient.Phase(targetPhase.ID, testData.Train)
	assert.NoError(t, err)

	jobs = targetPhase.Jobs
	assert.Equal(t, jobs[0].Result, types.JobResult(1))
	assert.NotNil(t, jobs[0].CompletedAt.Get())
}
