package core

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/Nextdoor/conductor/services/code"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/services/messaging"
	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/services/ticket"
	"github.com/Nextdoor/conductor/shared/datadog"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/types"
)

func jobEndpoints() []endpoint {
	return []endpoint{
		newEp("/api/train/{train_id:[0-9]+}/phase/{phase_id:[0-9]+}/job", get, fetchJobs),
		newEp("/api/train/{train_id:[0-9]+}/phase/{phase_id:[0-9]+}/job", post, startJob),
		newEp(`/api/train/{train_id:[0-9]+}/phase/{phase_id:[0-9]+}/job/{job_name:[a-zA-Z0-9_\-]+}`, post, completeJob),
	}
}

// Returns phase, or a response if there was an error.
func parseJobQueryVars(r *http.Request, dataClient data.Client) (*types.Phase, *response) {
	vars := mux.Vars(r)

	trainIDStr := vars["train_id"]
	phaseIDStr := vars["phase_id"]

	trainID, err := strconv.ParseUint(trainIDStr, 10, 64)
	if err != nil {
		resp := errorResponse(
			fmt.Sprintf("Bad train_id value: %s", trainIDStr),
			http.StatusBadRequest)
		return nil, &resp
	}

	phaseID, err := strconv.ParseUint(phaseIDStr, 10, 64)
	if err != nil {
		resp := errorResponse(
			fmt.Sprintf("Bad phase_id value: %s", phaseIDStr),
			http.StatusBadRequest)
		return nil, &resp
	}

	train, err := dataClient.Train(trainID)
	if err != nil {
		resp := errorResponse(
			err.Error(),
			http.StatusInternalServerError)
		return nil, &resp
	}

	phase, err := dataClient.Phase(phaseID, train)
	if err != nil {
		resp := errorResponse(
			err.Error(),
			http.StatusInternalServerError)
		return nil, &resp
	}
	if phase.Train.ID != trainID {
		resp := errorResponse(
			"No phase found with that ID for this train.",
			http.StatusInternalServerError)
		return nil, &resp
	}

	return phase, nil
}

func fetchJobs(r *http.Request) response {
	dataClient := data.NewClient()
	targetPhase, resp := parseJobQueryVars(r, dataClient)
	if resp != nil {
		return *resp
	}

	return dataResponse(targetPhase.Jobs)
}

func isValidJobName(jobName string, phaseType types.PhaseType) bool {
	possibleJobNames := types.JobsForPhase(phaseType)
	for _, possibleJobName := range possibleJobNames {
		if possibleJobName == jobName {
			return true
		}
	}
	return false
}

func jobByName(jobName string, jobs []*types.Job) *types.Job {
	for _, job := range jobs {
		if job.Name == jobName {
			return job
		}
	}
	return nil
}

func startJob(r *http.Request) response {
	dataClient := data.NewClient()
	targetPhase, resp := parseJobQueryVars(r, dataClient)
	if resp != nil {
		return *resp
	}

	err := r.ParseForm()
	if err != nil {
		return errorResponse("Error parsing POST form", http.StatusBadRequest)
	}

	jobName := r.PostFormValue("name")
	url := r.PostFormValue("url")

	formErrs := make([]string, 0)
	if jobName == "" {
		formErrs = append(formErrs, "`name` must be set in POST form")
	}
	if url == "" {
		formErrs = append(formErrs, "`url` must be set in POST form")
	}
	if len(formErrs) > 0 {
		return errorResponse(
			fmt.Sprintf("Errors parsing form: %s", strings.Join(formErrs, ", ")),
			http.StatusBadRequest)
	}

	if !isValidJobName(jobName, targetPhase.Type) {
		return errorResponse(
			fmt.Sprintf("Job with name %s not expected for %s phase",
				jobName, targetPhase.Type.String()),
			http.StatusBadRequest)
	}

	activePhaseType := targetPhase.Train.ActivePhase
	if targetPhase.Before(activePhaseType) {
		return errorResponse(
			fmt.Sprintf(
				"Cannot start a job on a previous phase. Active phase is %s, target phase is %s.",
				activePhaseType.String(), targetPhase.Type.String()),
			http.StatusBadRequest)
	}

	if targetPhase.Type == types.Deploy {
		if targetPhase.Train.Blocked {
			return errorResponse(
				"Cannot start a deploy job for a blocked train.",
				http.StatusBadRequest)
		}
		if targetPhase.Train.CancelledAt.HasValue() {
			return errorResponse(
				"Cannot start a deploy job for a cancelled train.",
				http.StatusBadRequest)
		}
	}

	job := jobByName(jobName, targetPhase.Jobs)
	if job == nil {
		// This code path is only invoked if the pipeline and expected jobs change after phase creation.
		job, err = dataClient.CreateJob(targetPhase, jobName)
		if err != nil {
			return errorResponse("Error creating job", http.StatusInternalServerError)
		}
	}
	if job.StartedAt.HasValue() {
		logger.Error("Warning: Job with name %s has already been started for Train %d, Phase %d",
			jobName, targetPhase.Train.ID, targetPhase.ID)

		datadog.Incr("job.start", job.DatadogTags())
		datadog.Incr("job.restart", job.DatadogTags())
		err = dataClient.RestartJob(job, url)
		if err != nil {
			return errorResponse("Error restarting job", http.StatusInternalServerError)
		}
	} else {
		datadog.Incr("job.start", job.DatadogTags())
		err = dataClient.StartJob(job, url)
		if err != nil {
			return errorResponse("Error starting job", http.StatusInternalServerError)
		}
	}

	// Measure how long after the phase started did the job start.
	// If the job completes before the phase starts, we give a duration of 0.
	if job.Phase.StartedAt.HasValue() {
		timeSincePhaseStart := job.StartedAt.Value.Sub(job.Phase.StartedAt.Value)
		datadog.Gauge("job.start.time_since_phase_start", timeSincePhaseStart.Seconds(), job.DatadogTags())
	} else {
		datadog.Gauge("job.start.time_since_phase_start", 0, job.DatadogTags())
	}

	return dataResponse(job)
}

func completeJob(r *http.Request) response {
	dataClient := data.NewClient()
	targetPhase, resp := parseJobQueryVars(r, dataClient)
	if resp != nil {
		return *resp
	}

	vars := mux.Vars(r)
	jobName := vars["job_name"]

	err := r.ParseForm()
	if err != nil {
		return errorResponse("Error parsing POST form", http.StatusBadRequest)
	}

	result := r.PostFormValue("result")

	if result == "" {
		return errorResponse(
			"Errors parsing form: `result` must be set in POST form",
			http.StatusBadRequest)
	}

	if !isValidJobName(jobName, targetPhase.Type) {
		return errorResponse(
			fmt.Sprintf("Job with name %s not expected for %s phase",
				jobName, targetPhase.Type.String()),
			http.StatusBadRequest)
	}

	job := jobByName(jobName, targetPhase.Jobs)
	if job == nil || !job.StartedAt.HasValue() {
		return errorResponse(
			fmt.Sprintf("Job with name %s has not been started for Train %d, Phase %d",
				jobName, targetPhase.Train.ID, targetPhase.ID),
			http.StatusBadRequest)
	}
	if job.CompletedAt.HasValue() {
		return errorResponse(
			fmt.Sprintf("Job with name %s has already been completed for Train %d, Phase %d",
				jobName, targetPhase.Train.ID, targetPhase.ID),
			http.StatusBadRequest)
	}

	resultInt, err := strconv.Atoi(result)
	if err != nil {
		return errorResponse("Invalid result value", http.StatusBadRequest)
	}

	jobResult := types.JobResult(resultInt)
	if !jobResult.IsValid() {
		return errorResponse("Invalid result value", http.StatusBadRequest)
	}

	err = dataClient.CompleteJob(job, jobResult, "")
	if err != nil {
		return errorResponse("Error completing job", http.StatusInternalServerError)
	}

	messagingService := messaging.GetService()

	datadog.Incr("job.complete", job.DatadogTags())
	if jobResult == types.Ok {
		datadog.Incr("job.success", job.DatadogTags())
	} else {
		datadog.Incr("job.failure", job.DatadogTags())
		messagingService.JobFailed(job)
	}

	duration := job.CompletedAt.Value.Sub(job.StartedAt.Value)
	datadog.Gauge("job.duration", duration.Seconds(), job.DatadogTags())

	// Measure how long after the phase started did the job complete.
	// If the job completes before the phase starts, we give a duration of 0.
	if job.Phase.StartedAt.HasValue() {
		timeSincePhaseStart := job.CompletedAt.Value.Sub(job.Phase.StartedAt.Value)
		datadog.Gauge("job.complete.time_since_phase_start", timeSincePhaseStart.Seconds(), job.DatadogTags())
	} else {
		datadog.Gauge("job.complete.time_since_phase_start", 0, job.DatadogTags())
	}

	codeService := code.GetService()
	phaseService := phase.GetService()
	ticketService := ticket.GetService()
	checkPhaseCompletion(dataClient, codeService, messagingService, phaseService, ticketService, targetPhase)

	return emptyResponse()
}

func checkJobs(dataClient data.Client) {
	// TODO
}
