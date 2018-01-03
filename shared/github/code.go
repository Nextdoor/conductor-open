package github

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"

	"github.com/google/go-github/github"

	"github.com/Nextdoor/conductor/shared/logger"
)

const paginationMax = 100

type Code interface {
	CommitsOnBranch(string, int) ([]*github.RepositoryCommit, error)
	CommitsOnBranchAfter(string, string) ([]*github.RepositoryCommit, error)
	CompareRefs(string, string) ([]*github.RepositoryCommit, error)
	Revert(sha1, branch string) error
	ParseWebhookForBranch(*http.Request, *regexp.Regexp) (string, error)
}

type code struct {
	client        *github.Client
	repoOwner     string
	repo          string
	webhookSecret string
}

func NewCode(codeToken, repoOwner, repo, webhookURL, webhookSecret string) Code {
	client, err := newClient(codeToken)
	if err != nil {
		panic(err)
	}

	name := "web"
	hook := &github.Hook{
		Name:   &name,
		Events: []string{"push"},
		Config: map[string]interface{}{
			"url":          webhookURL,
			"content_type": "json",
			"secret":       webhookSecret,
		},
	}

	_, _, err = client.Repositories.CreateHook(repoOwner, repo, hook)
	if err != nil {
		if githubErr, ok := err.(*github.ErrorResponse); ok {
			if githubErr.Errors[0].Message != "Hook already exists on this repository" {
				panic(githubErr)
			}
		} else {
			logger.Error("Could not set up webhook: %v", err)
		}
	}

	return &code{
		client:        client,
		repoOwner:     repoOwner,
		repo:          repo,
		webhookSecret: webhookSecret,
	}
}

func (g *code) CommitsOnBranch(branch string, max int) ([]*github.RepositoryCommit, error) {
	iterator := newCommitIterator(g, branch)
	commits := make([]*github.RepositoryCommit, 0)
	for {
		newCommits, next, err := iterator.next()
		if err != nil {
			return nil, err
		}
		commits = append(commits, newCommits...)
		if len(commits) > max {
			// Prune to count.
			commits = commits[:max]
			break
		}
		if !next {
			break
		}
	}
	// Returned in reverse order.
	return reverse(commits), nil
}

func (g *code) CommitsOnBranchAfter(branch, sha string) ([]*github.RepositoryCommit, error) {
	iterator := newCommitIterator(g, branch)
	commits := make([]*github.RepositoryCommit, 0)
	for {
		newCommits, next, err := iterator.next()
		if err != nil {
			return nil, err
		}
		stopAt := len(newCommits)
		found := false
		for i, newCommit := range newCommits {
			if *newCommit.SHA == sha {
				stopAt = i
				found = true
				break
			}
		}
		commits = append(commits, newCommits[:stopAt]...)
		if found {
			break
		}
		if !next {
			return nil, fmt.Errorf("Could not find sha %s on branch %s", sha, branch)
		}
	}
	// Returned in reverse order.
	return reverse(commits), nil
}

// Gets all commits between oldRef and newRef, in order from [newest, ..., oldest]
func (g *code) CompareRefs(oldRef, newRef string) ([]*github.RepositoryCommit, error) {
	commits := make([]*github.RepositoryCommit, 0)

	skipLast := false
	for {
		// Do a comparison between oldRef <-> newRef.
		comparison, _, err := g.client.Repositories.CompareCommits(g.repoOwner, g.repo, oldRef, newRef)
		if err != nil {
			return nil, err
		}

		newCommits := make([]*github.RepositoryCommit, 0)
		for i := range comparison.Commits {
			if skipLast && i == len(comparison.Commits)-1 {
				break
			}
			newCommits = append(newCommits, &comparison.Commits[i])
		}
		commits = append(newCommits, commits...)

		oldestRefFound := *comparison.Commits[0].SHA
		if len(comparison.Commits) == *comparison.TotalCommits {
			// We're done - No commits left behind.
			break
		}

		// Since we didn't reach the end yet, there are more commits, and the api limited us.
		// Do the next comparison between oldRef <-> oldestRefFound.
		// This will effectively skip the commits we just went through, and get the next "page".
		newRef = oldestRefFound
		// Need to skip one commit at the end, because the pagination is end-inclusive,
		// so oldestRefFound will be in the result again.
		skipLast = true
	}
	return commits, nil
}

func (g *code) Revert(sha1, branch string) error {
	return nil
}

func (g *code) ParseWebhookForBranch(r *http.Request, branchPattern *regexp.Regexp) (string, error) {
	payload, err := github.ValidatePayload(r, []byte(g.webhookSecret))
	if err != nil {
		return "", err
	}
	messageType := github.WebHookType(r)
	if messageType == "ping" {
		// Do nothing with ping. `go-github` doesn't parse these.
		return "", nil
	}

	event, err := github.ParseWebHook(messageType, payload)
	if err != nil {
		return "", err
	}
	switch event := event.(type) {
	case *github.PushEvent:
		repo := *event.Repo
		if *repo.Name != g.repo || *repo.Owner.Name != g.repoOwner {
			err := fmt.Errorf("Got a webhook for an unexpected repo: %+v", event)
			logger.Error("%v", err)
			return "", err
		}
		results := branchPattern.FindStringSubmatch(*event.Ref)
		if results == nil {
			logger.Debug("Push ref %s doesn't match branch pattern %s; skipping", *event.Ref, branchPattern.String())
			return "", nil
		}
		if len(results) != 2 {
			err := fmt.Errorf(
				"Branch pattern must have only one matching group (the branch); pattern is %s", branchPattern.String())
			logger.Error("%v", err)
			return "", err
		}
		branch := results[1]
		if len(event.Commits) == 0 {
			logger.Debug("No commits in push event; skipping")
			return "", nil
		}
		if *event.Deleted {
			logger.Debug("Push event for deleted event; skipping")
			return "", nil
		}
		return branch, nil
	default:
		err := fmt.Errorf("Unexpected event type: %s", reflect.TypeOf(event))
		logger.Error("%v", err)
		return "", err
	}
	return "", nil
}

type commitIterator struct {
	sha string
	g   *code

	page int
}

func newCommitIterator(g *code, startRef string) *commitIterator {
	return &commitIterator{
		sha:  startRef,
		g:    g,
		page: 1,
	}
}

func (i *commitIterator) next() ([]*github.RepositoryCommit, bool, error) {
	options := github.CommitsListOptions{}
	options.SHA = i.sha
	if i.page == 0 {
		i.page = 1
	}
	options.Page = i.page
	options.PerPage = paginationMax
	commits, resp, err := i.g.client.Repositories.ListCommits(i.g.repoOwner, i.g.repo, &options)
	// Note: Commits returned from newest to oldest.
	if err != nil {
		return nil, false, err
	}
	i.page = resp.NextPage
	return commits, resp.NextPage > 0, nil
}

// Sometimes we need to reverse, because some api endpoints return newest -> oldest.
func reverse(commits []*github.RepositoryCommit) []*github.RepositoryCommit {
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}
	return commits
}
