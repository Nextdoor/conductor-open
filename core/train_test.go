// +build data

package core

import (
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

var newCommitSHA = "abcdef"
var newCommit = &types.Commit{
	Message:     "New commit",
	AuthorName:  "Author Name",
	AuthorEmail: "author@email.com",
	URL:         "https://github.com",
	SHA:         newCommitSHA,
}

// Be graceful when there are no new commits.
func TestCheckBranchNoNewCommits(t *testing.T) {
	_, testData := setup(t)
	initialHeadSHA := testData.Train.HeadSHA
	codeService := &code.CodeServiceMock{
		CommitsOnBranchAfterMock: func(branch string, head string) ([]*types.Commit, error) {
			return []*types.Commit{}, nil
		},
	}
	dataClient := data.NewClient()
	messagingService := messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}
	ticketService := &ticket.TicketServiceMock{}
	checkBranch(
		dataClient, codeService, messagingService, phaseService, ticketService,
		testData.Train.Branch, testData.User)
	train, _ := dataClient.Train(testData.Train.ID)
	assert.Equal(t, initialHeadSHA, train.HeadSHA)
}

// Train is extended when new commits are added to the branch in the grace period.
func TestCheckBranchExtend(t *testing.T) {
	_, testData := setup(t)
	codeService := &code.CodeServiceMock{
		CommitsOnBranchAfterMock: func(branch string, head string) ([]*types.Commit, error) {
			return []*types.Commit{newCommit}, nil
		},
	}
	dataClient := data.NewClient()
	messagingService := messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}
	ticketService := &ticket.TicketServiceMock{}
	checkBranch(
		dataClient, codeService, messagingService, phaseService, ticketService,
		testData.Train.Branch, testData.User)
	train, _ := dataClient.Train(testData.Train.ID)
	assert.Equal(t, newCommitSHA, train.HeadSHA)
}

// Case when there's never been a train before.
func TestCheckBranchFirstTrain(t *testing.T) {
	codeService := &code.CodeServiceMock{
		CommitsOnBranchMock: func(branch string, max int) ([]*types.Commit, error) {
			return []*types.Commit{newCommit}, nil
		},
	}
	dataClient := data.NewClient()
	messagingService := messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}
	ticketService := &ticket.TicketServiceMock{}
	user, err := dataClient.ReadOrCreateUser("test_user", "test_email")
	assert.NoError(t, err)
	checkBranch(
		dataClient, codeService, messagingService, phaseService, ticketService,
		"master", user)
	train, _ := dataClient.LatestTrain()
	assert.Equal(t, newCommitSHA, train.HeadSHA)
}

// Case when there's never been a train on this branch before.
func TestCheckBranchFirstTrainOnBranch(t *testing.T) {
	_, testData := setup(t)
	codeService := &code.CodeServiceMock{
		CompareRefsMock: func(headSHA string, branch string) ([]*types.Commit, error) {
			return []*types.Commit{newCommit}, nil
		},
	}
	dataClient := data.NewClient()
	messagingService := messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}
	ticketService := &ticket.TicketServiceMock{}
	checkBranch(
		dataClient, codeService, messagingService, phaseService, ticketService,
		"first_train_branch", testData.User)
	train, _ := dataClient.LatestTrain()
	// the commit on the new branch becomes a new train.
	assert.Equal(t, newCommitSHA, train.HeadSHA)
	assert.NotEqual(t, testData.Train.ID, train.ID)
}

// If current train is deploying, this starts a new train.
func TestCheckBranchLatestTrainDeploying(t *testing.T) {
	_, testData := setup(t)
	codeService := &code.CodeServiceMock{
		CommitsOnBranchAfterMock: func(branch string, head string) ([]*types.Commit, error) {
			return []*types.Commit{newCommit}, nil
		},
	}
	messagingService := messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}
	ticketService := &ticket.TicketServiceMock{}
	testData.Train.ActivePhases.Deploy.StartedAt = types.Time{time.Now()}
	dataClient := data.NewClient()
	dataClient.DeployTrain(testData.Train)
	checkBranch(
		dataClient, codeService, messagingService, phaseService, ticketService,
		testData.Train.Branch, testData.User)
	train, _ := dataClient.LatestTrain()
	assert.Equal(t, newCommitSHA, train.HeadSHA)
	assert.NotEqual(t, testData.Train.ID, train.ID)
}

// If current train finished deploying, this starts a new train.
func TestCheckBranchLatestTrainDeployed(t *testing.T) {
	_, testData := setup(t)
	codeService := &code.CodeServiceMock{
		CommitsOnBranchAfterMock: func(branch string, head string) ([]*types.Commit, error) {
			return []*types.Commit{newCommit}, nil
		},
	}
	dataClient := data.NewClient()
	dataClient.DeployTrain(testData.Train)
	dataClient.CompletePhase(testData.Train.ActivePhases.Deploy)
	messagingService := messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}

	closeTrainTicketsCalled := false
	ticketService := &ticket.TicketServiceMock{
		CloseTrainTicketsMock: func(train *types.Train) error {
			closeTrainTicketsCalled = true
			return nil
		},
	}

	checkBranch(
		dataClient, codeService, messagingService, phaseService, ticketService,
		testData.Train.Branch, testData.User)
	assert.Equal(t, true, closeTrainTicketsCalled)

	train, _ := dataClient.LatestTrain()
	assert.Equal(t, newCommitSHA, train.HeadSHA)
	assert.NotEqual(t, testData.Train.ID, train.ID)
}

// Queue commits when there is a locked train on the target branch.
func TestCheckBranchQueueCommits(t *testing.T) {
	_, testData := setup(t)
	codeService := &code.CodeServiceMock{
		CommitsOnBranchAfterMock: func(branch string, head string) ([]*types.Commit, error) {
			return []*types.Commit{newCommit}, nil
		},
	}
	dataClient := data.NewClient()
	dataClient.CloseTrain(testData.Train, false)
	messagingService := messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}
	ticketService := &ticket.TicketServiceMock{}
	checkBranch(
		dataClient, codeService, messagingService, phaseService, ticketService,
		testData.Train.Branch, testData.User)

	// Current train head should be unchanged.
	train, _ := dataClient.LatestTrain()
	assert.Equal(t, train.HeadSHA, train.HeadSHA)
}

// Duplicate the latest train on the branch when switching back from a different branch.
func TestCheckBranchDuplicateTrain(t *testing.T) {
	_, testData := setup(t)
	codeService := &code.CodeServiceMock{
		CommitsOnBranchAfterMock: func(branch string, head string) ([]*types.Commit, error) {
			return []*types.Commit{newCommit}, nil
		},
	}
	dataClient := data.NewClient()
	dataClient.CreateTrain("dup_test_branch", testData.User, []*types.Commit{
		{
			Message:     "Other branch commit",
			AuthorName:  "Author Name",
			AuthorEmail: "author@email.com",
			URL:         "https://github.com",
			SHA:         "a1a1a1",
		},
	})
	messagingService := messaging.MessagingServiceMock{}
	phaseService := &phase.PhaseServiceMock{}

	closeTrainTicketsCalled := false
	ticketService := &ticket.TicketServiceMock{
		CloseTrainTicketsMock: func(train *types.Train) error {
			closeTrainTicketsCalled = true
			return nil
		},
	}

	checkBranch(
		dataClient, codeService, messagingService, phaseService, ticketService,
		testData.Train.Branch, testData.User)
	assert.Equal(t, true, closeTrainTicketsCalled)

	// New train should clone the previous one on this branch, and add the newest commit.
	train, _ := dataClient.LatestTrain()
	assert.NotEqual(t, train.ID, testData.Train.ID)
	assert.Equal(t, train.HeadSHA, newCommitSHA)
	assert.Equal(t, train.TailSHA, testData.Train.TailSHA)
}

// Duplicate the latest train on the branch when switching back from a different branch.
func TestChooseEngineer(t *testing.T) {
	settings.CustomizeRobotUsers([]string{"robot-1", "robot-2"})

	dataClient := data.NewClient()
	engineer, err := chooseEngineer(dataClient, []*types.Commit{})
	assert.NoError(t, err)
	assert.Nil(t, engineer)

	engineer, err = chooseEngineer(dataClient, []*types.Commit{{
		AuthorName:  "Name",
		AuthorEmail: "person",
	}})
	assert.NoError(t, err)
	assert.Equal(t, engineer.Email, "person")

	engineer, err = chooseEngineer(dataClient, []*types.Commit{{
		AuthorName:  "Name",
		AuthorEmail: "robot-1",
	}})
	assert.NoError(t, err)
	assert.Nil(t, engineer)

	engineer, err = chooseEngineer(dataClient, []*types.Commit{{
		AuthorName:  "Name",
		AuthorEmail: "robot-2",
	}, {
		AuthorName:  "Other Name",
		AuthorEmail: "robot-1",
	}})
	assert.NoError(t, err)
	assert.Nil(t, engineer)

	engineer, err = chooseEngineer(dataClient, []*types.Commit{{
		AuthorName:  "Name",
		AuthorEmail: "robot-2",
	}, {
		AuthorName:  "Another Name",
		AuthorEmail: "person",
	}, {
		AuthorName:  "Other Name",
		AuthorEmail: "robot-1",
	}})
	assert.NoError(t, err)
	assert.Equal(t, engineer.Email, "person")
}

func TestCacheBackedLatestTrain(t *testing.T) {
	_, testData := setup(t)
	dataClient := data.NewClient()

	// Cache time is 0 - shouldn't get from cache, should update it.
	latestTrainCache = &types.Train{ID: 5000}
	latestTrainCacheUnixTime = 0
	train, err := getCacheBackedLatestTrain(dataClient, true)
	assert.NoError(t, err)
	assert.Equal(t, testData.Train.ID, train.ID)
	assert.Equal(t, testData.Train.ID, latestTrainCache.ID)
	assert.True(t, latestTrainCacheUnixTime > 0)

	// Cache time is recent - should get from cache, shouldn't update it.
	latestTrainCache = &types.Train{ID: 5000}
	cacheTime := time.Now().Unix() - 1
	latestTrainCacheUnixTime = cacheTime
	train, err = getCacheBackedLatestTrain(dataClient, true)
	assert.NoError(t, err)
	assert.Equal(t, 5000, int(train.ID))
	assert.Equal(t, 5000, int(latestTrainCache.ID))
	assert.True(t, latestTrainCacheUnixTime == cacheTime)

	// Cache time is too old - shouldn't get from cache, should update it.
	latestTrainCache = &types.Train{ID: 5000}
	cacheTime = time.Now().Unix() - TrainCacheTtl - 1
	latestTrainCacheUnixTime = cacheTime
	train, err = getCacheBackedLatestTrain(dataClient, true)
	assert.NoError(t, err)
	assert.Equal(t, testData.Train.ID, train.ID)
	assert.Equal(t, testData.Train.ID, latestTrainCache.ID)
	assert.True(t, latestTrainCacheUnixTime > cacheTime)

	// Cache time is good, but readFromCache parameter is false - shouldn't get from cache, should update it.
	latestTrainCache = &types.Train{ID: 5000}
	cacheTime = time.Now().Unix() - 1
	latestTrainCacheUnixTime = cacheTime
	train, err = getCacheBackedLatestTrain(dataClient, false)
	assert.NoError(t, err)
	assert.Equal(t, testData.Train.ID, train.ID)
	assert.Equal(t, testData.Train.ID, latestTrainCache.ID)
	assert.True(t, latestTrainCacheUnixTime > cacheTime)
}
