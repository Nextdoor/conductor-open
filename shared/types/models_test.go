package types

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	sha1 = "models_test_sha_1"
	sha2 = "models_test_sha_2"
	sha3 = "models_test_sha_3"
)

func TestNewCommitsNeedingTickets(t *testing.T) {
	commit1 := []*Commit{{SHA: sha1}}
	commit2 := []*Commit{{SHA: sha2}}
	bothCommits := append(commit1, commit2...)
	commit3 := []*Commit{{SHA: sha3, Message: "[no-verify] this change"}}
	allCommits := append(bothCommits, commit3...)

	train := &Train{
		Commits: allCommits,
		Tickets: []*Ticket{},
	}

	newCommits := train.NewCommitsNeedingTickets(sha1)
	assert.Equal(t, commit1, newCommits)

	newCommits = train.NewCommitsNeedingTickets(sha2)
	assert.Equal(t, bothCommits, newCommits)

	newCommits = train.NewCommitsNeedingTickets(sha3)
	assert.Equal(t, allCommits, newCommits)

	train = &Train{
		Commits: allCommits,
		Tickets: []*Ticket{
			{Commits: commit1},
		},
	}
	newCommits = train.NewCommitsNeedingTickets(sha2)
	assert.Equal(t, commit2, newCommits)

	train = &Train{
		Commits: allCommits,
		Tickets: []*Ticket{
			{Commits: commit2},
		},
	}
	newCommits = train.NewCommitsNeedingTickets(sha2)
	assert.Equal(t, commit1, newCommits)

	train = &Train{
		Commits: allCommits,
		Tickets: []*Ticket{
			{Commits: allCommits},
		},
	}
	newCommits = train.NewCommitsNeedingTickets(sha2)
	assert.Equal(t, []*Commit{}, newCommits)
}

func TestNotDeployableReason(t *testing.T) {
	var reason *string
	train := &Train{}

	reason = train.GetNotDeployableReason()
	assert.Nil(t, reason)

	train.ActivePhase = Verification
	train.ActivePhases = &PhaseGroup{}
	train.ActivePhases.Verification = &Phase{}

	var nextID uint64 = 1
	train.NextID = &nextID

	reason = train.GetNotDeployableReason()
	assert.Equal(t, "Not the latest train.", *reason)

	train.PreviousTrainDone = false
	train.NextID = nil

	reason = train.GetNotDeployableReason()
	assert.Equal(t, "Waiting for verification.", *reason)

	train.ActivePhases.Verification.CompletedAt = Time{time.Now()}

	reason = train.GetNotDeployableReason()
	assert.Equal(t, "Train is not closed.", *reason)

	train.Closed = true

	reason = train.GetNotDeployableReason()
	assert.Equal(t, "Previous train is still deploying.", *reason)

	train.PreviousTrainDone = true

	reason = train.GetNotDeployableReason()
	assert.Nil(t, reason)

	train.Blocked = true

	reason = train.GetNotDeployableReason()
	assert.Equal(t, "Train is blocked.", *reason)

	blockedReason := "test reason"
	train.BlockedReason = &blockedReason

	reason = train.GetNotDeployableReason()
	assert.Equal(t,
		fmt.Sprintf("Train is blocked due to %s.", blockedReason),
		*reason)
}
