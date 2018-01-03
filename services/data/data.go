/* Handles interfacing with the data store. */
package data

import (
	"fmt"
	"sync"

	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	implementationFlag = flags.EnvString("DATA_IMPL", "postgres")
	tablePrefix        = flags.EnvString("TABLE_PREFIX", "conductor_")
)

type Service interface {
	Client() Client
}

type Client interface {
	Config() (*types.Config, error)

	Mode() (types.Mode, error)
	SetMode(types.Mode) error

	Options() (*types.Options, error)
	SetOptions(*types.Options) error

	InCloseTime() (bool, error)
	IsTrainAutoCloseable(*types.Train) (bool, error)

	Train(uint64) (*types.Train, error)
	LatestTrain() (*types.Train, error)
	LatestTrainForBranch(string) (*types.Train, error)
	CreateTrain(string, *types.User, []*types.Commit) (*types.Train, error)
	ExtendTrain(*types.Train, *types.User, []*types.Commit) error
	DuplicateTrain(*types.Train, []*types.Commit) (*types.Train, error)
	ChangeTrainEngineer(*types.Train, *types.User) error
	CloseTrain(*types.Train, bool) error
	OpenTrain(*types.Train, bool) error
	BlockTrain(*types.Train, *string) error
	UnblockTrain(*types.Train) error
	DeployTrain(*types.Train) error
	CancelTrain(*types.Train) error

	Phase(uint64, *types.Train) (*types.Phase, error)
	StartPhase(*types.Phase) error
	ErrorPhase(*types.Phase, error) error
	UncompletePhase(*types.Phase) error
	CompletePhase(*types.Phase) error
	ReplacePhase(*types.Phase) (*types.Phase, error)

	CreateJob(*types.Phase, string) (*types.Job, error)
	StartJob(*types.Job, string) error
	CompleteJob(*types.Job, types.JobResult, string) error
	RestartJob(*types.Job, string) error

	WriteCommits([]*types.Commit) ([]*types.Commit, error)
	LatestCommitForTrain(*types.Train) (*types.Commit, error)
	TrainsByCommit(*types.Commit) ([]*types.Train, error)

	WriteToken(newToken, name, email, avatar, codeToken string) error
	RevokeToken(oldToken, email string) error
	ReadOrCreateUser(name, email string) (*types.User, error)
	UserByToken(token string) (*types.User, error)

	WriteTickets([]*types.Ticket) error
	UpdateTickets([]*types.Ticket) error

	MetadataListNamespaces() ([]string, error)
	MetadataListKeys(string) ([]string, error)
	MetadataGetKey(string, string) (string, error)
	MetadataSet(string, map[string]string) error
	MetadataDeleteNamespace(string) error
	MetadataDeleteKey(string, string) error
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

func NewClient() Client {
	return GetService().Client()
}

func newService() Service {
	logger.Info("Using %s implementation for Data service", implementationFlag)
	var service Service
	switch implementationFlag {
	case "postgres":
		service = newPostgres()
	default:
		panic(fmt.Errorf("Unknown Data Implementation: %s", implementationFlag))
	}
	return service
}
