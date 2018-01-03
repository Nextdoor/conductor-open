/* Slack messaging implementation. */
package messaging

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nlopes/slack"

	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	slackToken   = flags.EnvString("SLACK_TOKEN", "")
	slackChannel = flags.EnvString("SLACK_CHANNEL", "conductor")
)

var (
	slackEmailUserCache         map[string]*slack.User
	slackEmailUserCacheUnixTime int64
)

const SLACK_CACHE_TTL = 60

type slackEngine struct {
	api *slack.Client
}

func newSlackEngine() *Messenger {
	if slackToken == "" {
		panic(errors.New("slack_token flag must be set."))
	}
	api := slack.New(slackToken)
	return &Messenger{
		Engine: &slackEngine{api},
	}
}

func (e *slackEngine) send(text string) {
	e.sendToSlack(slackChannel, text)
}

func (e *slackEngine) sendDirect(name, email, text string) {
	slackUser, err := e.emailToSlackUser(email)
	if err != nil {
		logger.Error("Error looking up slack user for email %s", email)
		return
	}
	e.sendToSlack(fmt.Sprintf("@%s", slackUser.Name), text)
}

func (e *slackEngine) sendToSlack(destination, text string) {
	logger.Info("%s", text)
	params := slack.NewPostMessageParameters()
	params.AsUser = true
	params.EscapeText = false
	_, _, err := e.api.PostMessage(destination, text, params)
	if err != nil {
		logger.Error("%v", err)
	}
}

func (e *slackEngine) formatUser(user *types.User) string {
	return e.formatNameEmailNotification(user.Name, user.Email)
}

func (e *slackEngine) formatNameEmail(name, email string) string {
	return name
}

func (e *slackEngine) formatNameEmailNotification(name, email string) string {
	slackUser, err := e.emailToSlackUser(email)
	if err != nil {
		logger.Error("Error looking up slack user for email %s", email)
		return name
	} else {
		return fmt.Sprintf("<@%s|%s>", slackUser.ID, slackUser.Name)
	}
}

func (e *slackEngine) formatLink(url, text string) string {
	return fmt.Sprintf("<%s|%s>", url, text)
}

func (e *slackEngine) formatBold(text string) string {
	return fmt.Sprintf("*%s*", text)
}

func (e *slackEngine) formatMonospaced(text string) string {
	return fmt.Sprintf("`%s`", text)
}

func (e *slackEngine) indent(text string) string {
	return fmt.Sprintf("> %s", text)
}

func (e *slackEngine) escape(text string) string {
	text = strings.Replace(text, "&", "&amp;", -1)
	text = strings.Replace(text, "<", "&lt;", -1)
	text = strings.Replace(text, ">", "&gt;", -1)
	return text
}

func (m *slackEngine) cacheSlackUsers() (map[string]*slack.User, error) {
	// We maintain a cache of email address to Slack user.
	// This is required to map a commit author to a Slack user we can @-mention.
	// Since Slack users can change their handles, we only keep the cache for SLACK_CACHE_TTL seconds.
	now := time.Now()

	if slackEmailUserCacheUnixTime == 0 || now.Unix()-slackEmailUserCacheUnixTime > SLACK_CACHE_TTL {
		slackEmailUserCache = make(map[string]*slack.User, 200)

		users, err := m.api.GetUsers()
		if err != nil {
			logger.Error("Could not fetch Slack users list: %v", err)
			if slackEmailUserCacheUnixTime == 0 {
				// No cache - propagate error up.
				return nil, err
			} else {
				// There was an error refreshing the cache, but we can use the existing cache transparently.
				return slackEmailUserCache, nil
			}
		}

		slackEmailUserCacheUnixTime = now.Unix()

		for i := range users {
			user := users[i]
			if len(user.Profile.Email) > 0 {
				slackEmailUserCache[user.Profile.Email] = &user
			}
		}
	}

	return slackEmailUserCache, nil
}

func (m *slackEngine) emailToSlackUser(email string) (*slack.User, error) {
	users, err := m.cacheSlackUsers()
	if err != nil {
		return nil, err
	}

	if slackUser, ok := users[email]; ok {
		return slackUser, nil
	}

	return nil, fmt.Errorf("Could not find email %s in slack users list", email)
}
