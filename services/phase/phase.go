/* Handles triggering phases and checking their status. */
package phase

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	implementationFlag = flags.EnvString("PHASE_IMPL", "fake")
)

type Service interface {
	Start(phaseType types.PhaseType, trainID,
		deliveryPhaseID, verificationPhaseID, deployPhaseID uint64, branch, sha string,
		buildUser *types.User) error
}

type Completeable interface {
	IsComplete() bool
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

func newService() Service {
	logger.Info("Using %s implementation for Phase service", implementationFlag)
	var service Service
	switch implementationFlag {
	case "fake":
		service = newFake()
	case "jenkins":
		service = newJenkins()
	default:
		panic(fmt.Errorf("Unknown Phase Implementation: %s", implementationFlag))
	}
	return service
}

func IsComplete(phaseType types.PhaseType, completedJobs []string, extraChecks ...Completeable) bool {
	if !AllJobsComplete(phaseType, completedJobs) {
		return false
	}

	for _, extraCheck := range extraChecks {
		if !extraCheck.IsComplete() {
			return false
		}
	}

	return true
}

type fake struct{}

func newFake() *fake {
	types.CustomizeJobs(types.Delivery, []string{"delivery-1", "delivery-2", "delivery-3"})
	types.CustomizeJobs(types.Verification, []string{"verification-1", "verification-2"})
	types.CustomizeJobs(types.Deploy, []string{"deploy-1", "deploy-2", "deploy-3"})
	return &fake{}
}

func (p *fake) Start(phaseType types.PhaseType, trainID,
	deliveryPhaseID, verificationPhaseID, deployPhaseID uint64, branch, sha string,
	buildUser *types.User) error {
	switch phaseType {
	case types.Delivery:
		return fakeDelivery(trainID, deliveryPhaseID, verificationPhaseID, deployPhaseID)
	case types.Verification:
		return fakeVerification(trainID, deliveryPhaseID, verificationPhaseID, deployPhaseID)
	case types.Deploy:
		return fakeDeploy(trainID, deliveryPhaseID, verificationPhaseID, deployPhaseID)
	}
	return nil
}

func fakeDelivery(trainID, deliveryPhaseID, verificationPhaseID, _ uint64) error {
	logger.Info("Starting fake delivery phase...")
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	defer waitGroup.Wait()
	go func() {
		defer waitGroup.Done()
		err := fakeJobRun("verification-1", trainID, verificationPhaseID, 0, time.Duration(0), time.Second*1)
		if err != nil {
			logger.Error("%v", err)
		}
	}()
	err := fakeJobRun("delivery-1", trainID, deliveryPhaseID, 0, time.Duration(0), time.Second*1)
	if err != nil {
		return err
	}
	err = fakeJobRun("delivery-2", trainID, deliveryPhaseID, 0, time.Duration(0), time.Second*1)
	if err != nil {
		return err
	}
	err = fakeJobRun("delivery-3", trainID, deliveryPhaseID, 0, time.Duration(0), time.Second*1)
	if err != nil {
		return err
	}
	return nil
}

func fakeVerification(trainID, _, verificationPhaseID, _ uint64) error {
	logger.Info("Starting fake verification phase...")
	err := fakeJobRun("verification-2", trainID, verificationPhaseID, 0, time.Duration(0), time.Second*5)
	if err != nil {
		return err
	}
	return nil
}

func fakeDeploy(trainID, _, _, deployPhaseID uint64) error {
	logger.Info("Starting fake deploy phase...")
	err := fakeJobRun("deploy-1", trainID, deployPhaseID, 0, time.Duration(0), time.Second*5)
	if err != nil {
		return err
	}

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	defer waitGroup.Wait()
	go func() {
		defer waitGroup.Done()
		fakeJobRun("deploy-2", trainID, deployPhaseID, 0, time.Duration(0), time.Second*5)
		if err != nil {
			logger.Error("%v", err)
		}
	}()

	err = fakeJobRun("deploy-3", trainID, deployPhaseID, 0, time.Duration(0), time.Second*8)
	if err != nil {
		return err
	}
	return nil
}

func fakeJobRun(jobName string, trainID, phaseID, result uint64,
	startDelay time.Duration, runtime time.Duration) error {
	time.Sleep(startDelay)
	err := fakeStartJob(jobName, trainID, phaseID)
	if err != nil {
		return err
	}
	time.Sleep(runtime)
	err = fakeCompleteJob(jobName, trainID, phaseID, result)
	if err != nil {
		return err
	}
	return nil
}

func fakeStartJob(jobName string, trainID, phaseID uint64) error {
	jobForm := url.Values{
		"name": []string{jobName},
		"url":  []string{fmt.Sprintf("http://job.com/%s", jobName)},
	}
	path := fmt.Sprintf("http://localhost/api/train/%d/phase/%d/job", trainID, phaseID)
	req, err := http.NewRequest("POST", path, strings.NewReader(jobForm.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("%v", err)
		if resp != nil {
			body, _ := ioutil.ReadAll(resp.Body)
			logger.Error("Body: %s", string(body))
		}
		return err
	}
	return nil
}

func fakeCompleteJob(jobName string, trainID, phaseID, result uint64) error {
	resultForm := url.Values{"result": []string{strconv.FormatUint(result, 10)}}
	path := fmt.Sprintf("http://localhost/api/train/%d/phase/%d/job/%s", trainID, phaseID, jobName)
	req, err := http.NewRequest("POST", path, strings.NewReader(resultForm.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("%v", err)
		if resp != nil {
			body, _ := ioutil.ReadAll(resp.Body)
			logger.Error("Body: %s", string(body))
		}
		return err
	}
	return nil
}
