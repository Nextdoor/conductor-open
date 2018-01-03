package ticket

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Nextdoor/conductor/shared/types"
)

const (
	authorName1 = "A Developer"
	email1      = "test@gmail.com"
	email2      = "test2@example.com"
	message1    = "commit number 1"
	message2    = "a second commit"
	message3    = "the third commit message\nanother line"
	sha1        = "0"
	sha2        = "1"
	sha3        = "2"
)

func TestDescriptionFromCommits(t *testing.T) {
	testCommits := []*types.Commit{
		{AuthorEmail: email1, Message: message1, AuthorName: authorName1, SHA: sha1},
		{AuthorEmail: email1, Message: message2, AuthorName: authorName1, SHA: sha2},
		{AuthorEmail: email2, Message: message3, AuthorName: authorName1, SHA: sha3}}
	desc, err := descriptionFromCommits(testCommits)
	assert.NoError(t, err)
	assert.Contains(t, desc, message1)
	assert.Contains(t, desc, message2)
	// Contains the first line of the commit message but not the second
	assert.Contains(t, desc, strings.Split(message3, "\n")[0])
	assert.NotContains(t, desc, strings.Split(message3, "\n")[1])
	assert.Contains(t, desc, authorName1)
}
