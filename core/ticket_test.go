// +build data,ticket

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Nextdoor/conductor/services/code"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/services/messaging"
	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/services/ticket"
	"github.com/Nextdoor/conductor/shared/types"
)

const (
	branch   = "foobar"
	email1   = "test@example.com"
	email2   = "test2@example.com"
	message1 = "commit number 1"
	message2 = "a second commit"
	message3 = "the third commit message"
	sha1     = "0"
	sha2     = "1"
	sha3     = "2"
)

func TestSyncTickets(t *testing.T) {
	dataClient := data.NewClient()
	codeService := &code.CodeServiceMock{}
	messagingService := messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}
	// This is a real ticketing system (JIRA) test.
	ticketService := ticket.GetService()
	authorName := ticket.DefaultAccountID
	testCommits := []*types.Commit{
		{AuthorEmail: email1, AuthorName: authorName, SHA: sha1, Message: message1},
		{AuthorEmail: email1, AuthorName: authorName, SHA: sha2, Message: message2},
		{AuthorEmail: email2, AuthorName: authorName, SHA: sha3, Message: message3}}

	train, err := dataClient.CreateTrain(branch, testData.User, testCommits)
	assert.NoError(t, err)

	err = dataClient.StartPhase(train.ActivePhases.Verification)
	assert.NoError(t, err)

	// Create some tickets in DB and ticket service.
	newTickets, err := ticketService.CreateTickets(train, testCommits)
	assert.NoError(t, err)
	assert.Len(t, newTickets, 2)
	dataClient.WriteTickets(newTickets)

	// Close them only in the remote ticket service.
	err = ticketService.CloseTickets(newTickets)
	assert.NoError(t, err)

	// Rely on syncTickets to update the database state from ticket service.
	syncTickets(dataClient, codeService, messagingService, phaseService, ticketService)

	latestTrain, err := dataClient.LatestTrain()
	assert.NoError(t, err)
	// Expect that it has been marked as closed in the DB
	assert.True(t, latestTrain.Tickets[0].ClosedAt.HasValue())

	// Clean up
	err = ticketService.DeleteTickets(train)
	assert.NoError(t, err)
}
