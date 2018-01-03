package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	adminUserFlag = "admin-1, admin-2,admin-3"
	noStagingUsersFlag = "no-staging-1,    no-staging-2"
	robotUserFlag = "robot-1,robot-2"
	deliveryJobsFlag = "delivery-1"
	verificationJobsFlag = "verification-1, verification-2"
	deployJobsFlag = "deploy-1"

	defer clearFlags()

	parseFlags()

	assert.Equal(t, "admin-1", AdminUsers[0])
	assert.Equal(t, "admin-2", AdminUsers[1])
	assert.Equal(t, "admin-3", AdminUsers[2])

	assert.Equal(t, "no-staging-1", NoStagingUsers[0])
	assert.Equal(t, "no-staging-2", NoStagingUsers[1])

	assert.Equal(t, "robot-1", RobotUsers[0])
	assert.Equal(t, "robot-2", RobotUsers[1])

	assert.Equal(t, "delivery-1", DeliveryJobs[0])

	assert.Equal(t, "verification-1", VerificationJobs[0])
	assert.Equal(t, "verification-2", VerificationJobs[1])

	assert.Equal(t, "deploy-1", DeployJobs[0])
}

func clearFlags() {
	adminUserFlag = ""
	noStagingUsersFlag = ""
	robotUserFlag = ""
	deliveryJobsFlag = ""
	verificationJobsFlag = ""
	deployJobsFlag = ""
}
