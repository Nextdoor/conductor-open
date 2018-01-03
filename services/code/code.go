/* Handles retrieving descriptions of the code and reverting. */
package code

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	implementationFlag = flags.EnvString("CODE_IMPL", "fake")
	// Target branch pattern (regex).
	branchPattern = flags.EnvString("BRANCH_PATTERN", "refs/heads/(master)")
)

var branchRegex *regexp.Regexp

type Service interface {
	CommitsOnBranch(string, int) ([]*types.Commit, error)
	CommitsOnBranchAfter(string, string) ([]*types.Commit, error)
	CompareRefs(string, string) ([]*types.Commit, error)
	Revert(sha1, branch string) error
	ParseWebhookForBranch(r *http.Request) (string, error)
}

var (
	service Service
	getOnce sync.Once
)

func GetService() Service {
	getOnce.Do(func() {
		service = newService()
	})
	return service
}

type fake struct{}

func newService() Service {
	branchRegex = regexp.MustCompile(branchPattern)

	logger.Info("Using %s implementation for Code service", implementationFlag)
	var service Service
	switch implementationFlag {
	case "fake":
		service = newFake()
	case "github":
		service = newGithub()
	default:
		panic(fmt.Errorf("Unknown Code Implementation: %s", implementationFlag))
	}
	return service
}

func newFake() *fake {
	return &fake{}
}

func (c *fake) CommitsOnBranch(branch string, max int) ([]*types.Commit, error) {
	return nil, nil
}

func (c *fake) CommitsOnBranchAfter(branch string, sha string) ([]*types.Commit, error) {
	return nil, nil
}

func (c *fake) CompareRefs(oldRef, newRef string) ([]*types.Commit, error) {
	return nil, nil
}

func (c *fake) Revert(sha1, branch string) error {
	return nil
}

func (c *fake) ParseWebhookForBranch(r *http.Request) (string, error) {
	return "", nil
}
