package core

import (
	"net/http"

	"github.com/Nextdoor/conductor/services/code"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/services/messaging"
	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/services/ticket"
	"github.com/Nextdoor/conductor/shared/logger"
)

func codeEndpoints() []endpoint {
	return []endpoint{
		newOpenEp("/api/code/webhook", post, codeWebhook),
	}
}

func codeWebhook(r *http.Request) response {
	codeService := code.GetService()
	messagingService := messaging.GetService()
	phaseService := phase.GetService()
	ticketService := ticket.GetService()
	branch, err := codeService.ParseWebhookForBranch(r)
	if err != nil {
		return errorResponse(err.Error(), http.StatusInternalServerError)
	}

	if branch != "" {
		logger.Info("There was a push event to branch %s", branch)
		go checkBranch(
			data.NewClient(), codeService, messagingService, phaseService, ticketService,
			branch, nil)
	}

	return emptyResponse()
}
