package github

import (
	"errors"
	"fmt"
	"net/url"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"

	"github.com/Nextdoor/conductor/shared/flags"
)

var (
	// No trailing slash.
	githubHost = flags.EnvString("GITHUB_HOST", "")
)

func newClient(accessToken string) (*github.Client, error) {
	if githubHost == "" {
		return nil, errors.New("github_host flag must be set")
	}
	githubURL, err := url.Parse(fmt.Sprintf("%s/api/v3/", githubHost))
	if err != nil {
		return nil, err
	}

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tokenClient := oauth2.NewClient(oauth2.NoContext, tokenSource)

	client := github.NewClient(tokenClient)
	client.BaseURL = githubURL
	return client, nil
}
