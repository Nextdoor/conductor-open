package auth

import (
	"errors"
	"net/http"

	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/github"
)

var (
	githubAuthClientID     = flags.EnvString("GITHUB_AUTH_CLIENT_ID", "")
	githubAuthClientSecret = flags.EnvString("GITHUB_AUTH_CLIENT_SECRET", "")
)

type githubAuth struct {
	authClient github.Auth
}

func newGithubAuth() *githubAuth {
	if githubAuthClientID == "" {
		panic(errors.New("github_auth_client_id flag must be set."))
	}
	if githubAuthClientSecret == "" {
		panic(errors.New("github_auth_client_secret flag must be set."))
	}
	return &githubAuth{
		authClient: github.NewAuth(githubAuthClientID, githubAuthClientSecret),
	}
}

func (a *githubAuth) AuthProvider() string {
	return "Github"
}

func (a *githubAuth) AuthURL(hostname string) string {
	req, _ := http.NewRequest("GET", a.authClient.AuthorizeURL(), nil)

	q := req.URL.Query()
	q.Add("client_id", githubAuthClientID)
	q.Add("redirect_uri", redirectEndpoint(hostname))
	q.Add("scope", "user repo")
	req.URL.RawQuery = q.Encode()

	return req.URL.String()
}

func (a *githubAuth) Login(code string) (string, string, string, string, error) {
	accessToken, err := a.authClient.AccessToken(code)
	if err != nil {
		return "", "", "", "", err
	}

	name, email, avatar, err := a.authClient.UserInfo(accessToken)
	if err != nil {
		return "", "", "", "", err
	}
	return name, email, avatar, accessToken, err
}
