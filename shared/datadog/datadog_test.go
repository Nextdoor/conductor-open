package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	assert.NotNil(t, c)
	log("%s testing", "conductor")
}
