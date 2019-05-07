// +build messaging

package messaging

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEmailToSlackUser(t *testing.T) {
	slackService := newSlackEngine()
	slackEngine := slackService.Engine.(*slackEngine)

	now := time.Now()
	// Test the cache TTL. We should get back the currently cached version, which is an empty map.
	slackEmailUserCacheUnixTime = now.Unix()
	users, err := slackEngine.cacheSlackUsers()
	assert.NoError(t, err)
	assert.Empty(t, users)

	// Reset the cache timestamp to force a lookup.
	slackEmailUserCacheUnixTime = 0

	users, err = slackEngine.cacheSlackUsers()
	assert.NoError(t, err)
	assert.NotEmpty(t, users)

	// Pick first email in the cache
	for email, _ := range users {
		user, err := slackEngine.emailToSlackUser(email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		return
	}
}
