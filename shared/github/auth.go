package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Auth interface {
	AuthorizeURL() string
	AccessTokenURL() string
	AccessToken(code string) (string, error)
	UserInfo(string) (string, string, string, error)
}

type auth struct {
	clientID     string
	clientSecret string
}

func NewAuth(clientID, clientSecret string) Auth {
	return &auth{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (g *auth) AuthorizeURL() string {
	return fmt.Sprintf("%s/login/oauth/authorize", githubHost)
}

func (g *auth) AccessTokenURL() string {
	return fmt.Sprintf("%s/login/oauth/access_token", githubHost)
}

func (g *auth) AccessToken(code string) (string, error) {
	data := url.Values{
		"client_id":     []string{g.clientID},
		"client_secret": []string{g.clientSecret},
		"code":          []string{code},
	}
	req, err := http.NewRequest("POST", g.AccessTokenURL(), strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	loginResponse := githubLoginResponse{}
	err = json.Unmarshal(b, &loginResponse)
	if err != nil {
		return "", err
	}
	if loginResponse.Error != "" {
		err = errors.New(loginResponse.Error)
		return "", err
	}

	return loginResponse.AccessToken, nil
}

type githubLoginResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error_description"`
}

func (g *auth) UserInfo(accessToken string) (string, string, string, error) {
	client, err := newClient(accessToken)
	if err != nil {
		return "", "", "", err
	}

	user, _, err := client.Users.Get("")
	if err != nil {
		return "", "", "", err
	}

	userEmails, _, err := client.Users.ListEmails(nil)
	if err != nil {
		return "", "", "", err
	}

	var email string
	for _, userEmail := range userEmails {
		if *userEmail.Primary {
			email = *userEmail.Email
			break
		}
	}
	if email == "" {
		return "", "", "", errors.New("No primary email.")
	}
	return *user.Name, email, *user.AvatarURL, nil
}
