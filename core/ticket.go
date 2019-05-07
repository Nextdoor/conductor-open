package core

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Nextdoor/conductor/services/code"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/services/messaging"
	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/services/ticket"
	"github.com/Nextdoor/conductor/shared/datadog"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/types"
)

var ticketModificationLock sync.Mutex

func ticketEndpoints() []endpoint {
	return []endpoint{
		newEp("/api/ticket/open", get, openTicketsEndpoint),
	}
}

func openTicketsEndpoint(_ *http.Request) response {
	dataClient := data.NewClient()
	latestTrain, err := dataClient.LatestTrain()
	if err != nil {
		return errorResponse(err.Error(), http.StatusInternalServerError)
	}
	if err != nil {
		return errorResponse(err.Error(), http.StatusInternalServerError)
	}
	return dataResponse(latestTrain.Tickets)
}

// Synchronize train's local ticket state
// with remote ticket service state.
func syncTickets(
	dataClient data.Client,
	codeService code.Service,
	messagingService messaging.Service,
	phaseService phase.Service,
	ticketService ticket.Service) {
	ticketModificationLock.Lock()
	defer ticketModificationLock.Unlock()

	latestTrain, err := dataClient.LatestTrain()
	if err != nil {
		logger.Error("Error getting train: %v", err)
		return
	}

	if latestTrain == nil {
		return
	}

	if latestTrain.IsDeploying() || latestTrain.IsDeployed() {
		return
	}

	newTickets, updatedTickets, err := ticketService.SyncTickets(latestTrain)
	err = dataClient.WriteTickets(newTickets)
	if err != nil {
		logger.Error("Error writing tickets: %v", err)
		return
	}
	err = dataClient.UpdateTickets(updatedTickets)
	if err != nil {
		logger.Error("Error updating tickets: %v", err)
		return
	}

	datadog.Count("ticket.count", len(newTickets), latestTrain.DatadogTags())
	for _, updatedTicket := range updatedTickets {
		if updatedTicket.ClosedAt.HasValue() || updatedTicket.DeletedAt.HasValue() {
			var finished time.Time
			if updatedTicket.ClosedAt.HasValue() {
				finished = updatedTicket.ClosedAt.Value
			} else {
				finished = updatedTicket.DeletedAt.Value
			}
			duration := finished.Sub(updatedTicket.CreatedAt.Value)
			tags := latestTrain.DatadogTags()
			tags = append(tags, fmt.Sprintf("ticket_user:%s", updatedTicket.AssigneeEmail))
			datadog.Gauge("ticket.duration", duration.Seconds(), tags)
		}
	}

	switch latestTrain.ActivePhase {
	case types.Verification:
		checkPhaseCompletion(
			dataClient, codeService, messagingService, phaseService, ticketService,
			latestTrain.ActivePhases.Verification)
	case types.Deploy:
		if latestTrain.ActivePhases.Deploy.StartedAt.HasValue() {
			logger.Error("A ticket was updated, but the deploy phase has already begun: %v", err)
			return
		}
		err = dataClient.UncompletePhase(latestTrain.ActivePhases.Verification)
		if err != nil {
			logger.Error("Error uncompleting verification phase: %v", err)
			return
		}
		checkPhaseCompletion(
			dataClient, codeService, messagingService, phaseService, ticketService,
			latestTrain.ActivePhases.Verification)
	}
}
