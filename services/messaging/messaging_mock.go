package messaging

import (
	"github.com/Nextdoor/conductor/shared/types"
)

type MessagingServiceMock struct {
	Engine                Engine
	TrainCreationMock     func(*types.Train, []*types.Commit)
	TrainExtensionMock    func(*types.Train, []*types.Commit, *types.User)
	TrainDuplicationMock  func(*types.Train, *types.Train, []*types.Commit)
	TrainDeliveredMock    func(*types.Train, []*types.Commit, []*types.Ticket)
	TrainVerifiedMock     func(*types.Train)
	TrainUnverifiedMock   func(*types.Train)
	TrainDeployingMock    func()
	TrainDeployedMock     func(*types.Train)
	TrainClosedMock       func(*types.Train, *types.User)
	TrainOpenedMock       func(*types.Train, *types.User)
	TrainBlockedMock      func(*types.Train, *types.User)
	TrainUnblockedMock    func(*types.Train, *types.User)
	TrainCancelledMock    func(*types.Train, *types.User)
	RollbackInitiatedMock func(*types.Train, *types.User)
	RollbackInfoMock      func(*types.User)
	JobFailedMock         func(*types.Job)
}

func (m MessagingServiceMock) TrainCreation(train *types.Train, commits []*types.Commit) {
	if m.TrainCreationMock != nil {
		m.TrainCreationMock(train, commits)
	}
}

func (m MessagingServiceMock) TrainExtension(train *types.Train, commits []*types.Commit, user *types.User) {
	if m.TrainExtensionMock != nil {
		m.TrainExtensionMock(train, commits, user)
	}
}

func (m MessagingServiceMock) TrainDuplication(train *types.Train, trainFrom *types.Train, commits []*types.Commit) {
	if m.TrainDuplicationMock != nil {
		m.TrainDuplicationMock(train, trainFrom, commits)
	}
}

func (m MessagingServiceMock) TrainDelivered(train *types.Train, commits []*types.Commit, tickets []*types.Ticket) {
	if m.TrainDeliveredMock != nil {
		m.TrainDeliveredMock(train, commits, tickets)
	}
}

func (m MessagingServiceMock) TrainVerified(train *types.Train) {
	if m.TrainVerifiedMock != nil {
		m.TrainVerifiedMock(train)
	}
}

func (m MessagingServiceMock) TrainUnverified(train *types.Train) {
	if m.TrainUnverifiedMock != nil {
		m.TrainUnverifiedMock(train)
	}
}

func (m MessagingServiceMock) TrainDeploying() {
	if m.TrainDeployingMock != nil {
		m.TrainDeployingMock()
	}
}

func (m MessagingServiceMock) TrainDeployed(train *types.Train) {
	if m.TrainDeployedMock != nil {
		m.TrainDeployedMock(train)
	}
}

func (m MessagingServiceMock) TrainClosed(train *types.Train, user *types.User) {
	if m.TrainClosedMock != nil {
		m.TrainClosedMock(train, user)
	}
}

func (m MessagingServiceMock) TrainOpened(train *types.Train, user *types.User) {
	if m.TrainOpenedMock != nil {
		m.TrainOpenedMock(train, user)
	}
}

func (m MessagingServiceMock) TrainBlocked(train *types.Train, user *types.User) {
	if m.TrainBlockedMock != nil {
		m.TrainBlockedMock(train, user)
	}
}

func (m MessagingServiceMock) TrainUnblocked(train *types.Train, user *types.User) {
	if m.TrainUnblockedMock != nil {
		m.TrainUnblockedMock(train, user)
	}
}

func (m MessagingServiceMock) TrainCancelled(train *types.Train, user *types.User) {
	if m.TrainCancelledMock != nil {
		m.TrainCancelledMock(train, user)
	}
}

func (m MessagingServiceMock) RollbackInitiated(train *types.Train, user *types.User) {
	if m.RollbackInitiatedMock != nil {
		m.RollbackInitiatedMock(train, user)
	}
}

func (m MessagingServiceMock) RollbackInfo(user *types.User) {
	if m.RollbackInfoMock != nil {
		m.RollbackInfoMock(user)
	}
}

func (m MessagingServiceMock) JobFailed(job *types.Job) {
	if m.JobFailedMock != nil {
		m.JobFailedMock(job)
	}
}
