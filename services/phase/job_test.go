package phase

/* TODO: Complete after implementing background worker loops.
import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/shared/types"
)

func TestParseJobsString(t *testing.T) {
	emptyString := ""
	assert.Len(t, parseJobsString(emptyString), 0)

	oneJob := "delivery"
	assert.Len(t, parseJobsString(oneJob), 1)
	assert.Equal(t, parseJobsString(oneJob)[0], oneJob)

	threeJobs := "delivery   ,  test,build "
	assert.Len(t, parseJobsString(threeJobs), 3)
	assert.Equal(t, parseJobsString(threeJobs)[0], "delivery")
	assert.Equal(t, parseJobsString(threeJobs)[1], "test")
	assert.Equal(t, parseJobsString(threeJobs)[2], "build")
}

func TestGetExpectedJobsForPhase(t *testing.T) {
	mockPhase := data.Phase{Type: types.Delivery}
	*deliveryJobs = "test,   build,  delivery, static"
	*deployJobs = "delivery"

	res, _ := expectedJobsForPhase(&mockPhase)
	assert.Len(t, res, 4)
	assert.Equal(t, res[0], "test")
	assert.Equal(t, res[1], "build")
	assert.Equal(t, res[2], "delivery")
	assert.Equal(t, res[3], "static")

	mockPhase = data.Phase{Type: data.Deploy}
	res, _ = expectedJobsForPhase(&mockPhase)
	assert.Len(t, res, 1)
	assert.Equal(t, res[0], "delivery")
}

func TestGetExpectedJobsNotStarted(t *testing.T) {
	phaseStart := time.Time{}
	// Test that is time.Now() against MaxExpectedJobStartDelay, so we test this.
	beforeTimeout := phaseStart.Add(MaxJobStartDelay - 1)
	afterTimeout := phaseStart.Add(MaxJobStartDelay + 1)

	startedJobs := []*data.Job{
		{Name: "expected", StartedAt: &phaseStart},
	}

	expectedJobs := []string{"expected"}
	missingJobs := expectedJobsNotStarted(phaseStart, beforeTimeout, expectedJobs, startedJobs)
	assert.Empty(t, missingJobs)

	expectedJobs = []string{"expected", "expected_but_not_started"}
	missingJobs = expectedJobsNotStarted(phaseStart, afterTimeout, expectedJobs, startedJobs)
	assert.Equal(t, missingJobs[0], "expected_but_not_started")
}

func TestGetExpectedJobsExceededRuntime(t *testing.T) {
	phaseStart := time.Time{}
	almostExceeded := phaseStart.Add(MaxJobRuntime)
	reallyExceeded := phaseStart.Add(MaxJobRuntime + 1)

	// Test that an job exceeding the runtime is caught.
	jobs := []*data.Job{
		{Name: "one", StartedAt: &almostExceeded},
		{Name: "two", StartedAt: &reallyExceeded},
	}
	exceededJobs := expectedJobsExceededRuntime(phaseStart, jobs)
	assert.Equal(t, len(exceededJobs), 1)
	assert.Equal(t, exceededJobs[0].Name, jobs[1].Name)
}
*/
