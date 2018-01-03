/* Handles creating verification tickets and checking their status. */
package ticket

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	implementationFlag = flags.EnvString("TICKET_IMPL", "fake")
)

const descriptionTemplate = `{{ .AuthorName }}

{{ range .Commits }}	- http://c/{{ .ShortSHA }} - {{ index (Split .Message "\n") 0 }}
{{ end }}`

type Service interface {
	CreateTickets(*types.Train, []*types.Commit) ([]*types.Ticket, error)
	CloseTickets([]*types.Ticket) error
	DeleteTickets(*types.Train) error
	SyncTickets(*types.Train) ([]*types.Ticket, []*types.Ticket, error)
	CloseTrainTickets(*types.Train) error
}

var (
	service               Service
	getOnce               sync.Once
	DefaultTicketUsername string
)

func GetService() Service {
	getOnce.Do(func() {
		service = newService()
	})
	return service
}

func newService() Service {
	logger.Info("Using %s implementation for Ticket service", implementationFlag)
	var service Service
	switch implementationFlag {
	case "fake":
		service = newFake()
	case "jira":
		service = newJIRA()
	default:
		panic(fmt.Errorf("Unknown Verification Implementation: %s", implementationFlag))
	}
	return service
}

type fake struct{}

func newFake() *fake {
	return &fake{}
}

func (t *fake) CreateTickets(train *types.Train, commits []*types.Commit) ([]*types.Ticket, error) {
	return nil, nil
}

func (t *fake) CloseTickets(tickets []*types.Ticket) error {
	return nil
}

func (t *fake) DeleteTickets(train *types.Train) error {
	return nil
}

func (t *fake) SyncTickets(train *types.Train) ([]*types.Ticket, []*types.Ticket, error) {
	return nil, nil, nil
}

func (t *fake) CloseTrainTickets(train *types.Train) error {
	return nil
}

func descriptionFromCommits(commits []*types.Commit) (string, error) {
	var output bytes.Buffer
	tplFuncMap := make(template.FuncMap)
	tplFuncMap["Split"] = strings.Split
	descTemplate, err := template.New("description").Funcs(tplFuncMap).Parse(descriptionTemplate)
	if err != nil {
		return "", err
	}

	authorName := commits[0].AuthorName

	templateCtx := make(map[string]interface{}, 0)
	templateCtx["AuthorName"] = authorName
	templateCtx["Commits"] = commits
	err = descTemplate.Execute(&output, templateCtx)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func summaryForCommit(commit *types.Commit) string {
	return fmt.Sprintf("Verify %s's changes", commit.AuthorName)
}
