// Package for shared Conductor settings.
package settings

import (
	"strings"

	"github.com/Nextdoor/conductor/shared/flags"
)

var (
	Hostname = flags.EnvString("HOSTNAME", "localhost")

	JenkinsRollbackJob = flags.EnvString("JENKINS_ROLLBACK_JOB", "")

	// Whether to not require staging verification for a commit by default.
	// If set, staging verification will only be required if the commit message has [needs-staging].
	NoStagingVerification = flags.EnvBool("NO_STAGING_VERIFICATION", false)

	// Comma-separated list of admin user emails that can deploy and change mode.
	adminUserFlag = flags.EnvString("ADMIN_USERS", "")

	// Comma-separated list of user emails who don't use staging by default.
	// This list is ignored if noStagingVerification is set.
	noStagingVerificationUsersFlag = flags.EnvString("NO_STAGING_VERIFICATION_USERS", "")

	// Comma-separated list of robot user emails that push commits.
	// Tickets will be assigned to the default user, they won't get notifications,
	// and they won't get engineer status.
	robotUserFlag = flags.EnvString("ROBOT_USERS", "")

	AdminUsers                 []string
	RobotUsers                 []string
	NoStagingVerificationUsers []string

	CustomAdminUsers                 []string
	CustomRobotUsers                 []string
	CustomNoStagingVerificationUsers []string
)

// Settings for job names to accept for delivery, verification, and deploy phases.
// These job names are customizable for tests.
// The logic below ensures that tests can modify them unperturbed by calls to ParseFlags.
// Calls to CustomizeJobs should only occur in tests.
var (
	// Comma-separated list of expected jobs for the delivery phase.
	deliveryJobsFlag = flags.EnvString("DELIVERY_JOBS", "")

	// Comma-separated list of expected jobs for the verification phase.
	verificationJobsFlag = flags.EnvString("VERIFICATION_JOBS", "")

	// Comma-separated list of expected jobs for the deploy phase.
	deployJobsFlag = flags.EnvString("DEPLOY_JOBS", "")

	DeliveryJobs     []string
	VerificationJobs []string
	DeployJobs       []string

	CustomDeliveryJobs     []string
	CustomVerificationJobs []string
	CustomDeployJobs       []string
)

func init() {
	parseFlags()
}

func parseFlags() {
	AdminUsers = parseListString(adminUserFlag)
	RobotUsers = parseListString(robotUserFlag)
	NoStagingVerificationUsers = parseListString(noStagingVerificationUsersFlag)

	DeliveryJobs = parseListString(deliveryJobsFlag)
	VerificationJobs = parseListString(verificationJobsFlag)
	DeployJobs = parseListString(deployJobsFlag)
}

// Take a comma-separated string and split on commas, stripping any whitespace.
func parseListString(s string) []string {
	f := func(c rune) bool {
		return c == ','
	}
	// Split by comma
	split := strings.FieldsFunc(s, f)
	var result = make([]string, len(split))
	// Trim any whitespace
	for i, s := range split {
		result[i] = strings.TrimSpace(s)
	}
	return result
}

func StringInList(text string, list []string) bool {
	for _, line := range list {
		if line == text {
			return true
		}
	}
	return false
}

func IsNoStagingVerificationUser(email string) bool {
	if CustomNoStagingVerificationUsers != nil {
		return StringInList(email, CustomNoStagingVerificationUsers)
	}
	return StringInList(email, NoStagingVerificationUsers)
}

// Should only be used for tests.
func CustomizeNoStagingVerificationUsers(noStagingVerificationUsers []string) {
	CustomNoStagingVerificationUsers = noStagingVerificationUsers
}

// Should only be used for tests.
func CustomizeAdminUsers(adminUsers []string) {
	CustomAdminUsers = adminUsers
}

// Should only be used for tests.
func CustomizeRobotUsers(robotUsers []string) {
	CustomRobotUsers = robotUsers
}

func GetHostname() string {
	return Hostname
}

func GetJenkinsRollbackJob() string {
	return JenkinsRollbackJob
}

func IsAdminUser(email string) bool {
	if CustomAdminUsers != nil {
		return StringInList(email, CustomAdminUsers)
	}
	return StringInList(email, AdminUsers)
}

func IsRobotUser(email string) bool {
	if CustomRobotUsers != nil {
		return StringInList(email, CustomRobotUsers)
	}
	return StringInList(email, RobotUsers)
}
