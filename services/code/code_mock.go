package code

import (
	"net/http"

	"github.com/Nextdoor/conductor/shared/types"
)

type CodeServiceMock struct {
	CommitsOnBranchMock       func(string, int) ([]*types.Commit, error)
	CommitsOnBranchAfterMock  func(string, string) ([]*types.Commit, error)
	CompareRefsMock           func(string, string) ([]*types.Commit, error)
	RevertMock                func(sha1, branch string) error
	ParseWebhookForBranchMock func(r *http.Request) (string, error)
}

func (m *CodeServiceMock) CommitsOnBranch(branch string, max int) ([]*types.Commit, error) {
	if m.CommitsOnBranchMock == nil {
		return nil, nil
	}
	return m.CommitsOnBranchMock(branch, max)
}

func (m *CodeServiceMock) CommitsOnBranchAfter(branch string, sha string) ([]*types.Commit, error) {
	if m.CommitsOnBranchAfterMock == nil {
		return nil, nil
	}
	return m.CommitsOnBranchAfterMock(branch, sha)
}

func (m *CodeServiceMock) CompareRefs(oldRef, newRef string) ([]*types.Commit, error) {
	if m.CompareRefsMock == nil {
		return nil, nil
	}
	return m.CompareRefsMock(oldRef, newRef)
}

func (m *CodeServiceMock) Revert(sha1, branch string) error {
	if m.RevertMock == nil {
		return nil
	}
	return m.RevertMock(sha1, branch)
}

func (m *CodeServiceMock) ParseWebhookForBranch(r *http.Request) (string, error) {
	if m.ParseWebhookForBranchMock == nil {
		return "", nil
	}
	return m.ParseWebhookForBranchMock(r)
}
