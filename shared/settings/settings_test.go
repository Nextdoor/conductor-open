package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	adminEmailsFlag = "admin-1, admin-2,admin-3"
	noStagingVerificationUsersFlag = "no-staging-1,    no-staging-2"
	robotUserFlag = "robot-1,robot-2"
	deliveryJobsFlag = "delivery-1"
	verificationJobsFlag = "verification-1, verification-2"
	deployJobsFlag = "deploy-1"

	defer clearFlags()

	parseFlags()

	assert.Equal(t, "admin-1", AdminEmails[0])
	assert.Equal(t, "admin-2", AdminEmails[1])
	assert.Equal(t, "admin-3", AdminEmails[2])

	assert.Equal(t, "no-staging-1", NoStagingVerificationUsers[0])
	assert.Equal(t, "no-staging-2", NoStagingVerificationUsers[1])

	assert.Equal(t, "robot-1", RobotUsers[0])
	assert.Equal(t, "robot-2", RobotUsers[1])

	assert.Equal(t, "delivery-1", DeliveryJobs[0])

	assert.Equal(t, "verification-1", VerificationJobs[0])
	assert.Equal(t, "verification-2", VerificationJobs[1])

	assert.Equal(t, "deploy-1", DeployJobs[0])
}

func clearFlags() {
	adminEmailsFlag = ""
	noStagingVerificationUsersFlag = ""
	robotUserFlag = ""
	deliveryJobsFlag = ""
	verificationJobsFlag = ""
	deployJobsFlag = ""
}
