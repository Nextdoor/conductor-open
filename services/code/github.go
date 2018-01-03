package code

import (
	"errors"
	"net/http"

	githubRaw "github.com/google/go-github/github"

	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/github"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	// Github OAuth2 token for the Conductor application, from a user who has admin rights to the target repo.
	githubAdminToken = flags.EnvString("GITHUB_ADMIN_TOKEN", "")

	githubRepo          = flags.EnvString("GITHUB_REPO", "")
	githubRepoOwner     = flags.EnvString("GITHUB_REPO_OWNER", "")
	githubWebhookURL    = flags.EnvString("GITHUB_WEBHOOK_URL", "")
	githubWebhookSecret = flags.EnvString("GITHUB_WEBHOOK_SECRET", "")
)

type githubCode struct {
	codeClient github.Code
}

func newGithub() *githubCode {
	if githubAdminToken == "" {
		panic(errors.New("github_admin_token flag must be set."))
	}
	if githubRepoOwner == "" {
		panic(errors.New("github_repo_owner flag must be set."))
	}
	if githubRepo == "" {
		panic(errors.New("github_repo flag must be set."))
	}
	if githubWebhookURL == "" {
		panic(errors.New("github_webhook_url flag must be set."))
	}
	if githubWebhookSecret == "" {
		panic(errors.New("github_webhook_secret flag must be set."))
	}

	return &githubCode{
		codeClient: github.NewCode(
			githubAdminToken,
			githubRepoOwner,
			githubRepo,
			githubWebhookURL,
			githubWebhookSecret,
		)}
}

func (c *githubCode) CommitsOnBranch(branch string, max int) ([]*types.Commit, error) {
	apiCommits, err := c.codeClient.CommitsOnBranch(branch, max)
	if err != nil {
		return nil, err
	}
	return c.convertCommits(apiCommits, branch), nil
}

func (c *githubCode) CommitsOnBranchAfter(branch string, sha string) ([]*types.Commit, error) {
	apiCommits, err := c.codeClient.CommitsOnBranchAfter(branch, sha)
	if err != nil {
		return nil, err
	}
	return c.convertCommits(apiCommits, branch), nil
}

func (c *githubCode) CompareRefs(oldRef, newRef string) ([]*types.Commit, error) {
	apiCommits, err := c.codeClient.CompareRefs(oldRef, newRef)
	if err != nil {
		return nil, err
	}
	return c.convertCommits(apiCommits, newRef), nil
}

func (c *githubCode) Revert(sha1, branch string) error {
	return c.codeClient.Revert(sha1, branch)
}

func (c *githubCode) ParseWebhookForBranch(r *http.Request) (string, error) {
	return c.codeClient.ParseWebhookForBranch(r, branchRegex)
}

// Convert slice of github.RepositoryCommit into internal commit slice.
func (c *githubCode) convertCommits(apiCommits []*githubRaw.RepositoryCommit, branch string) []*types.Commit {
	commits := make([]*types.Commit, len(apiCommits))
	for i, apiCommit := range apiCommits {
		commits[i] = &types.Commit{
			SHA:         *apiCommit.SHA,
			Message:     *apiCommit.Commit.Message,
			Branch:      branch,
			AuthorName:  *apiCommit.Commit.Author.Name,
			AuthorEmail: *apiCommit.Commit.Author.Email,
			URL:         *apiCommit.HTMLURL,
		}
	}
	return commits
}
