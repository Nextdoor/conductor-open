package ticket

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	jira "github.com/niallo/go-jira"

	"github.com/Nextdoor/conductor/shared/datadog"
	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/settings"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	jiraURL             = flags.EnvString("JIRA_URL", "")
	jiraUsername        = flags.EnvString("JIRA_USERNAME", "")
	jiraPassword        = flags.EnvString("JIRA_PASSWORD", "")
	jiraProject         = flags.EnvString("JIRA_PROJECT", "")
	jiraParentIssueType = flags.EnvString("JIRA_PARENT_ISSUE_TYPE", "")
	jiraIssueType       = flags.EnvString("JIRA_ISSUE_TYPE", "")

	jiraClient *jira.Client

	ErrIssueNotFound = errors.New("Could not find parent issue for train")
)

type JIRA struct{}

const (
	parentIssueQueryJQL = "project = %s and summary ~ '%s' ORDER BY createdDate DESC"
	ticketsQueryJQL     = "project = %s and parent in (%s) and issuetype = '%s'"
	doneTransition      = "Done"
)

func newJIRA() *JIRA {
	if jiraURL == "" {
		panic(errors.New("jira_url flag must be set."))
	}
	if jiraUsername == "" {
		panic(errors.New("jira_username flag must be set."))
	}
	if jiraPassword == "" {
		panic(errors.New("jira_password flag must be set."))
	}
	if jiraProject == "" {
		panic(errors.New("jira_project flag must be set."))
	}
	if jiraParentIssueType == "" {
		panic(errors.New("jira_parent_issue_type flag must be set."))
	}
	if jiraIssueType == "" {
		panic(errors.New("jira_issue_type flag must be set."))
	}

	var err error
	jiraClient, err = jira.NewClient(nil, jiraURL)
	if err != nil {
		panic(err)
	}

	_, err = jiraClient.Authentication.AcquireSessionCookie(jiraUsername, jiraPassword)
	if err != nil {
		panic(err)
	}

	DefaultTicketUsername = jiraUsername
	return &JIRA{}
}

func parentIssueQuery(jiraProject, summary string) string {
	return fmt.Sprintf(parentIssueQueryJQL, jiraProject, summary)
}

func ticketsQuery(jiraProject, parentIssueKey, jiraIssueType string) string {
	return fmt.Sprintf(ticketsQueryJQL, jiraProject, parentIssueKey, jiraIssueType)
}

func (t *JIRA) CreateTickets(train *types.Train, commits []*types.Commit) ([]*types.Ticket, error) {
	tickets, err := ticketsFromCommits(train, commits)
	if err != nil {
		return nil, err
	}
	return tickets, nil
}

func (t *JIRA) CloseTickets(tickets []*types.Ticket) error {
	keys := make([]string, len(tickets))
	for i := range tickets {
		keys[i] = tickets[i].Key
	}
	err := t.closeIssuesByKeys(keys)
	return err
}

func (t *JIRA) DeleteTickets(train *types.Train) error {
	parentIssue, err := getParentIssue(train)
	if err != nil {
		return err
	}
	resp, err := jiraClient.Issue.Delete(parentIssue.Key, true)
	if err != nil {
		return parseBodyError(resp, err)
	}
	return nil
}

func (t *JIRA) SyncTickets(train *types.Train) ([]*types.Ticket, []*types.Ticket, error) {
	parentIssue, err := getParentIssue(train)
	if err != nil {
		return nil, nil, err
	}
	jql := ticketsQuery(jiraProject, parentIssue.Key, jiraIssueType)
	issues, resp, err := jiraClient.Issue.Search(jql, nil)
	if err != nil {
		return nil, nil, parseBodyError(resp, err)
	}

	// Tickets on the train that are not found in JIRA have been destroyed.
	// Tickets in JIRA that are not on the train are new.
	// Tickets in both are checked for updates.

	keyToTicket := make(map[string]*types.Ticket)
	for i := range train.Tickets {
		ticket := train.Tickets[i]
		keyToTicket[ticket.Key] = ticket
	}
	keyToIssue := make(map[string]*jira.Issue)
	for i := range issues {
		issue := issues[i]
		keyToIssue[issue.Key] = &issue
	}

	newTickets := make([]*types.Ticket, 0)
	updatedTickets := make([]*types.Ticket, 0)

	// Check JIRA state for new tickets.
	for key, issue := range keyToIssue {
		if _, found := keyToTicket[key]; !found {
			// Unknown ticket found in JIRA.
			email, name := getUserForIssue(issue)
			ticket := createTicket(train, issue.Key, issue.Fields.Summary,
				email, name, nil)
			newTickets = append(newTickets, ticket)
		}
	}

	// Check all known tickets against JIRA state.
	for key, ticket := range keyToTicket {
		if issue, found := keyToIssue[key]; found {
			// Known ticket was found in JIRA.

			// Update ticket if any fields have changed.
			updated := false

			issueDone := issue.Fields.Status.Name == doneTransition
			ticketDone := ticket.ClosedAt.HasValue()
			if issueDone != ticketDone {
				if issueDone {
					// Newly closed
					ticket.ClosedAt = types.Time{time.Now()}
				} else {
					// Newly open/Re-open
					ticket.ClosedAt = types.Time{}
				}
				updated = true
			}

			if issue.Fields.Summary != ticket.Summary {
				ticket.Summary = issue.Fields.Summary
				updated = true
			}

			email, name := getUserForIssue(issue)

			if email != ticket.AssigneeEmail {
				ticket.AssigneeEmail = email
				updated = true
			}

			if name != ticket.AssigneeName {
				ticket.AssigneeName = name
				updated = true
			}

			if updated {
				updatedTickets = append(updatedTickets, ticket)
			}
		} else {
			// Known ticket was not found in JIRA.
			// Deleted
			if !ticket.DeletedAt.HasValue() {
				ticket.DeletedAt = types.Time{time.Now()}
				updatedTickets = append(updatedTickets, ticket)
			}
		}
	}

	return newTickets, updatedTickets, nil
}

func (t *JIRA) CloseTrainTickets(train *types.Train) error {
	// Close all the train's issues: children and the parent.
	parentIssue, err := getParentIssue(train)
	if err != nil {
		return err
	}
	jql := ticketsQuery(jiraProject, parentIssue.Key, jiraIssueType)
	issues, resp, err := jiraClient.Issue.Search(jql, nil)
	if err != nil {
		return parseBodyError(resp, err)
	}

	keys := make([]string, len(issues)+1)
	// Add children issues.
	for i := range issues {
		keys[i] = issues[i].Key
	}
	// Add parent issue.
	keys[len(issues)] = parentIssue.Key

	// TODO: Need system that properly moves tickets that were closed by Conductor to a new train
	// in the case of branch switching.
	err = t.closeIssuesByKeys(keys)
	if err != nil {
		return err
	}

	return nil
}

func (t *JIRA) closeIssuesByKeys(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// Just fetch the transitions for the first issue. The API request is
	// slow and they should change exceedingly rarely.
	transitions, resp, err := jiraClient.Issue.GetTransitions(keys[0])
	if err != nil {
		return parseBodyError(resp, err)
	}

	doneTransitionID := ""
	for _, transition := range transitions {
		if transition.Name == doneTransition {
			doneTransitionID = transition.ID
			break
		}
	}
	if len(doneTransitionID) == 0 {
		return fmt.Errorf("Could not find JIRA transition ID for transition named %s",
			doneTransition)
	}
	closed := make([]string, 0)
	for i := range keys {
		resp, err := jiraClient.Issue.DoTransition(keys[i], doneTransitionID)
		if err != nil {
			return parseBodyError(resp, err)
		}
		closed = append(closed, fmt.Sprintf("%v", keys[i]))
	}
	datadog.Info("Closed issues by keys: %v", strings.Join(closed, "\n"))
	return nil
}

func createTicket(train *types.Train, key, summary, assigneeEmail, assigneeName string, commits []*types.Commit) *types.Ticket {
	datadog.Info("Created ticket (Key, Summary, AssigneeName) %v, %v, %v", key, summary, assigneeName)
	return &types.Ticket{
		Key:           key,
		Summary:       summary,
		AssigneeEmail: assigneeEmail,
		AssigneeName:  assigneeName,
		Commits:       commits,
		Train:         train,
		URL:           fmt.Sprintf("%s/browse/%s", jiraURL, key),
	}
}

func parentSummary(train *types.Train) string {
	return fmt.Sprintf("Train %d", train.ID)
}

func createParentIssue(train *types.Train) (*jira.Issue, error) {
	// Create parent issue.
	// Individual tickets are linked to it as a sub-task.
	i := jira.Issue{
		Fields: &jira.IssueFields{
			Assignee: &jira.User{
				Name: jiraUsername,
			},
			Reporter: &jira.User{
				Name: jiraUsername,
			},
			Type: jira.IssueType{
				Name: jiraParentIssueType,
			},
			Project: jira.Project{
				Key: jiraProject,
			},
			Summary: parentSummary(train),
			// TODO: Description should probably have a link
			// back to this train in conductor.
		},
	}
	// Create just returns a minimal issue struct
	parentIssue, resp, err := jiraClient.Issue.Create(&i)
	if err != nil {
		return nil, parseBodyError(resp, err)
	}
	// Call Get on the minimal issue to get the fully-populated struct
	parentIssue, resp, err = jiraClient.Issue.Get(parentIssue.ID, nil)
	if err != nil {
		return nil, parseBodyError(resp, err)
	}
	datadog.Info("Created parent issue %v", parentIssue.ID)
	return parentIssue, nil
}

func getParentIssue(train *types.Train) (*jira.Issue, error) {
	jql := parentIssueQuery(jiraProject, parentSummary(train))
	issues, resp, err := jiraClient.Issue.Search(jql, nil)
	if err != nil {
		return nil, parseBodyError(resp, err)
	}
	if len(issues) == 0 {
		return nil, ErrIssueNotFound
	}
	if len(issues) > 1 {
		logger.Error("Danger: More than one parent issue for train ID %d in project %s",
			train.ID, jiraProject)
	}
	return &issues[0], nil
}

func createSubIssue(parentIssue *jira.Issue, username string, commits []*types.Commit) (*jira.Issue, error) {
	desc, err := descriptionFromCommits(commits)
	if err != nil {
		logger.Error("Error generating descriptionFromCommits: %v", err)
		return nil, err
	}
	issue := &jira.Issue{
		Fields: &jira.IssueFields{
			Assignee: &jira.User{
				Name: username,
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
			Summary:     summaryForCommit(commits[0]),
			Description: desc,
			Parent: &jira.Parent{
				ID: parentIssue.ID,
			},
		},
	}
	// Create just returns a minimal issue struct.
	issue, resp, err := jiraClient.Issue.Create(issue)
	if err != nil {
		return nil, parseBodyError(resp, err)
	}
	// We need to make another request to get the full issue object.
	issue, resp, err = jiraClient.Issue.Get(issue.ID, nil)
	if err != nil {
		return nil, parseBodyError(resp, err)
	}
	datadog.Info("Created sub issue %v", issue.ID)
	return issue, nil
}

func ticketsFromCommits(train *types.Train, commits []*types.Commit) ([]*types.Ticket, error) {
	if len(commits) == 0 {
		return nil, fmt.Errorf("No commits passed to ticketsFromCommits")
	}

	parentIssue, err := getParentIssue(train)
	if err != nil && err != ErrIssueNotFound {
		return nil, err
	}
	if parentIssue == nil {
		parentIssue, err = createParentIssue(train)
		if err != nil {
			return nil, err
		}
	}

	commitsMap := commitsByEmail(commits)

	i := 0
	tickets := make([]*types.Ticket, len(commitsMap))
	for email, commits := range commitsMap {
		username := emailToUsernameInJIRA(email)
		subissue, err := createSubIssue(parentIssue, username, commits)
		if err != nil {
			return nil, err
		}
		email, name := getUserForIssue(subissue)
		ticket := createTicket(train, subissue.Key, subissue.Fields.Summary,
			email, name, commits)
		tickets[i] = ticket
		i += 1
	}
	return tickets, nil
}

func commitsByEmail(commits []*types.Commit) map[string][]*types.Commit {
	commitsByEmail := make(map[string][]*types.Commit, 0)
	for _, commit := range commits {
		email := commit.AuthorEmail
		if settings.IsRobotUser(email) {
			continue
		}
		commitsByEmail[email] = append(
			commitsByEmail[email], commit)
	}

	return commitsByEmail
}

// If not found, returns DefaultTicketUsername.
func emailToUsernameInJIRA(email string) string {
	users, resp, err := jiraClient.User.FindUsers(email, nil)
	if err != nil {
		err = parseBodyError(resp, err)
		logger.Error("Error finding JIRA user for email %s: %v: %s", email, err)
		return DefaultTicketUsername
	}

	if len(users) == 0 || users[0].EmailAddress != email {
		logger.Error("Could not find JIRA user for email %s", email)
		return DefaultTicketUsername
	}

	if len(users) > 1 {
		logger.Error("Warning: Found multiple JIRA users for email %s: %v", email, users)
	}

	return users[0].Name
}

func getUserForIssue(issue *jira.Issue) (string, string) {
	// Fall back to reporter if there is no assignee.
	// Every issue will have a reporter.
	var email string
	var name string
	if issue.Fields.Assignee != nil {
		email = issue.Fields.Assignee.EmailAddress
		name = issue.Fields.Assignee.DisplayName
	} else {
		email = issue.Fields.Reporter.EmailAddress
		name = issue.Fields.Reporter.DisplayName
	}
	return email, name
}

func parseBodyError(resp *jira.Response, err error) error {
	if resp != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("%v: %s", err, string(body))
	}
	return err
}
