package phase

import (
	"sort"
	"time"

	"github.com/Nextdoor/conductor/shared/types"
)

const (
	// Jobs are allowed 5 minutes grace time to report they have started
	MaxJobStartDelay = time.Minute * 5

	// Jobs must complete within 30 minutes of when they have started
	MaxJobRuntime = time.Minute * 30
)

func AllJobsComplete(phaseType types.PhaseType, completedJobs []string) bool {
	expectedJobs := types.JobsForPhase(phaseType)

	if completedJobs == nil && expectedJobs == nil {
		return true
	}

	if completedJobs == nil || expectedJobs == nil {
		return false
	}

	if len(completedJobs) != len(expectedJobs) {
		return false
	}

	sort.Strings(completedJobs)
	sort.Strings(expectedJobs)

	for i := range completedJobs {
		if completedJobs[i] != expectedJobs[i] {
			return false
		}
	}

	return true
}

/*
TODO Need a Job type in phase?
// Check that the given expected jobs have started within the grace period.
//
// Returns string array of any missing jobs.
func MissingJobs(timeStarted, currentTime time.Time, expectedJobs []string, startedJobs []*data.Job) []string {
	expectedMap := make(map[string]bool)
	var missing []string
	for _, expectedJob := range expectedJobs {
		expectedMap[expectedJob] = false
	}
	// Track the expected jobs which have been reported as started to us
	for _, startedJob := range startedJobs {
		expectedMap[startedJob.Name] = true
	}
	// Given the map, make sure any missing expected jobs which haven't
	// been reported are within their reporting deadlines.
	for _, expectedJob := range expectedJobs {
		// An expected job hasn't started yet.
		if !expectedMap[expectedJob] {
			// Are we past the MaxExpectedJobStartDelay deadline?
			if currentTime.Sub(timeStarted) > MaxExpectedJobStartDelay {
				missing = append(missing, expectedJob)
			}
		}
	}

	return missing
}

func CheckJobRuntime(phaseStart time.Time, runningJobs []*data.Job) []*data.Job {
	var exceeded []*data.Job
	for _, job := range runningJobs {
		if job.StartedAt != nil && job.StartedAt.Sub(phaseStart) > MaxExpectedJobRuntime {
			exceeded = append(exceeded, job)
		}
	}
	return exceeded
}
*/
