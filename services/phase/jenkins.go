package phase

import (
	"strconv"

	"github.com/Nextdoor/conductor/services/build"
	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/settings"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	// These are the names of the Jenkins jobs which will kick off each phase.
	jenkinsDeliveryJob     = flags.EnvString("JENKINS_DELIVERY_JOB", "")
	jenkinsVerificationJob = flags.EnvString("JENKINS_VERIFICATION_JOB", "")
	jenkinsDeployJob       = flags.EnvString("JENKINS_DEPLOY_JOB", "")
)

type jenkinsPhase struct{}

func newJenkins() *jenkinsPhase {
	return &jenkinsPhase{}
}

func (p *jenkinsPhase) Start(phaseType types.PhaseType, trainID,
	deliveryPhaseID, verificationPhaseID, deployPhaseID uint64, branch, sha string,
	buildUser *types.User) error {

	params := make(map[string]string)
	params["TRAIN_ID"] = strconv.FormatUint(trainID, 10)
	params["DELIVERY_PHASE_ID"] = strconv.FormatUint(deliveryPhaseID, 10)
	params["VERIFICATION_PHASE_ID"] = strconv.FormatUint(verificationPhaseID, 10)
	params["DEPLOY_PHASE_ID"] = strconv.FormatUint(deployPhaseID, 10)
	params["BRANCH"] = branch
	params["SHA"] = sha
	params["CONDUCTOR_HOSTNAME"] = settings.GetHostname()
	if buildUser != nil {
		params["BUILD_USER"] = buildUser.Name
	} else {
		params["BUILD_USER"] = "Conductor"
	}

	var job string
	switch phaseType {
	case types.Delivery:
		job = jenkinsDeliveryJob
	case types.Verification:
		job = jenkinsVerificationJob
	case types.Deploy:
		job = jenkinsDeployJob
	}

	if job == "" {
		return nil
	}

	return build.Jenkins().TriggerJob(job, params)
}
