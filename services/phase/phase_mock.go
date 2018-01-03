package phase

import (
	"github.com/Nextdoor/conductor/shared/types"
)

type PhaseServiceMock struct {
	StartMock func(
		phaseType types.PhaseType, trainID,
		deliveryPhaseID, verificationPhaseID, deployPhaseID uint64, branch, sha string,
		buildUser *types.User) error
}

func (m *PhaseServiceMock) Start(
	phaseType types.PhaseType,
	trainID, deliveryPhaseID, verificationPhaseID, deployPhaseID uint64,
	branch, sha string,
	buildUser *types.User) error {

	if m.StartMock == nil {
		return nil
	}
	return m.StartMock(
		phaseType, trainID, deliveryPhaseID, verificationPhaseID, deployPhaseID,
		branch, sha, buildUser)
}
