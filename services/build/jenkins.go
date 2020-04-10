package build

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Nextdoor/conductor/shared/datadog"
	"github.com/Nextdoor/conductor/shared/flags"
)

var (
	jenkinsURL      = flags.EnvString("JENKINS_URL", "")
	jenkinsUsername = flags.EnvString("JENKINS_USERNAME", "")
	jenkinsPassword = flags.EnvString("JENKINS_PASSWORD", "")
	jenkinsService  *jenkins
)

func Jenkins() Service {
	if jenkinsService == nil {
		// Initialize Jenkins.
		if jenkinsURL == "" {
			panic(errors.New("jenkins_url flag must be set."))
		}
		if jenkinsUsername == "" {
			panic(errors.New("jenkins_username flag must be set."))
		}
		if jenkinsPassword == "" {
			panic(errors.New("jenkins_password flag must be set."))
		}

		jenkinsService = &jenkins{
			URL:      jenkinsURL,
			Username: jenkinsUsername,
			Password: jenkinsPassword}

		err := jenkinsService.TestAuth()
		if err != nil {
			panic(err)
		}
	}

	return jenkinsService
}

type jenkins struct {
	URL      string
	Username string
	Password string
}

func (j jenkins) TestAuth() error {
	baseUrl, err := url.Parse(j.URL)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", baseUrl.String(), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(j.Username, j.Password)

	resp, err := j.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error connecting to Jenkins: %s", resp.Status)
	}
	return nil
}

func (j jenkins) CancelJob(jobURL string) error {

	jobURL = strings.TrimSuffix(jobURL, "/console")
	jobURL = jobURL + "/stop"

	req, err := http.NewRequest("POST", jobURL, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(j.Username, j.Password)

	resp, err := j.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (j jenkins) TriggerJob(jobName string, params map[string]string) error {
	datadog.Info("Triggering Jenkins Job \"%s\", Params: %s", jobName, params)
	buildUrl, err := url.Parse(fmt.Sprintf("%s/job/%s/buildWithParameters", j.URL, jobName))
	if err != nil {
		return err
	}

	urlParams := url.Values{}
	for k, v := range params {
		urlParams.Add(k, v)
	}
	buildUrl.RawQuery = urlParams.Encode()

	req, err := http.NewRequest("POST", buildUrl.String(), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(j.Username, j.Password)

	resp, err := j.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("Error building Jenkins job: %s", resp.Status)
	}
	return nil
}

func (j jenkins) Do(req *http.Request) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 15,
	}
	return client.Do(req)
}
