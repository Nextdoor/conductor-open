// +build data,phase

// To run Jenkins integration tests:
// Make sure you set the necessary environment variables, and then set the
// "jenkins" build flag.
//
// i.e.:
// PHASE_IMPL=jenkins \
// JENKINS_USERNAME=jenkins JENKINS_PASSWORD='supersecret' JENKINS_URL='https://jenkins-server' \
// JENKINS_BUILD_JOB='conductor-integration-build' JENKINS_DEPLOY_JOB='conductor-integration-deploy' \
// go test -tags=jenkins ./...
package core

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/shared/types"
)

func TestJenkinsPhase(t *testing.T) {
	server, testData := setup(t)
	go listen(t, server)

	phaseService := phase.GetService()

	err := phaseService.Start(types.Delivery,
		testData.Train.ID,
		testData.Train.ActivePhases.Delivery.ID,
		testData.Train.ActivePhases.Verification.ID,
		testData.Train.ActivePhases.Deploy.ID,
		"branch", "sha", nil)
	assert.NoError(t, err)

	// TODO: Test the job API calls.
}
