// +build data

package core

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Nextdoor/conductor/services/code"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/services/messaging"
	"github.com/Nextdoor/conductor/services/phase"
	"github.com/Nextdoor/conductor/services/ticket"
	"github.com/Nextdoor/conductor/shared/settings"
	"github.com/Nextdoor/conductor/shared/types"
)

type TrainDeliveredCall struct {
	train   *types.Train
	commits []*types.Commit
	tickets []*types.Ticket
}

type SendCall struct {
	text string
}

type SendDirectCall struct {
	name  string
	email string
	text  string
}

func TestStartPhaseVerification(t *testing.T) {
	_, testData := setup(t)
	dataClient := data.NewClient()
	codeService := &code.CodeServiceMock{}
	var trainDeliveredCalls []TrainDeliveredCall
	messagingService := messaging.MessagingServiceMock{
		TrainDeliveredMock: func(train *types.Train, commits []*types.Commit, tickets []*types.Ticket) {
			trainDeliveredCalls = append(
				trainDeliveredCalls, TrainDeliveredCall{train, commits, tickets})
		},
	}
	phaseService := &phase.PhaseServiceMock{}
	ticketService := &ticket.TicketServiceMock{}
	oldVerificationPhaseStartTime := testData.Train.ActivePhases.Verification.StartedAt
	startPhase(
		dataClient, codeService, messagingService, phaseService, ticketService,
		testData.Train.ActivePhases.Verification, testData.User)
	train, err := dataClient.LatestTrain()
	assert.NoError(t, err)
	// Phase is started.
	assert.NotEqual(t, oldVerificationPhaseStartTime, train.ActivePhases.Verification.StartedAt)
	var expectedTickets []*types.Ticket
	// Delivery messaging is triggered.
	assert.Equal(t,
		[]TrainDeliveredCall{
			{testData.Train, testData.Train.Commits, expectedTickets},
		},
		trainDeliveredCalls)
}

func TestCompletePhaseOutOfOrder(t *testing.T) {
	_, testData := setup(t)
	dataClient := data.NewClient()
	codeService := &code.CodeServiceMock{}
	messagingService := messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}
	ticketService := &ticket.TicketServiceMock{}
	err := dataClient.StartPhase(testData.Train.ActivePhases.Verification)
	assert.NoError(t, err)
	// Complete verification phase. Should not be complete afterwards unless delivery is started and completed.
	checkPhaseCompletion(dataClient, codeService, messagingService, phaseService, ticketService,
		testData.Train.ActivePhases.Verification)
	train, err := dataClient.LatestTrain()
	assert.NoError(t, err)
	// Phase is not complete.
	assert.False(t, train.ActivePhases.Verification.IsComplete())
	// Complete delivery phase, then try again. Should successfully complete this time.
	err = dataClient.CompletePhase(train.ActivePhases.Delivery)
	assert.NoError(t, err)
	checkPhaseCompletion(dataClient, codeService, messagingService, phaseService, ticketService,
		train.ActivePhases.Verification)
	train, err = dataClient.LatestTrain()
	assert.NoError(t, err)
	assert.True(t, train.ActivePhases.Verification.IsComplete())
}

func TestCompletePhaseBeforeStart(t *testing.T) {
	_, testData := setup(t)
	dataClient := data.NewClient()
	codeService := &code.CodeServiceMock{}
	messagingService := messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}
	ticketService := &ticket.TicketServiceMock{}
	// Try to complete delivery phase. Should not be complete afterwards because delivery hasn't been started yet.
	checkPhaseCompletion(dataClient, codeService, messagingService, phaseService, ticketService,
		testData.Train.ActivePhases.Delivery)
	train, err := dataClient.LatestTrain()
	assert.NoError(t, err)
	// Phase is not complete.
	assert.False(t, train.ActivePhases.Delivery.IsComplete())
	// Start delivery phase, then try again. Should successfully complete this time.
	err = dataClient.StartPhase(train.ActivePhases.Delivery)
	assert.NoError(t, err)
	checkPhaseCompletion(dataClient, codeService, messagingService, phaseService, ticketService,
		train.ActivePhases.Delivery)
	train, err = dataClient.LatestTrain()
	assert.NoError(t, err)
	assert.True(t, train.ActivePhases.Delivery.IsComplete())
}

func TestUnverifiedPhaseUncomplete(t *testing.T) {
	_, testData := setup(t)
	dataClient := data.NewClient()
	codeService := &code.CodeServiceMock{}

	trainVerifiedCalled := false
	trainUnverifiedCalled := false
	messagingService := messaging.MessagingServiceMock{
		TrainVerifiedMock: func(*types.Train) {
			trainVerifiedCalled = true
		},
		TrainUnverifiedMock: func(*types.Train) {
			trainUnverifiedCalled = true
		},
	}

	phaseService := &phase.PhaseServiceMock{}
	ticketService := &ticket.TicketServiceMock{}

	// Start and complete delivery and verification phases.
	err := dataClient.StartPhase(testData.Train.ActivePhases.Delivery)
	assert.NoError(t, err)
	err = dataClient.CompletePhase(testData.Train.ActivePhases.Delivery)
	assert.NoError(t, err)

	err = dataClient.StartPhase(testData.Train.ActivePhases.Verification)
	assert.NoError(t, err)
	err = dataClient.CompletePhase(testData.Train.ActivePhases.Verification)
	assert.NoError(t, err)

	// Add an unverified ticket to the train.
	newTickets := []*types.Ticket{{Commits: nil, Train: testData.Train}}
	err = dataClient.WriteTickets(newTickets)
	assert.NoError(t, err)

	testData.Train.Tickets = newTickets

	// Check verification phase completion. Should uncomplete verification phase and call messaging.TrainUnverified.
	checkPhaseCompletion(dataClient, codeService, messagingService, phaseService, ticketService,
		testData.Train.ActivePhases.Verification)
	train, err := dataClient.LatestTrain()
	assert.NoError(t, err)

	assert.False(t, train.ActivePhases.Verification.IsComplete())
	assert.False(t, trainVerifiedCalled)
	assert.True(t, trainUnverifiedCalled)

	// Verify the ticket and try again.
	newTickets[0].ClosedAt = types.Time{time.Now()}
	err = dataClient.UpdateTickets(newTickets)
	assert.NoError(t, err)

	// Reset called values.
	trainVerifiedCalled = false
	trainUnverifiedCalled = false

	train.Tickets = newTickets

	// Check verification phase completion. Should now complete verification phase and call messaging.TrainVerified.
	checkPhaseCompletion(dataClient, codeService, messagingService, phaseService, ticketService,
		train.ActivePhases.Verification)
	train, err = dataClient.LatestTrain()
	assert.NoError(t, err)

	assert.True(t, train.ActivePhases.Verification.IsComplete())
	assert.True(t, trainVerifiedCalled)
	assert.False(t, trainUnverifiedCalled)
}

// Tests logic for which commits and tickets are passed to the messaging service.
func TestDeliveryFinishedMessaging(t *testing.T) {
	oldCommit := &types.Commit{
		Message:     "previous phase group commit",
		AuthorName:  "Author Name",
		AuthorEmail: "author@email.com",
		URL:         "https://github.com",
		SHA:         "111",
	}
	ticketedCommit := &types.Commit{
		Message:     "ticketed commit",
		AuthorName:  "Author Name",
		AuthorEmail: "author@email.com",
		URL:         "https://github.com",
		SHA:         "222",
	}
	noVerifyCommit := &types.Commit{
		Message:     "no verify commit [no-verify]",
		AuthorName:  "Author Name",
		AuthorEmail: "author@email.com",
		URL:         "https://github.com",
		SHA:         "333",
	}
	vanillaCommit := &types.Commit{
		Message:     "vanilla commit",
		AuthorName:  "Author Name",
		AuthorEmail: "author@email.com",
		URL:         "https://github.com",
		SHA:         "444",
	}

	allCommits := []*types.Commit{oldCommit, ticketedCommit, noVerifyCommit, vanillaCommit}

	dataClient := data.NewClient()
	dataClient.CreateTrain("branch", testData.User, []*types.Commit{oldCommit})
	train, _ := dataClient.LatestTrain()
	dataClient.ExtendTrain(train, train.Engineer, []*types.Commit{
		ticketedCommit, noVerifyCommit, vanillaCommit})

	user, err := dataClient.ReadOrCreateUser("test_user", "test_email")
	assert.NoError(t, err)

	var sendCalls []SendCall
	var sendDirectCalls []SendDirectCall
	messagingService := &messaging.Messenger{
		Engine: &messaging.EngineMock{
			SendMock: func(text string) {
				sendCalls = append(sendCalls, SendCall{text})
			},
			SendDirectMock: func(name string, email string, text string) {
				sendDirectCalls = append(sendDirectCalls, SendDirectCall{name, email, text})
			},
		},
	}

	var vanillaCommitTicket = &types.Ticket{Commits: []*types.Commit{vanillaCommit}, Train: train}
	type CreateTicketsCall struct {
		train   *types.Train
		commits []*types.Commit
	}

	var ticketedCommitTicket = &types.Ticket{Commits: []*types.Commit{ticketedCommit}, Train: train}
	var oldCommitTicket = &types.Ticket{Commits: []*types.Commit{oldCommit}, Train: train}
	train.Tickets = []*types.Ticket{ticketedCommitTicket, oldCommitTicket}
	train.LastDeliveredSHA = &oldCommit.SHA

	var createTicketsCalls []CreateTicketsCall
	ticketService := &ticket.TicketServiceMock{
		CreateTicketsMock: func(train *types.Train, commits []*types.Commit) ([]*types.Ticket, error) {
			createTicketsCalls = append(
				createTicketsCalls, CreateTicketsCall{train, commits})
			return []*types.Ticket{vanillaCommitTicket}, nil
		},
	}
	phaseService := &phase.PhaseServiceMock{}
	codeService := &code.CodeServiceMock{}
	startPhase(
		dataClient, codeService, messagingService, phaseService, ticketService,
		train.ActivePhases.Verification, user)
	assert.Equal(t, 1, len(createTicketsCalls))
	assert.Equal(t, createTicketsCalls[0].train, train)
	assert.Equal(t, 1, len(createTicketsCalls))
	assert.Equal(t, createTicketsCalls[0].train, train)

	// Ensure tickets are created for the right commits - no-verify commits,
	// and commits with pre-existing tickets - are excluded.
	SHAsForTickets := make(map[string]struct{})
	for _, commit := range createTicketsCalls[0].commits {
		SHAsForTickets[commit.SHA] = struct{}{}
	}
	assert.Equal(t, map[string]struct{}{
		vanillaCommit.SHA: struct{}{}}, SHAsForTickets)

	// Ensure staging channel is messaged about the right commits.
	// No-verify commits are mentioned directly but not in the staging channel.
	stagingChannelMessage := sendCalls[1].text
	assert.False(t, strings.Contains(stagingChannelMessage, oldCommit.SHA))
	assert.False(t, strings.Contains(stagingChannelMessage, ticketedCommit.SHA))
	assert.False(t, strings.Contains(stagingChannelMessage, noVerifyCommit.SHA))
	assert.True(t, strings.Contains(stagingChannelMessage, vanillaCommit.SHA))

	SHAsInDirectMessages := make(map[string]struct{})
	for _, sendDirectCall := range sendDirectCalls {
		for _, commit := range allCommits {
			if strings.Contains(sendDirectCall.text, commit.SHA) {
				SHAsInDirectMessages[commit.SHA] = struct{}{}
			}
		}
	}
	// Ensure direct messages are sent about the right commits.
	assert.Equal(t,
		map[string]struct{}{
			vanillaCommit.SHA:  struct{}{},
			noVerifyCommit.SHA: struct{}{},
			ticketedCommit.SHA: struct{}{}},
		SHAsInDirectMessages)
}

// Tests messaging logic for users in the no-staging whitelist.
func TestDeliveryFinishedMessagingForNoStagingUser(t *testing.T) {
	vanillaCommit := &types.Commit{
		Message:     "no staging whitelist commit",
		AuthorName:  "Author Name",
		AuthorEmail: "no-staging@email.com",
		URL:         "https://github.com",
		SHA:         "aaa",
	}
	needsStagingOverrideCommit := &types.Commit{
		Message:     "no staging whitelist manual override commit [needs-staging]",
		AuthorName:  "Author Name",
		AuthorEmail: "no-staging@email.com",
		URL:         "https://github.com",
		SHA:         "bbb",
	}
	noVerifyCommit := &types.Commit{
		Message:     "no verify commit [no-verify]",
		AuthorName:  "Author Name",
		AuthorEmail: "no-staging@email.com",
		URL:         "https://github.com",
		SHA:         "ccc",
	}

	allCommits := []*types.Commit{noVerifyCommit, vanillaCommit, needsStagingOverrideCommit}

	// Put the commit author on the no staging whitelist.
	settings.CustomizeNoStagingUsers([]string{"no-staging@email.com"})

	dataClient := data.NewClient()
	train, _ := dataClient.CreateTrain("my_branch", testData.User, allCommits)

	user, err := dataClient.ReadOrCreateUser("test_user", "test_email")
	assert.NoError(t, err)

	var sendCalls []SendCall
	var sendDirectCalls []SendDirectCall
	messagingService := &messaging.Messenger{
		Engine: &messaging.EngineMock{
			SendMock: func(text string) {
				sendCalls = append(sendCalls, SendCall{text})
			},
			SendDirectMock: func(name string, email string, text string) {
				sendDirectCalls = append(sendDirectCalls, SendDirectCall{name, email, text})
			},
		},
	}

	var needsStagingOverrideCommitTicket = &types.Ticket{
		Commits: []*types.Commit{needsStagingOverrideCommit}, Train: train}
	type CreateTicketsCall struct {
		train   *types.Train
		commits []*types.Commit
	}

	var createTicketsCalls []CreateTicketsCall
	ticketService := &ticket.TicketServiceMock{
		CreateTicketsMock: func(train *types.Train, commits []*types.Commit) ([]*types.Ticket, error) {
			createTicketsCalls = append(
				createTicketsCalls, CreateTicketsCall{train, commits})
			return []*types.Ticket{needsStagingOverrideCommitTicket}, nil
		},
	}
	phaseService := &phase.PhaseServiceMock{}
	codeService := &code.CodeServiceMock{}
	startPhase(
		dataClient, codeService, messagingService, phaseService, ticketService,
		train.ActivePhases.Verification, user)

	// People on the no-staging whitelist don't get tickets unless their commit is marked [needs-staging].
	SHAsForTickets := make(map[string]struct{})
	for _, commit := range createTicketsCalls[0].commits {
		SHAsForTickets[commit.SHA] = struct{}{}
	}
	assert.Equal(t, map[string]struct{}{
		needsStagingOverrideCommit.SHA: struct{}{}}, SHAsForTickets)

	stagingChannelMessage := sendCalls[1].text

	// Staging channel won't mention no-staging whitelist commits, unless it is marked [needs-staging].
	assert.True(t, strings.Contains(stagingChannelMessage, needsStagingOverrideCommit.SHA))
	assert.False(t, strings.Contains(stagingChannelMessage, vanillaCommit.SHA))
	assert.False(t, strings.Contains(stagingChannelMessage, noVerifyCommit.SHA))

	SHAsInDirectMessages := make(map[string]struct{})
	for _, sendDirectCall := range sendDirectCalls {
		for _, commit := range allCommits {
			if strings.Contains(sendDirectCall.text, commit.SHA) {
				SHAsInDirectMessages[commit.SHA] = struct{}{}
			}
		}
	}
	// People on the no-staging whitelist don't get notified about their commit going to staging
	// unless it is marked [needs-staging].
	assert.Equal(
		t,
		map[string]struct{}{needsStagingOverrideCommit.SHA: struct{}{}},
		SHAsInDirectMessages)
}

func TestDeployableAfterVerification(t *testing.T) {
	_, testData := setup(t)
	dataClient := data.NewClient()
	codeService := &code.CodeServiceMock{}
	messagingService := &messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}
	ticketService := &ticket.TicketServiceMock{}

	err := dataClient.CloseTrain(testData.Train, true)
	assert.NoError(t, err)

	// Start and complete delivery phase.
	err = dataClient.StartPhase(testData.Train.ActivePhases.Delivery)
	assert.NoError(t, err)
	err = dataClient.CompletePhase(testData.Train.ActivePhases.Delivery)
	assert.NoError(t, err)

	startPhase(dataClient, codeService, messagingService, phaseService, ticketService,
		testData.Train.ActivePhases.Verification, nil)

	// Train should be deployable.
	testData.Train.PreviousTrainDone = true
	assert.True(t, testData.Train.IsDeployable())
}

// TODO: TestRestartPhase
// TODO: TestStartPhase
// TODO: TestHandlePhaseCompletion
