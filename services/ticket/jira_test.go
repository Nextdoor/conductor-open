// +build ticket

package ticket

import (
	"testing"

	jira "github.com/niallo/go-jira"
	"github.com/stretchr/testify/assert"

	"github.com/Nextdoor/conductor/shared/types"
)

func TestJIRATickets(t *testing.T) {
	jiraService := newJIRA()

	testCommits := []*types.Commit{
		{AuthorEmail: email1, Message: message1, AuthorName: jiraUsername, SHA: sha1},
		{AuthorEmail: email1, Message: message2, AuthorName: jiraUsername, SHA: sha2}}

	train := &types.Train{
		ID:     1,
		Branch: "branch",
	}

	// Test that closed is detected.
	newTickets, err := jiraService.CreateTickets(train, testCommits)
	assert.NoError(t, err)
	assert.Len(t, newTickets, 1)
	assert.False(t, newTickets[0].ClosedAt.HasValue())

	err = jiraService.CloseTickets(newTickets)
	assert.NoError(t, err)

	train.Tickets = newTickets

	newTickets, updatedTickets, err := jiraService.SyncTickets(train)
	assert.NoError(t, err)
	assert.Len(t, newTickets, 0)
	assert.Len(t, updatedTickets, 1)
	assert.True(t, updatedTickets[0].ClosedAt.HasValue())

	parentIssue, err := getParentIssue(train)
	assert.NoError(t, err)

	newCommits := []*types.Commit{
		{AuthorEmail: email2, Message: message3, AuthorName: jiraUsername, SHA: sha3}}
	newTickets, err = jiraService.CreateTickets(train, newCommits)
	assert.NoError(t, err)
	assert.Len(t, newTickets, 1)

	train.Tickets = append(train.Tickets, newTickets[0])

	// Test that the new tickets are appeneded to the same parent issue.
	newParentIssue, err := getParentIssue(train)
	assert.NoError(t, err)
	assert.Equal(t, parentIssue.Key, newParentIssue.Key)

	// Create new non-commit issue.
	summary := "Non-commit issue"
	issue := &jira.Issue{
		Fields: &jira.IssueFields{
			Assignee: &jira.User{
				Name: jiraUsername,
			},
			Reporter: &jira.User{
				Name: jiraUsername,
			},
			Type: jira.IssueType{
				Name: jiraIssueType,
			},
			Project: jira.Project{
				Key: jiraProject,
			},
			Summary: summary,
			Parent: &jira.Parent{
				ID: parentIssue.ID,
			},
		},
	}
	// Create just returns a minimal issue struct.
	issue, _, err = jiraClient.Issue.Create(issue)
	assert.NoError(t, err)

	// We need to make another request to get the full issue object.
	issue, _, err = jiraClient.Issue.Get(issue.ID, nil)
	assert.NoError(t, err)

	newTickets, updatedTickets, err = jiraService.SyncTickets(train)
	assert.NoError(t, err)
	assert.Len(t, newTickets, 1)
	assert.Len(t, updatedTickets, 0)

	train.Tickets = append(train.Tickets, newTickets[0])

	// Test that deletion is detected.
	_, err = jiraClient.Issue.Delete(train.Tickets[0].Key, false)
	assert.NoError(t, err)
	_, err = jiraClient.Issue.Delete(train.Tickets[1].Key, false)
	assert.NoError(t, err)

	newTickets, updatedTickets, err = jiraService.SyncTickets(train)
	assert.NoError(t, err)
	assert.Len(t, newTickets, 0)
	assert.Len(t, updatedTickets, 2)
	assert.True(t, updatedTickets[0].DeletedAt.HasValue())
	assert.True(t, updatedTickets[1].DeletedAt.HasValue())
	assert.True(t, updatedTickets[0].Key == train.Tickets[0].Key ||
		updatedTickets[0].Key == train.Tickets[1].Key)
	assert.True(t, updatedTickets[1].Key == train.Tickets[0].Key ||
		updatedTickets[1].Key == train.Tickets[1].Key)
	assert.True(t, updatedTickets[0].Key != updatedTickets[1].Key)

	// Clean up
	err = jiraService.DeleteTickets(train)
	assert.NoError(t, err)

	_, err = getParentIssue(train)
	assert.Equal(t, ErrIssueNotFound, err)
}

func TestCloseTrainTickets(t *testing.T) {
	jiraService := newJIRA()

	testCommits := []*types.Commit{
		{AuthorEmail: email1, Message: message1, AuthorName: jiraUsername, SHA: sha1},
		{AuthorEmail: email1, Message: message2, AuthorName: jiraUsername, SHA: sha2}}

	train := &types.Train{
		ID:     1,
		Branch: "branch",
	}

	newTickets, err := jiraService.CreateTickets(train, testCommits)
	assert.NoError(t, err)
	assert.Len(t, newTickets, 1)
	assert.False(t, newTickets[0].ClosedAt.HasValue())

	train.Tickets = newTickets

	err = jiraService.CloseTrainTickets(train)
	assert.NoError(t, err)

	// Test that children issues are closed by CloseTrainTickets.
	assert.Len(t, newTickets, 1)
	assert.False(t, newTickets[0].ClosedAt.HasValue())

	newTickets, updatedTickets, err := jiraService.SyncTickets(train)
	assert.NoError(t, err)
	assert.Len(t, newTickets, 0)
	assert.Len(t, updatedTickets, 1)
	assert.True(t, updatedTickets[0].ClosedAt.HasValue())

	// Test that parent issue is closed by CloseTrainTickets.
	parentIssue, err := getParentIssue(train)
	assert.NoError(t, err)

	assert.Equal(t, doneTransition, parentIssue.Fields.Status.Name)

	// Clean up
	err = jiraService.DeleteTickets(train)
	assert.NoError(t, err)

	_, err = getParentIssue(train)
	assert.Equal(t, ErrIssueNotFound, err)
}
