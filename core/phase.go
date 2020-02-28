package core

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"

	"github.com/Nextdoor/conductor/services/code"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/services/messaging"
	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/services/ticket"
	"github.com/Nextdoor/conductor/shared/datadog"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/settings"
	"github.com/Nextdoor/conductor/shared/types"
)

func phaseEndpoints() []endpoint {
	return []endpoint{
		newEp("/api/train/{train_id:[0-9]+}/phase/{phase_type:delivery|verification|deploy}/restart", post, triggerPhaseRestart),
	}
}

func triggerPhaseRestart(r *http.Request) response {
	dataClient := data.NewClient()

	vars := mux.Vars(r)
	trainID, err := strconv.ParseUint(vars["train_id"], 10, 64)
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Bad train_id value: %s", vars["train_id"]),
			http.StatusBadRequest)
	}

	phaseTypeStr := vars["phase_type"]
	phaseType, err := types.PhaseTypeFromString(
		strings.ToLower(phaseTypeStr))
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Bad `phase_type` value: %v", err),
			http.StatusBadRequest)
	}

	latestTrain, err := dataClient.LatestTrain()
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Problem getting the train: %v", err),
			http.StatusInternalServerError)
	}

	if trainID != latestTrain.ID && trainID+1 != latestTrain.ID {
		return errorResponse(
			fmt.Sprintf("Cannot restart phase %s on train %d - the active train is %d. "+
				"Phases can only be restarted on the latest train or the previous train.",
				phaseType, trainID, latestTrain.ID),
			http.StatusBadRequest)
	}

	targetTrain, err := dataClient.Train(trainID)
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Problem getting the train: %v", err),
			http.StatusInternalServerError)
	}

	phaseToRestart := targetTrain.Phase(phaseType)
	if phaseToRestart.IsComplete() {
		return errorResponse(
			"This phase has already completed.",
			http.StatusBadRequest)
	}

	replacedPhase, err := dataClient.ReplacePhase(phaseToRestart)
	if err != nil {
		return errorResponse(
			"Error restarting the phase.",
			http.StatusInternalServerError)
	}

	authedUser := r.Context().Value("user").(*types.User)

	codeService := code.GetService()
	messagingService := messaging.GetService()
	phaseService := phase.GetService()
	ticketService := ticket.GetService()
	go startPhase(data.NewClient(), codeService, messagingService, phaseService, ticketService, replacedPhase, authedUser)

	return emptyResponse()
}

func startPhase(
	dataClient data.Client,
	codeService code.Service,
	messagingService messaging.Service,
	phaseService phase.Service,
	ticketService ticket.Service,
	phaseToStart *types.Phase,
	user *types.User) {

	logger.Info("Starting phase %s for train %v (%s).\n\n%+v",
		phaseToStart.Type, phaseToStart.Train.ID, phaseToStart.Train.HeadSHA, phaseToStart.Train)

	// Pre-phase actions
	switch phaseToStart.Type {
	case types.Verification:
		logger.Info("Handling notification and ticket creation for Phase %v", phaseToStart.ID)
		err := phaseGroupDelivered(
			dataClient, messagingService, ticketService, phaseToStart.Train, phaseToStart.PhaseGroup)
		if err != nil {
			logger.Error("ErrorPhase: %v", err)
			err = dataClient.ErrorPhase(phaseToStart, err)
			if err != nil {
				logger.Error("%v", err)
			}
		}
	}

	err := dataClient.StartPhase(phaseToStart)
	if err != nil {
		logger.Error("%v", err)
		return
	}

	datadog.Incr("phase.start", phaseToStart.DatadogTags())

	switch phaseToStart.Type {
	case types.Deploy:
		// Check for any commits waiting for a train.
		checkBranch(
			dataClient, codeService, messagingService, phaseService, ticketService,
			phaseToStart.Train.Branch, nil)
	}

	err = phaseService.Start(phaseToStart.Type,
		phaseToStart.Train.ID,
		phaseToStart.PhaseGroup.Delivery.ID,
		phaseToStart.PhaseGroup.Verification.ID,
		phaseToStart.PhaseGroup.Deploy.ID,
		phaseToStart.Train.Branch, phaseToStart.PhaseGroup.HeadSHA,
		user)
	if err != nil {
		logger.Error("ErrorPhase: %v", err)
		err = dataClient.ErrorPhase(phaseToStart, err)
		if err != nil {
			logger.Error("%v", err)
		}
	}

	clearLatestTrainCache()

	checkPhaseCompletion(dataClient, codeService, messagingService, phaseService, ticketService, phaseToStart)
}

// The phase group was delivered to staging.
// Handle notification and ticket creation for these commits.
func phaseGroupDelivered(
	dataClient data.Client,
	messagingService messaging.Service,
	ticketService ticket.Service,
	train *types.Train,
	phaseGroup *types.PhaseGroup) error {

	ticketModificationLock.Lock()
	defer ticketModificationLock.Unlock()

	newCommitsNeedingTickets := train.NewCommitsNeedingTickets(phaseGroup.HeadSHA, settings.NoStagingVerification)
	var tickets []*types.Ticket
	var err error
	logger.Info("There are %v commits that need tickets", len(newCommitsNeedingTickets))
	if len(newCommitsNeedingTickets) > 0 {
		tickets, err = ticketService.CreateTickets(train, newCommitsNeedingTickets)
		if err != nil {
			return err
		}

		// Store tickets.
		err = dataClient.WriteTickets(tickets)
		if err != nil {
			return err
		}

		datadog.Count("ticket.count", len(tickets), train.DatadogTags())

		// Add these tickets to the train for anything that'll immediate check them.
		// There might be existing tickets, so append first.
		train.Tickets = append(train.Tickets, tickets...)
	}

	err = dataClient.LoadLastDeliveredSHA(train)
	if err != nil {
		return err
	}

	var newCommits []*types.Commit
	if train.LastDeliveredSHA == nil {
		newCommits = train.CommitsSince(phaseGroup.HeadSHA)
	} else {
		newCommits = train.CommitsBetween(phaseGroup.HeadSHA, *train.LastDeliveredSHA)
	}
	messagingService.TrainDelivered(train, newCommits, tickets)

	return nil
}

var phaseCompletionLock sync.Mutex

func checkPhaseCompletion(
	dataClient data.Client,
	codeService code.Service,
	messagingService messaging.Service,
	phaseService phase.Service,
	ticketService ticket.Service,
	targetPhase *types.Phase) {

	phaseCompletionLock.Lock()
	defer phaseCompletionLock.Unlock()

	train := targetPhase.Train

	var extraChecks []phase.Completeable
	switch targetPhase.Type {
	case types.Verification:
		for i := range train.Tickets {
			extraChecks = append(extraChecks, train.Tickets[i])
		}
	}

	phaseCompletedPreviously := targetPhase.IsComplete()
	phaseCurrentlyCompleted := phase.IsComplete(
		targetPhase.Type, targetPhase.Jobs.CompletedNames(), extraChecks...)

	logger.Info("Checking phase completion for phase %v, train %v (%v). "+
		"It has %d tickets, which will trigger %d extra completion checks.\n\nTrain: %+v",
		targetPhase.Type, train.ID, train.HeadSHA, len(train.Tickets), len(extraChecks), train)

	if phaseCompletedPreviously && phaseCurrentlyCompleted {
		// Completion already handled.
		return
	}

	if phaseCompletedPreviously && !phaseCurrentlyCompleted {
		// Phase is no longer completed - uncomplete it.
		datadog.Incr("phase.uncomplete", targetPhase.DatadogTags())
		err := dataClient.UncompletePhase(targetPhase)
		if err != nil {
			logger.Error("Error uncompleting phase: %v", err)
		} else {
			if targetPhase.Type == types.Verification {
				messagingService.TrainUnverified(train)
			}
		}
		return
	}

	if !phaseCurrentlyCompleted {
		// Phase is not complete.
		return
	}

	if !targetPhase.EarlierPhasesComplete() {
		// Cannot complete a phase if the earlier phases weren't completed yet.
		return
	}

	if !targetPhase.StartedAt.HasValue() {
		// Cannot complete a phase before it starts.
		return
	}

	err := dataClient.CompletePhase(targetPhase)
	if err != nil {
		logger.Error("Error completing phase: %v", err)
		return
	}

	datadog.Incr("phase.complete", targetPhase.DatadogTags())
	duration := targetPhase.CompletedAt.Value.Sub(targetPhase.StartedAt.Value)
	datadog.Gauge("phase.duration", duration.Seconds(), targetPhase.DatadogTags())

	logger.Info("Phase %s was completed for train %v (%s). "+
		"It had %d tickets causing %d extra checks.\n\n%+v",
		targetPhase.Type, train.ID, train.HeadSHA, len(train.Tickets), len(extraChecks), train)

	// Post-phase actions
	switch targetPhase.Type {
	case types.Delivery:
		go startPhase(
			data.NewClient(), codeService, messagingService, phaseService, ticketService,
			targetPhase.PhaseGroup.Verification, nil)
	case types.Verification:
		if targetPhase.IsInActivePhaseGroup() {
			// We only send this message if this is the most recent verification phase.
			// Otherwise, the train isn't fully verified yet.
			messagingService.TrainVerified(train)
		}
		go deployIfReady(data.NewClient(), messagingService, train)
	case types.Deploy:
		err = dataClient.DeployTrain(train)
		if err != nil {
			logger.Error("Error deploying train: %v", err)
			return
		}

		duration := train.DeployedAt.Value.Sub(train.CreatedAt.Value)
		datadog.Gauge("train.deploy.lifetime.all_hours", duration.Seconds(), train.DatadogTags())

		options, err := dataClient.Options()
		if err != nil {
			logger.Error("Error getting options: %v", err)
		} else {
			regularHoursDuration := options.CloseTimeOverlap(train.CreatedAt.Value, train.DeployedAt.Value)
			afterHoursDuration := duration - regularHoursDuration

			datadog.Gauge("train.deploy.lifetime.regular_hours", regularHoursDuration.Seconds(), train.DatadogTags())
			datadog.Gauge("train.deploy.lifetime.after_hours", afterHoursDuration.Seconds(), train.DatadogTags())
		}

		messagingService.TrainDeployed(train)

		checkBranch(
			dataClient, codeService, messagingService, phaseService, ticketService,
			targetPhase.Train.Branch, nil)

		if train.NextID != nil {
			latestTrain, err := dataClient.LatestTrain()
			if err != nil {
				logger.Error("Error getting latest train: %v", err)
				return
			}

			// Now that this train's deploy is finished,
			// check if the latest train can be deployed.
			go deployIfReady(data.NewClient(), messagingService, latestTrain)
		}
	}
}
