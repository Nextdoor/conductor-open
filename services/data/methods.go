package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"

	"github.com/Nextdoor/conductor/shared/datadog"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/settings"
	"github.com/Nextdoor/conductor/shared/types"
)

type data struct {
	// Overridden by implementation.
	RegisterDB func() error
}

type dataClient struct {
	Client orm.Ormer
}

func (d *data) initialize() {
	orm.DefaultTimeLoc = time.Local

	// Register models.
	orm.RegisterModelWithPrefix(tablePrefix, new(types.Config))
	orm.RegisterModelWithPrefix(tablePrefix, new(types.Train))
	orm.RegisterModelWithPrefix(tablePrefix, new(types.Phase))
	orm.RegisterModelWithPrefix(tablePrefix, new(types.PhaseGroup))
	orm.RegisterModelWithPrefix(tablePrefix, new(types.Job))
	orm.RegisterModelWithPrefix(tablePrefix, new(types.Commit))
	orm.RegisterModelWithPrefix(tablePrefix, new(types.Ticket))
	orm.RegisterModelWithPrefix(tablePrefix, new(types.User))
	orm.RegisterModelWithPrefix(tablePrefix, new(types.Auth))
	orm.RegisterModelWithPrefix(tablePrefix, new(types.Metadata))

	err := d.RegisterDB()
	if err != nil {
		panic(err)
	}

	orm.SetMaxOpenConns("default", 10)

	force := false
	verbose := false
	err = orm.RunSyncdb("default", force, verbose)
	if err != nil {
		panic(err)
	}
}

func (d *data) Client() Client {
	return &dataClient{
		Client: orm.NewOrm(),
	}
}

/* Config */

// Reads the config, and transparently ensures the config is valid.
// If no config is set, inserts the default config.
// If the config is invalid, replaces it with the default config, logging the errant config.
func (d *dataClient) Config() (*types.Config, error) {
	config := &types.Config{ID: 1}
	err := d.Client.Read(config)
	if err != nil {
		if err != orm.ErrNoRows {
			// Error reading config.
			if err != nil {
				return nil, err
			}
		}

		// No config found.
		err = d.InsertDefaultConfig()
		if err != nil {
			return nil, err
		}
		config = types.DefaultConfig
	}

	// Need to check config.Options.ValidationError, which is where parsing errors occur on options.
	// See: SetRaw in shared/types/options.go
	options := config.Options
	// Make sure options were valid.
	if options.ValidationError != nil {
		// There was a validation error.
		logger.Error("Options validation error; replacing with defaults. Options: %s, Error: %v",
			options.InvalidOptionsString,
			options.ValidationError.Error())

		err = d.SetDefaultConfig()
		if err != nil {
			return nil, err
		}

		return types.DefaultConfig, nil
	}
	return config, nil
}

func (d *dataClient) SetDefaultConfig() error {
	_, err := d.Client.Update(types.DefaultConfig)
	return err
}

func (d *dataClient) InsertDefaultConfig() error {
	// Insert default config.
	_, err := d.Client.Insert(types.DefaultConfig)
	if err != nil {
		return err
	}
	return nil
}

func (d *dataClient) Mode() (types.Mode, error) {
	config, err := d.Config()
	if err != nil {
		return types.Schedule, err
	}
	return config.Mode, nil
}

func (d *dataClient) SetMode(mode types.Mode) error {
	config, err := d.Config()
	if err != nil {
		return err
	}
	config.Mode = mode
	_, err = d.Client.Update(config, "Mode")
	return err
}

func (d *dataClient) Options() (*types.Options, error) {
	config, err := d.Config()
	if err != nil {
		return nil, err
	}
	return &config.Options, nil
}

func (d *dataClient) SetOptions(options *types.Options) error {
	config, err := d.Config()
	if err != nil {
		return err
	}
	config.Options = *options
	_, err = d.Client.Update(config, "Options")
	return err
}

func (d *dataClient) InCloseTime() (bool, error) {
	config, err := d.Config()
	if err != nil {
		return false, err
	}
	return config.Options.InCloseTime(), nil
}

func (d *dataClient) IsTrainAutoCloseable(train *types.Train) (bool, error) {
	mode, err := d.Mode()
	if err != nil {
		err = fmt.Errorf("Error getting Mode: %v", err)
		return false, err
	}
	if mode == types.Manual {
		return false, nil
	}
	inCloseTime, err := d.InCloseTime()
	if err != nil {
		err = fmt.Errorf("Error getting InCloseTime: %v", err)
		return false, err
	}
	return inCloseTime && train.Engineer != nil && !train.ScheduleOverride, nil
}

/* Train */

func (d *dataClient) Train(trainID uint64) (*types.Train, error) {
	train := types.Train{ID: trainID}
	err := d.Client.Read(&train)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	err = d.loadTrainRelated(&train)
	if err != nil {
		return nil, err
	}

	return &train, nil
}

func (d *dataClient) LatestTrain() (*types.Train, error) {
	train := &types.Train{}
	query := d.Client.QueryTable(train)
	query = query.OrderBy("-id")
	err := query.One(train)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	err = d.loadTrainRelated(train)
	if err != nil {
		return nil, err
	}
	return train, nil
}

func (d *dataClient) adjacentTrains(trainID uint64) (*types.Train, *types.Train, error) {
	// Get previous train.
	previousTrain := &types.Train{}
	err := d.Client.QueryTable(previousTrain).Filter("id__lt", trainID).OrderBy("-id").One(previousTrain)
	if err != nil {
		if err != orm.ErrNoRows {
			return nil, nil, err
		} else {
			previousTrain = nil
		}
	}

	// Get next train.
	nextTrain := &types.Train{}
	err = d.Client.QueryTable(nextTrain).Filter("id__gt", trainID).OrderBy("id").One(nextTrain)
	if err != nil {
		if err != orm.ErrNoRows {
			return nil, nil, err
		} else {
			nextTrain = nil
		}
	}

	return previousTrain, nextTrain, nil
}

func (d *dataClient) LatestTrainForBranch(branch string) (*types.Train, error) {
	train := &types.Train{}
	query := d.Client.QueryTable(train)
	query = query.Filter("branch", branch)
	query = query.OrderBy("-id")
	err := query.One(train)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	err = d.loadTrainRelated(train)
	if err != nil {
		return nil, err
	}
	return train, nil
}

func (d *dataClient) CreateTrain(branch string, engineer *types.User, commits []*types.Commit) (*types.Train, error) {
	if len(commits) == 0 {
		return nil, errors.New("Cannot create a train with no commits.")
	}

	err := d.Client.Begin()
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}

	_, err = d.WriteCommits(commits)
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}

	train := &types.Train{
		Branch:   branch,
		TailSHA:  commits[0].SHA,
		HeadSHA:  commits[len(commits)-1].SHA,
		Engineer: engineer,
	}

	closeable, err := d.IsTrainAutoCloseable(train)
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}
	if closeable {
		train.Closed = true
	}

	phaseGroup, err := d.createPhaseGroup(train)
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}
	train.ActivePhases = phaseGroup

	_, err = d.Client.Insert(train)
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}

	phaseGroup.Train = train
	_, err = d.Client.Update(phaseGroup)
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}

	m2m := d.Client.QueryM2M(train, "Commits")
	for _, commit := range commits {
		_, err = m2m.Add(commit)
		if err != nil {
			d.Client.Rollback()
			return nil, err
		}
	}

	// Populate train.
	err = d.loadTrainRelated(train)
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}

	err = d.Client.Commit()
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}
	datadog.Info("Created train (ID, Branch, HeadSHA, TailSHA) %v, %v, %v, %v", train.ID, train.Branch, train.HeadSHA, train.TailSHA)
	return train, nil
}

func (d *dataClient) ExtendTrain(train *types.Train, engineer *types.User, newCommits []*types.Commit) error {
	if len(newCommits) == 0 {
		return errors.New("Cannot extend a train with no new commits.")
	}

	err := d.Client.Begin()
	if err != nil {
		d.Client.Rollback()
		return err
	}

	_, err = d.WriteCommits(newCommits)
	if err != nil {
		d.Client.Rollback()
		return err
	}

	train.HeadSHA = newCommits[len(newCommits)-1].SHA
	train.Engineer = engineer

	closeable, err := d.IsTrainAutoCloseable(train)
	if err != nil {
		d.Client.Rollback()
		return err
	}
	if closeable {
		train.Closed = true
	}

	phaseGroup, err := d.createPhaseGroup(train)
	if err != nil {
		d.Client.Rollback()
		return err
	}
	train.ActivePhases = phaseGroup

	_, err = d.Client.Update(train)
	if err != nil {
		d.Client.Rollback()
		return err
	}

	phaseGroup.Train = train
	_, err = d.Client.Update(phaseGroup)
	if err != nil {
		d.Client.Rollback()
		return err
	}

	m2m := d.Client.QueryM2M(train, "Commits")
	for _, commit := range newCommits {
		_, err = m2m.Add(commit)
		if err != nil {
			d.Client.Rollback()
			return err
		}
	}

	// Populate train.
	err = d.loadTrainRelated(train)
	if err != nil {
		d.Client.Rollback()
		return err
	}

	err = d.Client.Commit()
	if err != nil {
		d.Client.Rollback()
		return err
	}
	datadog.Info("Extended train (ID, HeadSHA, TailSHA): %v, %v, %v", train.ID, train.HeadSHA, train.TailSHA)
	return nil
}

func (d *dataClient) DuplicateTrain(oldTrain *types.Train, newCommits []*types.Commit) (*types.Train, error) {
	err := d.Client.Begin()
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}

	// Clone the old train.
	newTrain := &types.Train{
		Branch:   oldTrain.Branch,
		TailSHA:  oldTrain.TailSHA,
		HeadSHA:  oldTrain.HeadSHA,
		Engineer: oldTrain.Engineer,
	}

	if len(newCommits) > 0 {
		_, err = d.WriteCommits(newCommits)
		if err != nil {
			d.Client.Rollback()
			return nil, err
		}

		newTrain.HeadSHA = newCommits[len(newCommits)-1].SHA
	}

	if oldTrain.ScheduleOverride {
		newTrain.Closed = oldTrain.Closed
		newTrain.ScheduleOverride = true
	} else {
		closeable, err := d.IsTrainAutoCloseable(newTrain)
		if err != nil {
			d.Client.Rollback()
			return nil, err
		}
		if closeable {
			newTrain.Closed = true
		}
	}

	phaseGroup, err := d.createPhaseGroup(newTrain)
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}
	newTrain.ActivePhases = phaseGroup

	_, err = d.Client.Insert(newTrain)
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}

	phaseGroup.Train = newTrain
	_, err = d.Client.Update(phaseGroup)
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}

	// Clone the old commit mappings.
	commits := make([]*types.Commit, len(oldTrain.Commits)+len(newCommits))
	index := 0
	for i := range oldTrain.Commits {
		commits[index] = oldTrain.Commits[i]
		index += 1
	}
	for i := range newCommits {
		commits[index] = newCommits[i]
		index += 1
	}

	m2m := d.Client.QueryM2M(newTrain, "Commits")
	for _, commit := range commits {
		_, err = m2m.Add(commit)
		if err != nil {
			d.Client.Rollback()
			return nil, err
		}
	}

	// Clone the old tickets.
	tickets := make([]*types.Ticket, len(oldTrain.Tickets))
	for i := range oldTrain.Tickets {
		newTicket := oldTrain.Tickets[i]
		newTicket.ID = 0
		newTicket.Train = newTrain
		tickets[i] = newTicket
	}

	err = d.WriteTickets(tickets)
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}

	// Populate train.
	err = d.loadTrainRelated(newTrain)
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}

	err = d.Client.Commit()
	if err != nil {
		d.Client.Rollback()
		return nil, err
	}
	datadog.Info("Duplicated train (ID, HeadSHA, TailSHA) %v, %v, %v", newTrain.ID, newTrain.HeadSHA, newTrain.TailSHA)
	return newTrain, nil
}

func (d *dataClient) ChangeTrainEngineer(train *types.Train, engineer *types.User) error {
	train.Engineer = engineer
	_, err := d.Client.Update(train, "Engineer")
	if err == nil {
		datadog.Info("Changed train engineer to %v", engineer.Name)
	}
	return err
}

func (d *dataClient) CloseTrain(train *types.Train, override bool) error {
	train.Closed = true
	train.ScheduleOverride = override
	_, err := d.Client.Update(train, "Closed", "ScheduleOverride")
	if err == nil {
		datadog.Info("Closed train (ID, HeadSHA, TailSHA) %v, %v, %v", train.ID, train.HeadSHA, train.TailSHA)
	}
	return err
}

func (d *dataClient) OpenTrain(train *types.Train, override bool) error {
	train.Closed = false
	train.ScheduleOverride = override
	_, err := d.Client.Update(train, "Closed", "ScheduleOverride")
	if err == nil {
		datadog.Info("Opened train (ID, HeadSHA, TailSHA) %v, %v, %v", train.ID, train.HeadSHA, train.TailSHA)
	}
	return err
}

func (d *dataClient) BlockTrain(train *types.Train, reason *string) error {
	train.Blocked = true
	train.BlockedReason = reason
	_, err := d.Client.Update(train, "Blocked", "BlockedReason")
	if err == nil {
		datadog.Info("Blocked train (ID, HeadSHA, TailSHA) %v, %v, %v", train.ID, train.HeadSHA, train.TailSHA)
	}
	return err
}

func (d *dataClient) UnblockTrain(train *types.Train) error {
	train.Blocked = false
	_, err := d.Client.Update(train, "Blocked")
	if err == nil {
		datadog.Info("Unblocked train (ID, HeadSHA, TailSHA) %v, %v, %v", train.ID, train.HeadSHA, train.TailSHA)
	}
	return err
}

func (d *dataClient) DeployTrain(train *types.Train) error {
	train.DeployedAt = types.Time{time.Now()}
	_, err := d.Client.Update(train, "DeployedAt")
	if err == nil {
		datadog.Info("Deployed train (ID, HeadSHA, TailSHA) %v, %v, %v", train.ID, train.HeadSHA, train.TailSHA)
	}
	return err
}

func (d *dataClient) CancelTrain(train *types.Train) error {
	train.CancelledAt = types.Time{time.Now()}
	_, err := d.Client.Update(train, "CancelledAt")
	if err == nil {
		datadog.Info("Cancelled train (ID, HeadSHA, TailSHA) %v, %v, %v", train.ID, train.HeadSHA, train.TailSHA)
	}
	return err
}

func (d *dataClient) loadTrainRelated(train *types.Train) error {
	if train.Engineer != nil {
		_, err := d.Client.LoadRelated(train, "Engineer")
		if err != nil {
			return err
		}
	}

	_, err := d.Client.LoadRelated(train, "Tickets")
	if err != nil {
		return err
	}
	sort.Sort(types.TicketsByID(train.Tickets))

	for _, ticket := range train.Tickets {
		_, err := d.Client.LoadRelated(ticket, "Commits")
		if err != nil {
			return err
		}
		sort.Sort(types.CommitsByID(ticket.Commits))
	}

	_, err = d.Client.LoadRelated(train, "Commits")
	if err != nil {
		return err
	}
	sort.Sort(types.CommitsByID(train.Commits))

	_, err = d.Client.LoadRelated(train, "ActivePhases", 1)
	if err != nil {
		return err
	}

	for _, phase := range train.ActivePhases.Phases() {
		_, err = d.Client.LoadRelated(phase, "Jobs")
		sort.Sort(types.JobsByID(phase.Jobs))
		if err != nil {
			return err
		}
	}

	train.ActivePhases.SetReferences(train)

	train.SetActivePhase()

	previousTrain, nextTrain, err := d.adjacentTrains(train.ID)
	if err != nil {
		return err
	}

	if previousTrain != nil {
		train.PreviousID = &previousTrain.ID
		train.PreviousTrainDone = previousTrain.IsDone()
	} else {
		train.PreviousTrainDone = false
	}

	if nextTrain != nil {
		train.NextID = &nextTrain.ID
	}

	train.NotDeployableReason = train.GetNotDeployableReason()

	train.Done = train.IsDone()

	train.CanRollback = train.Done && settings.GetJenkinsRollbackJob() != ""

	return nil
}

func (d *dataClient) loadAllTrainPhaseGroups(train *types.Train) error {
	if train.AllPhaseGroups != nil {
		// Already loaded.
		return nil
	}

	_, err := d.Client.LoadRelated(train, "AllPhaseGroups", 2)
	if err != nil {
		return err
	}

	for _, phaseGroup := range train.AllPhaseGroups {
		for _, phase := range phaseGroup.Phases() {
			_, err = d.Client.LoadRelated(phase, "Jobs")
			if err != nil {
				return err
			}
			sort.Sort(types.JobsByID(phase.Jobs))
		}
		phaseGroup.SetReferences(train)
	}

	return nil
}

func (d *dataClient) LoadLastDeliveredSHA(train *types.Train) error {
	if train.LastDeliveredSHA != nil {
		// Already loaded.
		return nil
	}

	if train.AllPhaseGroups == nil {
		err := d.loadAllTrainPhaseGroups(train)
		if err != nil {
			return err
		}
	}

	// Find LastDeliveredSHA based on phase groups.
	if len(train.AllPhaseGroups) <= 1 {
		train.LastDeliveredSHA = nil
	} else {
		for i := len(train.AllPhaseGroups) - 2; i >= 0; i-- {
			previousPhaseGroup := train.AllPhaseGroups[i]
			if previousPhaseGroup.Delivery.CompletedAt.HasValue() {
				train.LastDeliveredSHA = &previousPhaseGroup.HeadSHA
				break
			}
		}
	}

	return nil
}

/* Phase */

func (d *dataClient) Phase(phaseID uint64, train *types.Train) (*types.Phase, error) {
	err := d.loadAllTrainPhaseGroups(train)
	if err != nil {
		return nil, err
	}

	phaseGroups := make([]*types.PhaseGroup, 1+len(train.AllPhaseGroups))
	phaseGroups[0] = train.ActivePhases
	for i, phaseGroup := range train.AllPhaseGroups {
		phaseGroups[i+1] = phaseGroup
	}
	var phase *types.Phase
	for _, phaseGroup := range phaseGroups {
		if phaseGroup.Delivery.ID == phaseID {
			phase = phaseGroup.Delivery
			break
		} else if phaseGroup.Verification.ID == phaseID {
			phase = phaseGroup.Verification
			break
		} else if phaseGroup.Deploy.ID == phaseID {
			phase = phaseGroup.Deploy
			break
		}
	}
	if phase != nil {
		_, err := d.Client.LoadRelated(phase, "Jobs")
		if err != nil {
			return nil, err
		}
		for _, job := range phase.Jobs {
			job.Phase = phase
		}
		phase.Train = train
		return phase, nil
	}
	return nil, fmt.Errorf("No phase with ID %d found for train %d", phaseID, train.ID)
}

func (d *dataClient) StartPhase(phase *types.Phase) error {
	phase.StartedAt = types.Time{time.Now()}
	_, err := d.Client.Update(phase, "StartedAt")
	phase.Train.SetActivePhase()
	if err == nil {
		datadog.Info("Started phase %v", phase.ID)
	}
	return err
}

func (d *dataClient) ErrorPhase(phase *types.Phase, phaseErr error) error {
	phase.Error = phaseErr.Error()
	_, err := d.Client.Update(phase, "Error")
	if err == nil {
		datadog.Error("Phase %v error", phase)
	}
	return err
}

func (d *dataClient) UncompletePhase(phase *types.Phase) error {
	phase.CompletedAt = types.Time{}
	_, err := d.Client.Update(phase, "CompletedAt")
	if err == nil {
		datadog.Error("Phase %v uncomplete", phase.ID)
	}
	return err
}

func (d *dataClient) CompletePhase(phase *types.Phase) error {
	phase.CompletedAt = types.Time{time.Now()}
	_, err := d.Client.Update(phase, "CompletedAt")
	if err == nil {
		datadog.Info("Completed phase %v", phase.ID)
	}
	return err
}

func (d *dataClient) ReplacePhase(phase *types.Phase) (*types.Phase, error) {
	newPhase := phase.PhaseGroup.AddNewPhase(phase.Type, phase.Train)
	_, err := d.Client.Insert(newPhase)
	if err != nil {
		return nil, err
	}
	err = d.createPhaseJobs(newPhase)
	if err != nil {
		return nil, err
	}
	_, err = d.Client.Update(newPhase.PhaseGroup, newPhase.Type.String())
	if err != nil {
		return nil, err
	}
	datadog.Info("Replaced phase %v", phase.ID)
	return newPhase, nil
}

func (d *dataClient) createPhaseGroup(train *types.Train) (*types.PhaseGroup, error) {
	phaseGroup := &types.PhaseGroup{HeadSHA: train.HeadSHA}
	phaseGroup.AddNewPhase(types.Delivery, train)
	phaseGroup.AddNewPhase(types.Verification, train)
	phaseGroup.AddNewPhase(types.Deploy, train)

	_, err := d.Client.Insert(phaseGroup.Delivery)
	if err != nil {
		return nil, err
	}
	_, err = d.Client.Insert(phaseGroup.Verification)
	if err != nil {
		return nil, err
	}
	_, err = d.Client.Insert(phaseGroup.Deploy)
	if err != nil {
		return nil, err
	}

	err = d.createPhaseJobs(phaseGroup.Delivery)
	if err != nil {
		return nil, err
	}
	err = d.createPhaseJobs(phaseGroup.Verification)
	if err != nil {
		return nil, err
	}
	err = d.createPhaseJobs(phaseGroup.Deploy)
	if err != nil {
		return nil, err
	}

	_, err = d.Client.Insert(phaseGroup)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	datadog.Info("Created phase group (ID, HeadSHA) %v, %v", phaseGroup.ID, phaseGroup.HeadSHA)
	return phaseGroup, nil
}

/* Job */
func (d *dataClient) CreateJob(phase *types.Phase, name string) (*types.Job, error) {
	job := types.Job{Name: name, Phase: phase}
	_, err := d.Client.Insert(&job)
	if err != nil {
		return nil, err
	}
	datadog.Info("Created job (ID, Name) %v, %v", job.ID, job.Name)
	return &job, nil
}

func (d *dataClient) StartJob(job *types.Job, url string) error {
	job.StartedAt = types.Time{time.Now()}
	job.URL = &url
	_, err := d.Client.Update(job, "StartedAt", "URL")
	if err == nil {
		datadog.Info("Started job (ID, Name) %v, %v", job.ID, job.Name)
	}
	return err
}

// Mark an job as finished and set its result
func (d *dataClient) CompleteJob(job *types.Job, result types.JobResult, metadata string) error {
	job.CompletedAt = types.Time{time.Now()}
	job.Result = result
	job.Metadata = metadata
	_, err := d.Client.Update(job, "CompletedAt", "Result", "Metadata")
	if err == nil {
		datadog.Info("Completed job (ID, Name) %v, %v", job.ID, job.Name)
	}
	return err
}

func (d *dataClient) RestartJob(job *types.Job, url string) error {
	job.StartedAt = types.Time{time.Now()}
	job.URL = &url
	job.CompletedAt = types.Time{}
	job.Result = types.JobResult(0)
	job.Metadata = ""
	_, err := d.Client.Update(job,
		"StartedAt", "URL",
		"CompletedAt", "Result", "Metadata")
	if err == nil {
		datadog.Info("Restarted job (ID, Name) %v, %v", job.ID, job.Name)
	}
	return err
}

func (d *dataClient) createPhaseJobs(phase *types.Phase) error {
	for _, jobName := range types.JobsForPhase(phase.Type) {
		_, err := d.CreateJob(phase, jobName)
		if err != nil {
			return err
		}
	}
	return nil
}

/* Commit */
func (d *dataClient) WriteCommits(commits []*types.Commit) ([]*types.Commit, error) {
	newCommits := make([]*types.Commit, 0)
	wrote := make([]string, 0)
	for _, commit := range commits {
		if commit.ID > 0 {
			// Already written.
			continue
		}
		created, _, err := d.Client.ReadOrCreate(commit, "SHA")
		if err != nil {
			return nil, err
		}
		if created {
			newCommits = append(newCommits, commit)
			wrote = append(wrote, fmt.Sprintf("(ID, SHA, Branch, AuthorName) %v, %v, %v, %v", commit.ID, commit.SHA, commit.Branch, commit.AuthorName))
		}
	}
	datadog.Info("Wrote commits: %v", strings.Join(wrote, "\n"))
	return newCommits, nil
}

func (d *dataClient) LatestCommitForTrain(train *types.Train) (*types.Commit, error) {
	commit := &types.Commit{}
	query := d.Client.QueryTable(commit)
	query = query.Filter("SHA", train.HeadSHA)
	err := query.One(commit)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (d *dataClient) TrainsByCommit(commit *types.Commit) ([]*types.Train, error) {
	trains := make([]*types.Train, 0)
	_, err := d.Client.QueryTable(types.Train{}).
		Filter("Commits__ConductorCommit__SHA__contains", commit.SHA).
		Distinct().
		OrderBy("-id").
		All(&trains)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return trains, nil
}

/* User */

func (d *dataClient) WriteToken(newToken, name, email, avatar, codeToken string) error {
	user := types.User{
		Email: email,
	}

	// Read by email.
	_, _, err := d.Client.ReadOrCreate(&user, "Email")
	if err != nil {
		return err
	}

	auth := types.Auth{
		User:      &user,
		Token:     newToken,
		CodeToken: codeToken,
	}

	tokens := make([]*types.Auth, 0)
	query := d.Client.QueryTable(&types.Auth{})
	query = query.Filter("User", &user)
	_, err = query.All(&tokens)
	if err != nil && err != orm.ErrNoRows {
		return err
	} else {
		// Insert a new token (including the code token).
		_, err = d.Client.Insert(&auth)
		if err != nil && !(err.Error() == "LastInsertId is not supported by this driver" || err.Error() == "no LastInsertId available") {
			return err
		}
	}
	// Update name and avatar.
	user.Name = name
	user.AvatarURL = avatar
	_, err = d.Client.Update(&user)
	if err != nil {
		return err
	}
	datadog.Info("Wrote token %v", auth.Token)
	return nil
}

func (d *dataClient) RevokeToken(oldToken, email string) error {
	// Find auth data for token.
	auth := types.Auth{
		Token: oldToken,
	}
	err := d.Client.Read(&auth, "Token")
	if err != nil {
		return err
	}

	// Verify email / token combination.
	_, err = d.Client.LoadRelated(&auth, "User")
	if err != nil {
		return err
	}
	if auth.User.Email != email {
		return errors.New("Token and email don't match.")
	}

	// Delete.
	_, err = d.Client.Delete(&auth)
	datadog.Info("Revoked token %v", auth.Token)
	return err
}

func (d *dataClient) ReadOrCreateUser(name, email string) (*types.User, error) {
	user := types.User{Name: name, Email: email}

	// Read by email.
	_, _, err := d.Client.ReadOrCreate(&user, "Email")
	if err != nil {
		return nil, err
	}
	datadog.Info("Read/Created user (ID, Name) %v, %v", user.ID, user.Name)
	return &user, nil
}

func (d *dataClient) UserByToken(token string) (*types.User, error) {
	auth := types.Auth{Token: token}
	err := d.Client.Read(&auth)
	if err != nil {
		return nil, err
	}

	_, err = d.Client.LoadRelated(&auth, "User")
	if err != nil {
		return nil, err
	}

	auth.User.Token = token
	return auth.User, nil
}

/* Ticket */
func (d *dataClient) WriteTickets(tickets []*types.Ticket) error {
	wrote := make([]string, 0)
	for i := range tickets {
		ticket := tickets[i]
		_, err := d.Client.Insert(ticket)
		if err != nil {
			return err
		}

		ticketCommits := ticket.Commits
		m2m := d.Client.QueryM2M(ticket, "Commits")
		for _, commit := range ticketCommits {
			_, err = m2m.Add(commit)
			if err != nil {
				return err
			}
		}
		_, err = d.Client.LoadRelated(ticket, "Commits")
		if err != nil {
			return err
		}
		wrote = append(wrote, fmt.Sprintf("(ID, Summary, AssigneeName) %v, %v, %v", ticket.ID, ticket.Summary, ticket.AssigneeName))
	}
	if len(wrote) > 0 {
		datadog.Info("Wrote tickets: %v", strings.Join(wrote, "\n"))
	}
	return nil
}

func (d *dataClient) UpdateTickets(tickets []*types.Ticket) error {
	updated := make([]string, 0)
	for i := range tickets {
		ticket := tickets[i]
		_, err := d.Client.Update(ticket)
		if err != nil {
			return err
		}
		updated = append(updated, fmt.Sprintf("(ID, Summary, AssigneeName) %v %v %v", ticket.ID, ticket.Summary, ticket.AssigneeName))
	}
	if len(updated) > 0 {
		datadog.Info("Updated tickets: %v", strings.Join(updated, "\n"))
	}
	return nil
}

/* Metadata */

var ErrNoSuchNamespaceOrKey = errors.New("No such namespace or key")

func (d *dataClient) MetadataListNamespaces() ([]string, error) {
	query := d.Client.Raw("SELECT namespace FROM conductor_metadata")
	results := make([]*types.Metadata, 0)
	_, err := query.QueryRows(&results)
	if err != nil {
		return nil, err
	}

	namespaces := make([]string, len(results))
	for i, result := range results {
		namespaces[i] = result.Namespace
	}

	return namespaces, nil
}

func (d *dataClient) MetadataListKeys(namespace string) ([]string, error) {
	query := d.Client.Raw(`
		SELECT jsonb_object_keys(data)
		FROM conductor_metadata
		WHERE namespace = ?`,
		namespace)
	results := make([]string, 0)
	_, err := query.QueryRows(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (d *dataClient) MetadataGetKey(namespace, key string) (string, error) {
	query := d.Client.Raw(`
		SELECT data ->> ? FROM conductor_metadata
		WHERE namespace = ?`,
		key, namespace)
	var value string
	err := query.QueryRow(&value)
	if err != nil {
		if err == orm.ErrNoRows {
			return "", ErrNoSuchNamespaceOrKey
		}
		return "", err
	}
	if value == "" {
		return "", ErrNoSuchNamespaceOrKey
	}
	return value, nil
}

func (d *dataClient) MetadataSet(namespace string, newData map[string]string) error {
	b, _ := json.Marshal(newData)
	jsonData := string(b)
	query := d.Client.Raw(`
		INSERT INTO conductor_metadata
		(namespace, data)
		VALUES (?, ?)
		ON CONFLICT(namespace) DO UPDATE
			SET data = (
				SELECT data FROM conductor_metadata
				WHERE namespace = ?) || ?`,
		namespace, jsonData,
		namespace, jsonData)
	_, err := query.Exec()
	return err
}

func (d *dataClient) MetadataDeleteNamespace(namespace string) error {
	query := d.Client.Raw(`
		DELETE FROM conductor_metadata
		WHERE namespace = ?`,
		namespace)
	_, err := query.Exec()
	return err
}

func (d *dataClient) MetadataDeleteKey(namespace, key string) error {
	query := d.Client.Raw(`
		UPDATE conductor_metadata
		SET data = (
			SELECT data FROM conductor_metadata
			WHERE namespace = ?) - ?
		WHERE namespace = ?`,
		namespace, key, namespace)
	_, err := query.Exec()
	return err
}
