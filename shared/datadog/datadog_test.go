package datadog

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitstatsdHostValid(t *testing.T) {
	os.Setenv("STATSD_HOST", "localhost")
	Initstatsd()
	assert.NotNil(t, C)
}

func TestInitstatsdHostInvalid(t *testing.T) {
	os.Setenv("STATSD_HOST", "localhost-invalid")
	Initstatsd()
	assert.Nil(t, C)
}

func TestLog(t *testing.T) {
}
