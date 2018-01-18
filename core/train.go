package core

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/Nextdoor/conductor/services/build"
	"github.com/Nextdoor/conductor/services/code"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/services/messaging"
	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/services/ticket"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/settings"
	"github.com/Nextdoor/conductor/shared/types"
)

var checkBranchLock sync.Mutex

func checkBranch(
	dataClient data.Client,
	codeService code.Service,
	messagingService messaging.Service,
	phaseService phase.Service,
	ticketService ticket.Service,
	branch string,
	requester *types.User) {
	checkBranchLock.Lock()
	defer checkBranchLock.Unlock()

	latestTrain, err := dataClient.LatestTrain()
	if err != nil {
		logger.Error("Error getting latest train: %v", err)
		return
	}
	latestTrainForBranch, err := dataClient.LatestTrainForBranch(branch)
	if err != nil {
		logger.Error("Error getting latest train for branch: %v", err)
		return
	}

	commits, err := getNewCommitsForBranch(codeService, branch, latestTrain, latestTrainForBranch)
	if err != nil {
		return
	}
	handleNewCommitsForBranch(
		dataClient, codeService, messagingService, phaseService, ticketService,
		branch, latestTrain, latestTrainForBranch, commits, requester)
}

func getNewCommitsForBranch(
	codeService code.Service,
	branch string,
	latestTrain *types.Train,
	latestTrainForBranch *types.Train) ([]*types.Commit, error) {

	var commits []*types.Commit
	var err error
	if latestTrain == nil {
		// This is the first train. Get 20 commits on the branch.
		commits, err = codeService.CommitsOnBranch(branch, 20)
		if err != nil {
			logger.Error("Error getting commits on branch: %v", err)
			return commits, err
		}
	} else if latestTrainForBranch == nil {
		// Compare the latest train to the new train.
		commits, err = codeService.CompareRefs(latestTrain.HeadSHA, branch)
		if err != nil {
			logger.Error("Error comparing branches: %v", err)
			return commits, err
		}
	} else {
		commits, err = codeService.CommitsOnBranchAfter(branch, latestTrainForBranch.HeadSHA)
		if err != nil {
			logger.Error("Error getting new commits on branch: %v", err)
			return commits, err
		}
	}
	return commits, nil
}

func handleNewCommitsForBranch(
	dataClient data.Client,
	codeService code.Service,
	messagingService messaging.Service,
	phaseService phase.Service,
	ticketService ticket.Service,
	branch string,
	latestTrain *types.Train,
	latestTrainForBranch *types.Train,
	newCommits []*types.Commit,
	requester *types.User) {

	if len(newCommits) == 0 {
		return
	}
	train := latestTrainForBranch
	if latestTrain == nil || latestTrainForBranch == nil || latestTrain.IsDeploying() || latestTrain.Done {
		if latestTrain != nil {
			// Clean up old train.
			err := ticketService.CloseTrainTickets(latestTrain)
			if err != nil {
				logger.Error("Error closing old train tickets: %v", err)
			}
		}
		train = CreateTrain(dataClient, messagingService, branch, newCommits)
	} else if latestTrainForBranch.ID == latestTrain.ID {
		// The latest train is for this branch.
		if !latestTrain.Closed {
			ExtendTrain(dataClient, messagingService, train, newCommits, requester)
		} else {
			QueueCommits(dataClient, newCommits)
			return
		}
	} else {
		// Clean up old train.
		err := ticketService.CloseTrainTickets(train)
		if err != nil {
			logger.Error("Error closing old train tickets: %v", err)
		}
		// The latest train for the branch can be repurposed.
		train = DuplicateTrain(dataClient, messagingService, train, newCommits)
	}

	if train != nil {
		go StartTrain(data.NewClient(), codeService, messagingService, phaseService, ticketService, train)
	}
}

func CreateTrain(
	dataClient data.Client,
	messagingService messaging.Service,
	branch string,
	commits []*types.Commit) *types.Train {

	engineer, err := chooseEngineer(dataClient, commits)
	if err != nil {
		logger.Error("Error choosing engineer: %v", err)
	}

	train, err := dataClient.CreateTrain(branch, engineer, commits)
	if err != nil {
		logger.Error("Error creating train: %v", err)
		return nil
	}

	messagingService.TrainCreation(train, commits)

	clearLatestTrainCache()

	return train
}

func ExtendTrain(
	dataClient data.Client,
	messagingService messaging.Service,
	train *types.Train,
	commits []*types.Commit,
	requester *types.User) {

	var err error

	engineer := train.Engineer
	if engineer == nil {
		engineer, err = chooseEngineer(dataClient, commits)
		if err != nil {
			logger.Error("Error choosing engineer: %v", err)
		}
	}

	err = dataClient.ExtendTrain(train, engineer, commits)
	if err != nil {
		logger.Error("Error extending train: %v", err)
		return
	}

	messagingService.TrainExtension(train, commits, requester)

	clearLatestTrainCache()
}

func DuplicateTrain(
	dataClient data.Client,
	messagingService messaging.Service,
	oldTrain *types.Train,
	commits []*types.Commit) *types.Train {

	train, err := dataClient.DuplicateTrain(oldTrain, commits)
	if err != nil {
		logger.Error("Error duplicating train: %v", err)
		return nil
	}

	messagingService.TrainDuplication(train, oldTrain, commits)

	clearLatestTrainCache()

	return train
}

func QueueCommits(dataClient data.Client, commits []*types.Commit) {
	if len(commits) == 0 {
		return
	}

	_, err := dataClient.WriteCommits(commits)
	if err != nil {
		logger.Error("Error queueing commits: %v", err)
	}

}

func StartTrain(
	dataClient data.Client,
	codeService code.Service,
	messagingService messaging.Service,
	phaseService phase.Service,
	ticketService ticket.Service,
	train *types.Train) {

	startPhase(dataClient, codeService, messagingService, phaseService, ticketService, train.ActivePhases.Delivery, nil)
}

func chooseEngineer(dataClient data.Client, commits []*types.Commit) (*types.User, error) {
	// Get a random user from the set of commits.
	filteredCommits := make([]*types.Commit, 0)
	for _, commit := range commits {
		if settings.IsRobotUser(commit.AuthorEmail) {
			continue
		}
		filteredCommits = append(filteredCommits, commit)
	}

	if len(filteredCommits) == 0 {
		return nil, nil
	}

	commit := filteredCommits[rand.Intn(len(filteredCommits))]
	engineer, err := dataClient.ReadOrCreateUser(commit.AuthorName, commit.AuthorEmail)
	if err != nil {
		return nil, err
	}

	return engineer, nil
}

func deployIfReady(
	dataClient data.Client,
	messagingService messaging.Service,
	train *types.Train) {

	if train.IsDeployable() {
		deployTrain(dataClient, messagingService, train)
	}
}

var deployTrainLock sync.Mutex

func deployTrain(
	dataClient data.Client,
	messagingService messaging.Service,
	train *types.Train) {

	// Use a lock to handle multiple deploy race conditions.
	deployTrainLock.Lock()
	defer deployTrainLock.Unlock()

	// Reload train object from database, in case it's changed.
	train, err := dataClient.Train(train.ID)
	if err != nil {
		logger.Error("Error getting train: %v", err)
		return
	}

	if !train.IsDeployable() {
		// The train isn't deployable.
		return
	}

	messagingService.TrainDeploying()
	codeService := code.GetService()
	phaseService := phase.GetService()
	ticketService := ticket.GetService()
	startPhase(dataClient, codeService, messagingService, phaseService, ticketService,
		train.ActivePhases.Deploy, train.Engineer)
}

func trainEndpoints() []endpoint {
	return []endpoint{
		newEp("/api/train", get, fetchTrain),
		newEp("/api/train/{train_id:[0-9]+}", get, fetchTrain),
		newEp("/api/train/{train_id:[0-9]+}/close", post, closeTrain),
		newEp("/api/train/{train_id:[0-9]+}/open", post, openTrain),
		newEp("/api/train/{train_id:[0-9]+}/extend", post, extendTrain),
		newEp("/api/train/{train_id:[0-9]+}/block", post, blockTrain),
		newEp("/api/train/{train_id:[0-9]+}/unblock", post, unblockTrain),
		newEp("/api/train/{train_id:[0-9]+}/cancel", post, cancelTrain),
		newEp("/api/train/{train_id:[0-9]+}/rollback", post, rollbackTrain),
	}
}

// Returns train, or a response if there was an error.
func parseTrainVars(r *http.Request, dataClient data.Client, readFromCache bool) (*types.Train, *response) {
	vars := mux.Vars(r)

	trainIDStr, trainIDSpecified := vars["train_id"]
	if !trainIDSpecified {
		train, err := getCacheBackedLatestTrain(dataClient, readFromCache)
		if err != nil {
			resp := errorResponse(
				fmt.Sprintf("Error getting train: %v", err),
				http.StatusBadRequest)
			return nil, &resp
		}
		return train, nil
	}

	trainID, err := strconv.ParseUint(trainIDStr, 10, 64)
	if err != nil {
		resp := errorResponse(
			fmt.Sprintf("Bad train_id value: %s", trainIDStr),
			http.StatusBadRequest)
		return nil, &resp
	}

	train, err := dataClient.Train(trainID)
	if err != nil {
		resp := errorResponse(
			fmt.Sprintf("Error getting train: %v", err),
			http.StatusBadRequest)
		return nil, &resp
	}

	if train == nil {
		resp := errorResponse("Train not found.", http.StatusNotFound)
		return nil, &resp
	}

	return train, nil
}

var latestTrainCache *types.Train
var latestTrainCacheUnixTime int64

const TRAIN_CACHE_TTL = 5

func getCacheBackedLatestTrain(dataClient data.Client, readFromCache bool) (*types.Train, error) {
	now := time.Now()
	if readFromCache && latestTrainCacheUnixTime != 0 && now.Unix()-latestTrainCacheUnixTime <= TRAIN_CACHE_TTL {
		return latestTrainCache, nil
	}

	// Not read from cache, read from database and update cache.
	train, err := dataClient.LatestTrain()
	if err != nil {
		return nil, err
	}
	latestTrainCache = train
	latestTrainCacheUnixTime = now.Unix()
	return train, nil
}

func clearLatestTrainCache() {
	latestTrainCache = nil
	latestTrainCacheUnixTime = 0
}

func validateMutableTrain(train *types.Train) *response {
	if train.NextID != nil {
		resp := errorResponse(
			fmt.Sprintf("Train %d is not the latest train.", train.ID),
			http.StatusBadRequest)
		return &resp
	}

	if train.IsDeployed() {
		resp := errorResponse("Train already deployed.", http.StatusBadRequest)
		return &resp
	}

	if train.IsDeploying() {
		resp := errorResponse("Train is deploying.", http.StatusBadRequest)
		return &resp
	}

	return nil
}

func fetchTrain(r *http.Request) response {
	dataClient := data.NewClient()

	train, resp := parseTrainVars(r, dataClient, true)
	if resp != nil {
		return *resp
	}

	return dataResponse(train)
}

/*
 Protect against race conditions in closing / opening / extending a train.
 These are possible if api calls are made in quick succession.
 We want to ensure a FIFO order of operations to prevent any duplicate work or notifications.
*/
var trainCloseModificationLock sync.Mutex

func closeTrain(r *http.Request) response {
	trainCloseModificationLock.Lock()
	defer trainCloseModificationLock.Unlock()

	dataClient := data.NewClient()

	train, resp := parseTrainVars(r, dataClient, false)
	if resp != nil {
		return *resp
	}

	resp = validateMutableTrain(train)
	if resp != nil {
		return *resp
	}

	if train.Closed {
		return errorResponse("Train already closed.", http.StatusBadRequest)
	}

	err := dataClient.CloseTrain(train, true)
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Error locking train: %v", err),
			http.StatusInternalServerError)
	}

	authedUser := r.Context().Value("user").(*types.User)

	messagingService := messaging.GetService()
	messagingService.TrainClosed(train, authedUser)

	deployIfReady(dataClient, messagingService, train)

	clearLatestTrainCache()

	return dataResponse(struct {
		Closed           bool `json:"closed"`
		ScheduleOverride bool `json:"schedule_override"`
	}{
		true,
		true,
	})
}

func openTrain(r *http.Request) response {
	trainCloseModificationLock.Lock()
	defer trainCloseModificationLock.Unlock()

	dataClient := data.NewClient()

	train, resp := parseTrainVars(r, dataClient, false)
	if resp != nil {
		return *resp
	}

	resp = validateMutableTrain(train)
	if resp != nil {
		return *resp
	}

	if !train.Closed {
		return errorResponse("Train already opened.", http.StatusBadRequest)
	}

	err := dataClient.OpenTrain(train, true)
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Error opening train: %v", err),
			http.StatusInternalServerError)
	}

	authedUser := r.Context().Value("user").(*types.User)

	messagingService := messaging.GetService()
	messagingService.TrainOpened(train, authedUser)

	codeService := code.GetService()
	phaseService := phase.GetService()
	ticketService := ticket.GetService()
	// Check the branch for any new commits, but don't pass requester
	// because that information is contained in the opened message.
	checkBranch(
		dataClient, codeService, messagingService, phaseService, ticketService,
		train.Branch, nil)

	clearLatestTrainCache()

	return dataResponse(struct {
		Closed           bool `json:"closed"`
		ScheduleOverride bool `json:"schedule_override"`
	}{
		false,
		true,
	})
}

func extendTrain(r *http.Request) response {
	trainCloseModificationLock.Lock()
	defer trainCloseModificationLock.Unlock()

	dataClient := data.NewClient()

	train, resp := parseTrainVars(r, dataClient, false)
	if resp != nil {
		return *resp
	}

	resp = validateMutableTrain(train)
	if resp != nil {
		return *resp
	}

	scheduleOverride := train.ScheduleOverride
	err := dataClient.OpenTrain(train, scheduleOverride)
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Error opening train: %v", err),
			http.StatusInternalServerError)
	}

	authedUser := r.Context().Value("user").(*types.User)

	codeService := code.GetService()
	messagingService := messaging.GetService()
	phaseService := phase.GetService()
	ticketService := ticket.GetService()
	checkBranch(
		dataClient, codeService, messagingService, phaseService, ticketService,
		train.Branch, authedUser)

	err = dataClient.CloseTrain(train, scheduleOverride)
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Error locking train: %v", err),
			http.StatusInternalServerError)
	}

	clearLatestTrainCache()

	return emptyResponse()
}

func blockTrain(r *http.Request) response {
	dataClient := data.NewClient()

	train, resp := parseTrainVars(r, dataClient, false)
	if resp != nil {
		return *resp
	}

	resp = validateMutableTrain(train)
	if resp != nil {
		return *resp
	}

	if train.Blocked {
		return errorResponse(
			"Train already blocked",
			http.StatusBadRequest)
	}

	err := dataClient.BlockTrain(train, nil)
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Error blocking train: %v", err),
			http.StatusInternalServerError)
	}

	authedUser := r.Context().Value("user").(*types.User)

	messagingService := messaging.GetService()
	messagingService.TrainBlocked(train, authedUser)

	clearLatestTrainCache()

	return emptyResponse()
}

func unblockTrain(r *http.Request) response {
	dataClient := data.NewClient()

	train, resp := parseTrainVars(r, dataClient, false)
	if resp != nil {
		return *resp
	}

	resp = validateMutableTrain(train)
	if resp != nil {
		return *resp
	}

	if !train.Blocked {
		return errorResponse(
			"Train is not blocked",
			http.StatusBadRequest)
	}

	err := dataClient.UnblockTrain(train)
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Error unblocking train: %v", err),
			http.StatusInternalServerError)
	}

	authedUser := r.Context().Value("user").(*types.User)

	messagingService := messaging.GetService()
	messagingService.TrainUnblocked(train, authedUser)

	deployIfReady(dataClient, messagingService, train)

	clearLatestTrainCache()

	return emptyResponse()
}

func cancelTrain(r *http.Request) response {
	dataClient := data.NewClient()

	train, resp := parseTrainVars(r, dataClient, false)
	if resp != nil {
		return *resp
	}

	if train.Done {
		return errorResponse(
			"Train is already done",
			http.StatusBadRequest)
	}

	err := dataClient.CancelTrain(train)
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Error cancelling train: %v", err),
			http.StatusInternalServerError)
	}

	authedUser := r.Context().Value("user").(*types.User)

	messagingService := messaging.GetService()
	messagingService.TrainCancelled(train, authedUser)

	if train.NextID != nil {
		latestTrain, err := dataClient.LatestTrain()
		if err != nil {
			return errorResponse(
				fmt.Sprintf("Error getting latest train: %v", err),
				http.StatusInternalServerError)
		}

		// Now that this train is cancelled,
		// check if the latest train can be deployed.
		go deployIfReady(data.NewClient(), messagingService, latestTrain)
	}

	clearLatestTrainCache()

	return emptyResponse()
}

func rollbackTrain(r *http.Request) response {
	dataClient := data.NewClient()

	train, resp := parseTrainVars(r, dataClient, false)
	if resp != nil {
		return *resp
	}

	if !train.CanRollback {
		return errorResponse(
			"Train cannot be rolled back",
			http.StatusBadRequest)
	}

	if settings.GetJenkinsRollbackJob() == "" {
		return errorResponse(
			"No rollback job configured",
			http.StatusBadRequest)
	}

	authedUser := r.Context().Value("user").(*types.User)

	messagingService := messaging.GetService()
	messagingService.RollbackInitiated(train, authedUser)

	params := make(map[string]string)
	params["TRAIN_ID"] = strconv.FormatUint(train.ID, 10)
	params["BRANCH"] = train.Branch
	params["SHA"] = train.HeadSHA
	params["CONDUCTOR_HOSTNAME"] = settings.GetHostname()
	params["BUILD_USER"] = authedUser.Name

	latestTrain, err := dataClient.LatestTrain()
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Error getting latest train: %v", err),
			http.StatusInternalServerError)
	}

	// Cancel a deploying train, and block the latest non-deploying train.
	if !latestTrain.Done {
		if latestTrain.IsDeploying() {
			err = dataClient.CancelTrain(latestTrain)
			if err != nil {
				return errorResponse(
					fmt.Sprintf("Error cancelling latest train: %v", err),
					http.StatusInternalServerError)
			}
		} else if !latestTrain.Blocked {
			blockedReason := fmt.Sprintf("rollback by %s", authedUser.Name)
			err := dataClient.BlockTrain(latestTrain, &blockedReason)
			if err != nil {
				return errorResponse(
					fmt.Sprintf("Error blocking latest train: %v", err),
					http.StatusInternalServerError)
			}

			messagingService.TrainBlocked(latestTrain, nil)
		}
	}

	if !latestTrain.PreviousTrainDone {
		previousTrain, err := dataClient.Train(*latestTrain.PreviousID)
		if err != nil {
			return errorResponse(
				fmt.Sprintf("Error getting previous train: %v", err),
				http.StatusInternalServerError)
		}

		err = dataClient.CancelTrain(previousTrain)
		if err != nil {
			return errorResponse(
				fmt.Sprintf("Error cancelling previous train: %v", err),
				http.StatusInternalServerError)
		}

		messagingService.TrainCancelled(previousTrain, nil)
	}

	messagingService.RollbackInfo(authedUser)

	err = build.Jenkins().TriggerJob(settings.GetJenkinsRollbackJob(), params)
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Error triggering rollback job: %v", err),
			http.StatusInternalServerError)
	}

	clearLatestTrainCache()

	return emptyResponse()
}

func checkTrainLock(
	dataClient data.Client,
	codeService code.Service,
	messagingService messaging.Service,
	phaseService phase.Service,
	ticketService ticket.Service) {

	trainCloseModificationLock.Lock()
	defer trainCloseModificationLock.Unlock()

	latestTrain, err := dataClient.LatestTrain()
	if err != nil {
		logger.Error("Error getting latest train: %v", err)
		return
	}

	if latestTrain == nil {
		return
	}

	if latestTrain.IsDeploying() ||
		latestTrain.IsDeployed() ||
		latestTrain.ScheduleOverride {
		return
	}

	mode, err := dataClient.Mode()
	if err != nil {
		logger.Error("Error getting mode: %v", err)
		return
	}
	if mode.IsManualMode() {
		return
	}

	closeable, err := dataClient.IsTrainAutoCloseable(latestTrain)
	if err != nil {
		logger.Error("Error getting IsTrainCloseable: %v", err)
		return
	}
	if closeable && !latestTrain.Closed {
		err = dataClient.CloseTrain(latestTrain, false)
		if err != nil {
			logger.Error("Error locking train: %v", err)
			return
		}

		deployIfReady(dataClient, messagingService, latestTrain)

		messagingService.TrainClosed(latestTrain, nil)

		clearLatestTrainCache()
	} else if !closeable && latestTrain.Closed {
		err = dataClient.OpenTrain(latestTrain, false)
		if err != nil {
			logger.Error("Error opening train: %v", err)
			return
		}

		messagingService.TrainOpened(latestTrain, nil)

		checkBranch(
			dataClient, codeService, messagingService, phaseService, ticketService,
			latestTrain.Branch, nil)

		clearLatestTrainCache()
	}
}
