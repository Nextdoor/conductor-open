/* Handles messaging users. */
package messaging

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/settings"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	implementationFlag = flags.EnvString("MESSAGING_IMPL", "fake")
)

type Service interface {
	TrainCreation(*types.Train, []*types.Commit)
	TrainExtension(*types.Train, []*types.Commit, *types.User)
	TrainDuplication(*types.Train, *types.Train, []*types.Commit)
	TrainDelivered(*types.Train, []*types.Commit, []*types.Ticket)
	TrainVerified(*types.Train)
	TrainUnverified(*types.Train)
	TrainDeploying()
	TrainDeployed(*types.Train)
	TrainClosed(*types.Train, *types.User)
	TrainOpened(*types.Train, *types.User)
	TrainBlocked(*types.Train, *types.User)
	TrainUnblocked(*types.Train, *types.User)
	TrainCancelled(*types.Train, *types.User)
	EngineerChanged(*types.Train, *types.User)
	RollbackInitiated(*types.Train, *types.User)
	RollbackInfo(*types.User)
	JobFailed(*types.Job)
}

type Messenger struct {
	Engine Engine
}

type Engine interface {
	send(text string)
	sendDirect(name, email, text string)
	formatUser(*types.User) string
	formatNameEmail(name, email string) string
	formatNameEmailNotification(name, email string) string
	formatLink(url, text string) string
	formatBold(text string) string
	formatMonospaced(text string) string
	indent(text string) string
	escape(text string) string
}

type nameFormat int

const (
	None nameFormat = iota
	Notify
	PlainText
)

// On train creation, send a link to the train to the slack channel,
// and send direct messages to all committers on the train.
func (m Messenger) TrainCreation(train *types.Train, commits []*types.Commit) {
	m.Engine.send(m.Engine.formatBold(
		fmt.Sprintf("%s going to staging.", m.formatTrainLink(train, "New train"))))

	if train.Engineer != nil {
		m.Engine.send(m.Engine.formatBold(
			fmt.Sprintf("%s is the engineer.",
				m.Engine.formatUser(train.Engineer))))

		m.Engine.sendDirect(train.Engineer.Name, train.Engineer.Email, m.Engine.formatBold(
			fmt.Sprintf("You are the engineer for the %s.",
				m.formatTrainLink(train, "train"))))
	}

	commitSets := m.commitSetsFromCommits(commits, true)

	m.sendCommitSetsDirectly(
		fmt.Sprintf("Your changes are %s", m.formatTrainLink(train, "going to staging")),
		commitSets)
}

func (m Messenger) TrainExtension(train *types.Train, commits []*types.Commit, user *types.User) {
	commitSets := m.commitSetsFromCommits(commits, true)
	// Note: We used to abort early if there are no commit sets to notify for.
	// Even if no commit sets, send train extension message for manual extensions.
	// If all the changes are no-verify, we still want to notify the staging room.

	trainLink := m.formatTrainLink(train, "Train extended")
	if user != nil {
		// Only send this for manual extensions.
		// Noisy when train is opened.
		m.Engine.send(m.Engine.formatBold(
			fmt.Sprintf("%s, new changes going to staging.",
				fmt.Sprintf("%s by %s", trainLink, m.Engine.formatUser(user)))))
	}

	m.sendCommitSetsDirectly(
		fmt.Sprintf("Your changes are %s", m.formatTrainLink(train, "going to staging")),
		commitSets)
}

func (m Messenger) TrainDuplication(train *types.Train, trainFrom *types.Train, commits []*types.Commit) {
	// Same message as create for now.
	m.TrainCreation(train, commits)
}

func (m Messenger) TrainDelivered(train *types.Train, commits []*types.Commit, tickets []*types.Ticket) {
	ticketedCommitSets, unticketedCommitSets := m.commitSetsFromCommitsAndTickets(commits, tickets)

	if len(ticketedCommitSets) > 0 {
		m.Engine.send(m.Engine.formatBold(
			fmt.Sprintf("%s delivered to staging.", m.formatTrainLink(train, "Train"))))
		m.Engine.send(m.formatCommitSets("Changes with tickets", PlainText, ticketedCommitSets))
	}

	m.sendCommitSetsDirectly(
		fmt.Sprintf("Your [no-verify] changes have %s",
			m.formatTrainLink(train, "arrived on staging")),
		unticketedCommitSets)
	m.sendCommitSetsDirectly(
		fmt.Sprintf("Your changes have %s and need verification",
			m.formatTrainLink(train, "arrived on staging")),
		ticketedCommitSets)
}

func (m Messenger) TrainVerified(train *types.Train) {
	if !train.Closed {
		// No message if verified but opened, because it's not yet actionable.
		return
	}

	m.Engine.send(m.Engine.formatBold(
		fmt.Sprintf("%s fully verified.", m.formatTrainLink(train, "Train"))))
}

func (m Messenger) TrainUnverified(train *types.Train) {
	if !train.Closed {
		// No message if unverified but opened, because it's not yet actionable.
		return
	}

	message := m.Engine.formatBold(
		fmt.Sprintf("%s no longer fully verified.", m.formatTrainLink(train, "Train")))
	m.Engine.send(message)

	if train.Engineer != nil {
		m.Engine.sendDirect(train.Engineer.Name, train.Engineer.Email, message)
	}
}

func (m Messenger) TrainDeploying() {
	m.Engine.send(m.Engine.formatBold("Deploy started."))
}

func (m Messenger) TrainDeployed(train *types.Train) {
	commitSets := m.commitSetsFromCommits(train.Commits, false)

	m.Engine.send(m.Engine.formatBold(fmt.Sprintf(
		"Deployed %s to production.",
		m.formatTrainLink(train,
			fmt.Sprintf("Train %d", train.ID)))))

	m.sendCommitSetsDirectly(
		fmt.Sprintf("Your changes were %s",
			m.formatTrainLink(train, "deployed to production")),
		commitSets)
}

func (m Messenger) TrainClosed(train *types.Train, user *types.User) {
	var text string
	if user != nil {
		text = m.Engine.formatBold(
			fmt.Sprintf("%s closed by %s.",
				m.formatTrainLink(train, "Train"),
				m.Engine.formatUser(user)))
	} else {
		text = m.Engine.formatBold(
			fmt.Sprintf("%s closed.",
				m.formatTrainLink(train, "Train")))
	}

	m.Engine.send(text)
}

func (m Messenger) TrainOpened(train *types.Train, user *types.User) {
	var text string
	if user != nil {
		text = m.Engine.formatBold(
			fmt.Sprintf("%s opened by %s.",
				m.formatTrainLink(train, "Train"),
				m.Engine.formatUser(user)))
	} else {
		text = m.Engine.formatBold(
			fmt.Sprintf("%s opened.",
				m.formatTrainLink(train, "Train")))
	}

	m.Engine.send(text)
}

func (m Messenger) TrainBlocked(train *types.Train, user *types.User) {
	var text string
	if user != nil {
		text = m.Engine.formatBold(
			fmt.Sprintf("%s blocked by %s.",
				m.formatTrainLink(train, "Train"),
				m.Engine.formatUser(user)))
	} else {
		text = m.Engine.formatBold(
			fmt.Sprintf("%s blocked.",
				m.formatTrainLink(train, "Train")))
	}

	m.Engine.send(text)
}

func (m Messenger) TrainUnblocked(train *types.Train, user *types.User) {
	var text string
	if user != nil {
		text = m.Engine.formatBold(
			fmt.Sprintf("%s unblocked by %s.",
				m.formatTrainLink(train, "Train"),
				m.Engine.formatUser(user)))
	} else {
		text = m.Engine.formatBold(
			fmt.Sprintf("%s unblocked.",
				m.formatTrainLink(train, "Train")))
	}

	m.Engine.send(text)
}

func (m Messenger) TrainCancelled(train *types.Train, user *types.User) {
	var text string
	if user != nil {
		text = m.Engine.formatBold(
			fmt.Sprintf("%s cancelled by %s.",
				m.formatTrainLink(train, "Train"),
				m.Engine.formatUser(user)))
	} else {
		text = m.Engine.formatBold(
			fmt.Sprintf("%s cancelled.",
				m.formatTrainLink(train, "Train")))
	}

	m.Engine.send(text)
}

func (m Messenger) EngineerChanged(train *types.Train, user *types.User) {
	var text = m.Engine.formatBold(
		fmt.Sprintf("Train %s is claimed by new engineer %s.",
			m.formatTrainLink(train, "Train"),
			m.Engine.formatUser(user)))

	m.Engine.send(text)
}

func (m Messenger) RollbackInitiated(train *types.Train, user *types.User) {
	var text string
	if user != nil {
		text = m.Engine.formatBold(
			fmt.Sprintf("Rollback to %s %d initiated by %s.",
				m.formatTrainLink(train, "Train"),
				train.ID,
				m.Engine.formatUser(user)))
	} else {
		text = m.Engine.formatBold(
			fmt.Sprintf("Rollback to %s %d initiated.",
				m.formatTrainLink(train, "Train"),
				train.ID))
	}

	m.Engine.send(text)
}

func (m Messenger) RollbackInfo(user *types.User) {
	text := "Make sure to extend the latest train with the fix / revert and unblock when ready."
	if user != nil {
		text = fmt.Sprintf("%s: %s", m.Engine.formatUser(user), text)
	}

	m.Engine.send(text)
}

func (m Messenger) JobFailed(job *types.Job) {
	if job.Phase.Train.Done || !job.Phase.IsInActivePhaseGroup() {
		// Don't notify if the train is done or if the job is not for the active phase group.
		return
	}
	if job.Phase.Type == types.Deploy {
		if job.Phase.Train.Blocked || job.Phase.Train.CancelledAt.HasValue() {
			// Don't notify deploy failures if the train is blocked or cancelled.
			// This is likely to happen in the event of a rollback.
			return
		}
	}
	jobFailedText := fmt.Sprintf("%s job failed", m.Engine.formatMonospaced(job.Name))
	if job.URL != nil {
		jobFailedText = m.Engine.formatLink(*job.URL, jobFailedText)
	}
	message := fmt.Sprintf("%s. Check failure and consider restarting the job.", jobFailedText)
	engineer := job.Phase.Train.Engineer
	if engineer != nil && job.Phase.Train.Closed {
		// Add @mention for the train engineer if the train is closed.
		message = fmt.Sprintf("%s: %s",
			m.Engine.formatNameEmailNotification(engineer.Name, engineer.Email),
			message)
	}
	m.Engine.send(m.Engine.formatBold(message))
}

func (m Messenger) formatTrainLink(train *types.Train, text string) string {
	return m.Engine.formatLink(fmt.Sprintf("%s/train/%d", settings.GetHostname(), train.ID), text)
}

type commitSet struct {
	CommitterName  string
	CommitterEmail string
	Commits        []*types.Commit
	Extra          *string
}

func (m Messenger) formatCommitSets(header string, nameFormatting nameFormat, commitSets []commitSet) string {
	var text bytes.Buffer
	text.WriteString(m.Engine.formatBold(fmt.Sprintf("%s:", header)))
	text.WriteString("\n")

	for _, commitSet := range commitSets {
		nameHeader := m.formatNameHeader(commitSet.CommitterName, commitSet.CommitterEmail, commitSet.Extra, nameFormatting)
		if nameHeader != nil {
			text.WriteString(fmt.Sprintf("%s\n", *nameHeader))
		}
		for _, commit := range commitSet.Commits {
			commitMessage := strings.Split(commit.Message, "\n")[0]
			commitMessage = m.Engine.escape(commitMessage)

			commitLink := m.Engine.formatLink(commit.URL,
				fmt.Sprintf("%s - %s",
					m.Engine.formatMonospaced(commit.ShortSHA()),
					commitMessage))

			text.WriteString(fmt.Sprintf("%s\n",
				m.Engine.indent(commitLink)))
		}
	}

	return text.String()
}

func (m Messenger) formatNameHeader(name, email string, extra *string, format nameFormat) *string {
	if format == None {
		return extra
	}
	var header string
	switch format {
	case Notify:
		header = m.Engine.formatNameEmailNotification(name, email)
	case PlainText:
		header = m.Engine.formatNameEmail(name, email)
	}
	if extra != nil {
		header = fmt.Sprintf("%s - %s", name, *extra)
	}
	return &header
}

// Groups commits by author email.
// If forStaging is true, this will filter commits that don't require staging notifications.
func (m Messenger) commitsByAuthorEmail(commits []*types.Commit, forStaging bool) map[string][]*types.Commit {
	commitsPerAuthorEmail := make(map[string][]*types.Commit)
	for _, commit := range commits {
		if settings.IsRobotUser(commit.AuthorEmail) {
			continue
		}
		if forStaging && !commit.DoesCommitNeedStagingNotification() {
			continue
		}
		if _, found := commitsPerAuthorEmail[commit.AuthorEmail]; !found {
			commitsPerAuthorEmail[commit.AuthorEmail] = make([]*types.Commit, 0)
		}
		commitsPerAuthorEmail[commit.AuthorEmail] = append(
			commitsPerAuthorEmail[commit.AuthorEmail], commit)
	}
	return commitsPerAuthorEmail
}

func (m Messenger) commitSetsFromCommits(commits []*types.Commit, forStaging bool) []commitSet {
	commitsPerAuthorEmailMap := m.commitsByAuthorEmail(commits, forStaging)
	commitSets := make([]commitSet, 0)
	for _, commits := range commitsPerAuthorEmailMap {
		commitSets = append(commitSets, commitSet{
			CommitterName:  commits[0].AuthorName,
			CommitterEmail: commits[0].AuthorEmail,
			Commits:        commits,
		})
	}
	return commitSets
}

func (m Messenger) commitSetsFromCommitsAndTickets(
	commits []*types.Commit, tickets []*types.Ticket) (ticketedCommitSets, unticketedCommitSets []commitSet) {

	// Add commit sets for each ticket, with the ticket link in the extra field.
	// While doing this, remember the commits that are referenced in a ticket.
	ticketedCommitSets = make([]commitSet, 0)
	ticketedCommitsBySHA := make(map[string]struct{})
	for _, ticket := range tickets {
		// Edge case - should never have a ticket with zero commits.
		if len(ticket.Commits) == 0 {
			logger.Error("Ticket %s has no commits", ticket.Key)
			continue
		}
		if settings.IsRobotUser(ticket.AssigneeEmail) {
			continue
		}
		ticketLink := m.Engine.formatLink(ticket.URL, ticket.Key)
		ticketedCommitSets = append(ticketedCommitSets, commitSet{
			CommitterName:  ticket.Commits[0].AuthorName,
			CommitterEmail: ticket.Commits[0].AuthorEmail,
			Commits:        ticket.Commits,
			Extra:          &ticketLink,
		})
		for _, commit := range ticket.Commits {
			ticketedCommitsBySHA[commit.SHA] = struct{}{}
		}
	}

	// Add commit sets for any authors that have any unticketed commits.
	unticketedCommitSets = make([]commitSet, 0)
	commitsPerAuthorEmailMap := m.commitsByAuthorEmail(commits, true)
	for _, commits := range commitsPerAuthorEmailMap {
		// Filter commits referenced in a ticket.
		filteredCommits := make([]*types.Commit, 0)
		for _, commit := range commits {
			if _, found := ticketedCommitsBySHA[commit.SHA]; !found {
				filteredCommits = append(filteredCommits, commit)
			}
		}

		if len(filteredCommits) == 0 {
			continue
		}

		if settings.IsRobotUser(filteredCommits[0].AuthorEmail) {
			continue
		}

		unticketedCommitSets = append(unticketedCommitSets, commitSet{
			CommitterName:  filteredCommits[0].AuthorName,
			CommitterEmail: filteredCommits[0].AuthorEmail,
			Commits:        filteredCommits,
		})
	}

	return ticketedCommitSets, unticketedCommitSets
}

func (m Messenger) sendCommitSetsDirectly(header string, commitSets []commitSet) {
	for _, set := range commitSets {
		m.Engine.sendDirect(set.CommitterName, set.CommitterEmail,
			m.formatCommitSets(header, None, []commitSet{set}))
	}
}

var (
	service Service
	getOnce sync.Once
)

func GetService() Service {
	getOnce.Do(func() {
		service = newService()
	})
	return service
}

func newService() Service {
	logger.Info("Using %s implementation for Messaging service", implementationFlag)
	var service Service
	switch implementationFlag {
	case "fake":
		service = newFakeEngine()
	case "slack":
		service = newSlackEngine()
	default:
		panic(fmt.Errorf("Unknown Messaging Implementation: %s", implementationFlag))
	}
	return service
}

type fakeEngine struct{}

func newFakeEngine() *Messenger {
	return &Messenger{
		Engine: fakeEngine{},
	}
}

func (e fakeEngine) send(text string) {
	logger.Info("%s", text)
}

func (e fakeEngine) sendDirect(name, email, text string) {
	logger.Info("%s: %s", name, text)
}

func (e fakeEngine) formatUser(user *types.User) string {
	return user.Name
}

func (e fakeEngine) formatNameEmail(name, email string) string {
	return name
}

func (e fakeEngine) formatNameEmailNotification(name, email string) string {
	return name
}

func (e fakeEngine) formatLink(url, text string) string {
	return fmt.Sprintf("%s: %s", text, url)
}

func (e fakeEngine) formatBold(text string) string {
	return text
}

func (e fakeEngine) formatMonospaced(text string) string {
	return text
}

func (e fakeEngine) indent(text string) string {
	return fmt.Sprintf("  %s", text)
}

func (e fakeEngine) escape(text string) string {
	return text
}
