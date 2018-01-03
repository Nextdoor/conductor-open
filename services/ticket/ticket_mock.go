package ticket

import "github.com/Nextdoor/conductor/shared/types"

type TicketServiceMock struct {
	CreateTicketsMock     func(*types.Train, []*types.Commit) ([]*types.Ticket, error)
	CloseTicketsMock      func([]*types.Ticket) error
	DeleteTicketsMock     func(*types.Train) error
	SyncTicketsMock       func(*types.Train) ([]*types.Ticket, []*types.Ticket, error)
	CloseTrainTicketsMock func(*types.Train) error
}

func (m *TicketServiceMock) CreateTickets(train *types.Train, commits []*types.Commit) ([]*types.Ticket, error) {
	if m.CreateTicketsMock == nil {
		return nil, nil
	}
	return m.CreateTicketsMock(train, commits)
}

func (m *TicketServiceMock) CloseTickets(tickets []*types.Ticket) error {
	if m.CloseTicketsMock == nil {
		return nil
	}
	return m.CloseTicketsMock(tickets)
}

func (m *TicketServiceMock) DeleteTickets(train *types.Train) error {
	if m.DeleteTicketsMock == nil {
		return nil
	}
	return m.DeleteTicketsMock(train)
}

func (m *TicketServiceMock) SyncTickets(train *types.Train) ([]*types.Ticket, []*types.Ticket, error) {
	if m.SyncTicketsMock == nil {
		return nil, nil, nil
	}
	return m.SyncTicketsMock(train)
}

func (m *TicketServiceMock) CloseTrainTickets(train *types.Train) error {
	if m.CloseTrainTicketsMock == nil {
		return nil
	}
	return m.CloseTrainTicketsMock(train)
}
