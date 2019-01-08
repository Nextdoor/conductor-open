// +build data

// To run data integration tests, make sure you have a dev postgres container running
// from `make postgres`. You probably want to set `POSTGRES_HOST=localhost`.

package data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Nextdoor/conductor/shared/types"
)

const (
	branch               = "foobar"
	branch2              = "foobar2"
	sha1                 = "methods_test_sha_1"
	sha2                 = "methods_test_sha_2"
	sha3                 = "methods_test_sha_3"
	userName             = "test"
	userName2            = "test2"
	email                = "test@example.com"
	email2               = "test2@example.com"
	ticketKey1           = "REL-1"
	ticketKey2           = "REL-2"
	ticketKey3           = "REL-3"
	ticketKey4           = "REL-4"
	ticketKey5           = "REL-5"
	ticketKey6           = "REL-6"
	ticketAssigneeEmail1 = "test-assignee1@example.com"
	ticketAssigneeEmail2 = "test-assignee2@example.com"
	ticketAssigneeName1  = "assignee 1"
	ticketAssigneeName2  = "assignee 2"
)

func TestDataCreateTrain(t *testing.T) {
	data := NewClient()

	user, err := data.ReadOrCreateUser("test-user", "test-user@email.com")
	assert.NoError(t, err)

	commits := []*types.Commit{{SHA: sha1}}
	train, err := data.CreateTrain(branch, user, commits)
	assert.NoError(t, err)
	assert.Equal(t, train.Branch, branch)
	assert.Equal(t, train.HeadSHA, sha1)
	assert.Equal(t, train.TailSHA, sha1)
	assert.Equal(t, train.Engineer.ID, user.ID)

	// Ensure tail->head order.
	commits = append(commits, &types.Commit{SHA: sha2})
	train, err = data.CreateTrain(branch, nil, commits)
	assert.NoError(t, err)
	assert.Equal(t, train.TailSHA, sha1)
	assert.Equal(t, train.HeadSHA, sha2)

	// Create a new train, and verify that is the new latest train.
	latestTrain, err := data.LatestTrain()
	assert.NoError(t, err)
	assert.Equal(t, train.HeadSHA, latestTrain.HeadSHA)

	train2, err := data.CreateTrain(branch2, nil, commits)
	latestTrain, err = data.LatestTrain()
	assert.NoError(t, err)
	assert.NotEqual(t, train.Branch, latestTrain.Branch)
	assert.Equal(t, train2.Branch, latestTrain.Branch)
	train2Len := len(train2.Commits)

	// Extend the train and verify it updated the DB correctly.
	// Ensure tail->head order.
	newCommits := []*types.Commit{{SHA: sha3}}

	err = data.ExtendTrain(train2, train2.Engineer, newCommits)
	assert.NoError(t, err)
	assert.Equal(t, train2.HeadSHA, sha3)
	assert.Equal(t, train2.TailSHA, sha1)
	// Should have one more commit than the train it extended.
	assert.Len(t, train2.Commits, train2Len+1)
}

func TestLoadTrainRelated(t *testing.T) {
	data := NewClient()

	commits := []*types.Commit{{SHA: sha1}, {SHA: sha2}}
	train, err := data.CreateTrain(branch, nil, commits)
	assert.NoError(t, err)

	tickets := []*types.Ticket{
		{Key: ticketKey5, AssigneeEmail: ticketAssigneeEmail1, AssigneeName: ticketAssigneeName1, Train: train, Commits: commits},
		{Key: ticketKey6, AssigneeEmail: ticketAssigneeEmail2, AssigneeName: ticketAssigneeName2, Train: train, Commits: commits}}

	err = data.WriteTickets(tickets)
	assert.NoError(t, err)
	train.Tickets = tickets

	assert.Len(t, train.Tickets, 2)
	assert.Len(t, train.Tickets[0].Commits, 2)

	d := data.(*dataClient)

	err = d.loadTrainRelated(train)
	assert.NoError(t, err)
	assert.Len(t, train.Tickets, 2)
	assert.Len(t, train.Tickets[0].Commits, 2)
}

func TestTrainPreviousID(t *testing.T) {
	data := NewClient()

	train, err := data.CreateTrain(branch, nil, []*types.Commit{{SHA: sha1}})
	assert.NoError(t, err)

	d := data.(*dataClient)

	err = d.loadTrainRelated(train)
	assert.NoError(t, err)

	assert.Equal(t, train.ID-1, *train.PreviousID)

	firstTrain := &types.Train{}
	err = d.Client.QueryTable(firstTrain).OrderBy("id").One(firstTrain)
	assert.NoError(t, err)

	err = d.loadTrainRelated(firstTrain)
	assert.NoError(t, err)

	assert.Nil(t, firstTrain.PreviousID)
}

func TestTrainNextID(t *testing.T) {
	data := NewClient()

	d := data.(*dataClient)

	firstTrain := &types.Train{}
	err := d.Client.QueryTable(firstTrain).OrderBy("id").One(firstTrain)
	assert.NoError(t, err)

	err = d.loadTrainRelated(firstTrain)
	assert.NoError(t, err)

	assert.Equal(t, firstTrain.ID+1, *firstTrain.NextID)

	train, err := data.CreateTrain(branch, nil, []*types.Commit{{SHA: sha1}})
	assert.NoError(t, err)

	err = d.loadTrainRelated(train)
	assert.NoError(t, err)

	assert.Nil(t, train.NextID)
}

func TestTrainDoneAfterDeployment(t *testing.T) {
	data := NewClient()

	train, err := data.CreateTrain(branch, nil, []*types.Commit{{SHA: sha1}})
	assert.NoError(t, err)

	d := data.(*dataClient)

	err = d.loadTrainRelated(train)
	assert.NoError(t, err)

	assert.Equal(t, false, train.Done)

	err = d.DeployTrain(train)
	assert.NoError(t, err)

	err = d.loadTrainRelated(train)
	assert.NoError(t, err)

	assert.Equal(t, true, train.Done)
}

func TestTrainDoneWhenCancelled(t *testing.T) {
	data := NewClient()

	train, err := data.CreateTrain(branch, nil, []*types.Commit{{SHA: sha1}})
	assert.NoError(t, err)

	d := data.(*dataClient)

	err = d.loadTrainRelated(train)
	assert.NoError(t, err)

	assert.Equal(t, false, train.Done)

	err = data.CancelTrain(train)
	assert.NoError(t, err)

	err = d.loadTrainRelated(train)
	assert.NoError(t, err)

	assert.Equal(t, true, train.Done)
}

func TestTrainNotDoneAfterAnotherTrainIfDeploying(t *testing.T) {
	data := NewClient()

	train, err := data.CreateTrain(branch, nil, []*types.Commit{{SHA: sha1}})
	assert.NoError(t, err)
	data.StartPhase(train.ActivePhases.Deploy)

	d := data.(*dataClient)

	_, err = data.CreateTrain(branch, nil, []*types.Commit{{SHA: sha1}})
	assert.NoError(t, err)

	err = d.loadTrainRelated(train)
	assert.NoError(t, err)

	assert.Equal(t, false, train.Done)
}

func TestTrainActivePhase(t *testing.T) {
	data := NewClient()

	train, err := data.CreateTrain(branch, nil, []*types.Commit{{SHA: sha1}})
	assert.NoError(t, err)

	d := data.(*dataClient)

	err = d.loadTrainRelated(train)
	assert.NoError(t, err)

	assert.Equal(t, types.Delivery, train.ActivePhase)

	err = data.StartPhase(train.ActivePhases.Verification)
	assert.NoError(t, err)

	err = d.loadTrainRelated(train)
	assert.NoError(t, err)

	assert.Equal(t, types.Verification, train.ActivePhase)

	err = data.StartPhase(train.ActivePhases.Deploy)
	assert.NoError(t, err)

	err = d.loadTrainRelated(train)
	assert.NoError(t, err)

	assert.Equal(t, types.Deploy, train.ActivePhase)
}

func TestReplacePhase(t *testing.T) {
	data := NewClient()

	commits := []*types.Commit{{SHA: sha1}, {SHA: sha2}}
	train, err := data.CreateTrain(branch, nil, commits)
	assert.NoError(t, err)

	oldDeliveryPhase := train.ActivePhases.Delivery
	oldVerificationPhase := train.ActivePhases.Verification
	oldDeployPhase := train.ActivePhases.Deploy

	newDeliveryPhase, err := data.ReplacePhase(oldDeliveryPhase)
	assert.NoError(t, err)

	newVerificationPhase, err := data.ReplacePhase(oldVerificationPhase)
	assert.NoError(t, err)

	newDeployPhase, err := data.ReplacePhase(oldDeployPhase)
	assert.NoError(t, err)

	assert.NotEqual(t, newDeliveryPhase, oldDeliveryPhase)
	assert.NotEqual(t, newVerificationPhase, oldVerificationPhase)
	assert.NotEqual(t, newDeployPhase, oldDeployPhase)

	assert.Equal(t, newDeliveryPhase, train.ActivePhases.Delivery)
	assert.Equal(t, newVerificationPhase, train.ActivePhases.Verification)
	assert.Equal(t, newDeployPhase, train.ActivePhases.Deploy)

	assert.NotEqual(t, 0, newDeliveryPhase.ID)
	assert.NotEqual(t, 0, newVerificationPhase.ID)
	assert.NotEqual(t, 0, newDeployPhase.ID)
}

func TestDataCreateTickets(t *testing.T) {
	data := NewClient()

	commits := []*types.Commit{{SHA: sha1}, {SHA: sha2}}
	train, err := data.CreateTrain(branch, nil, commits)
	assert.NoError(t, err)

	tickets := []*types.Ticket{
		{Key: ticketKey1, AssigneeEmail: ticketAssigneeEmail1, AssigneeName: ticketAssigneeName1, Train: train, Commits: commits},
		{Key: ticketKey2, AssigneeEmail: ticketAssigneeEmail2, AssigneeName: ticketAssigneeName2, Train: train, Commits: commits}}

	err = data.WriteTickets(tickets)
	assert.NoError(t, err)

	assert.Len(t, tickets, 2)
	assert.True(t, tickets[0].CreatedAt.HasValue())
	assert.False(t, tickets[0].ClosedAt.HasValue())
	assert.False(t, tickets[0].DeletedAt.HasValue())
	assert.Len(t, tickets[0].Commits, 2)
}

func TestDataUpdateTickets(t *testing.T) {
	data := NewClient()

	commits := []*types.Commit{{SHA: sha1}, {SHA: sha2}}
	train, err := data.CreateTrain(branch, nil, commits)
	assert.NoError(t, err)

	tickets := []*types.Ticket{
		{Key: ticketKey3, AssigneeEmail: ticketAssigneeEmail1, AssigneeName: ticketAssigneeName1, Train: train, Commits: commits},
		{Key: ticketKey4, AssigneeEmail: ticketAssigneeEmail2, AssigneeName: ticketAssigneeName2, Train: train, Commits: commits}}

	err = data.WriteTickets(tickets)
	assert.NoError(t, err)

	assert.Len(t, tickets, 2)
	assert.True(t, tickets[0].CreatedAt.HasValue())
	assert.False(t, tickets[0].ClosedAt.HasValue())
	assert.False(t, tickets[0].DeletedAt.HasValue())
	assert.Len(t, tickets[0].Commits, 2)

	// Set the ticket to closed
	tickets[0].ClosedAt = types.Time{time.Now()}

	err = data.UpdateTickets(tickets)
	assert.NoError(t, err)
	assert.Len(t, tickets, 2)
	assert.True(t, tickets[0].ClosedAt.HasValue())
}
