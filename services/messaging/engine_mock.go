package messaging

import (
	"fmt"

	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/types"
)

type EngineMock struct {
	SendMock                        func(string)
	SendDirectMock                  func(string, string, string)
	FormatUserMock                  func(*types.User) string
	FormatNameEmailMock             func(string, string) string
	FormatNameEmailNotificationMock func(string, string) string
	FormatLinkMock                  func(string, string) string
	FormatBoldMock                  func(string) string
	FormatMonospacedMock            func(string) string
	Indent                          func(string) string
	Escape                          func(string) string
}

func (m *EngineMock) send(text string) {
	if m.SendMock != nil {
		m.SendMock(text)
	}
	logger.Info("%s", text)
}

func (m *EngineMock) sendDirect(name, email, text string) {
	if m.SendDirectMock != nil {
		m.SendDirectMock(name, email, text)
	}
	logger.Info("%s: %s", name, text)
}

func (m *EngineMock) formatUser(user *types.User) string {
	if m.FormatUserMock != nil {
		return m.FormatUserMock(user)
	}
	return user.Name
}

func (m *EngineMock) formatNameEmail(name, email string) string {
	if m.FormatNameEmailMock != nil {
		return m.FormatNameEmailMock(name, email)
	}
	return name
}

func (m *EngineMock) formatNameEmailNotification(name, email string) string {
	if m.FormatNameEmailNotificationMock != nil {
		return m.FormatNameEmailNotificationMock(name, email)
	}
	return name
}

func (m *EngineMock) formatLink(url, name string) string {
	if m.FormatLinkMock != nil {
		return m.FormatLinkMock(url, name)
	}
	return fmt.Sprintf("%s: %s", name, url)
}

func (m *EngineMock) formatBold(text string) string {
	if m.FormatBoldMock != nil {
		return m.FormatBoldMock(text)
	}
	return text
}

func (m *EngineMock) formatMonospaced(text string) string {
	if m.FormatMonospacedMock != nil {
		return m.FormatMonospacedMock(text)
	}
	return text
}

func (m *EngineMock) indent(text string) string {
	if m.Indent != nil {
		return m.Indent(text)
	}
	return fmt.Sprintf("  %s", text)
}

func (m *EngineMock) escape(text string) string {
	if m.Escape != nil {
		return m.Escape(text)
	}
	return text
}
