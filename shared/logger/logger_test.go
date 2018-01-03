package logger

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogUnstructured(t *testing.T) {
	var buffer bytes.Buffer
	output = &buffer

	log("level", "format %s", "string")

	output = os.Stdout

	assert.Equal(t, buffer.String(), "level: format string\n")
}

func TestLogStructured(t *testing.T) {
	var buffer bytes.Buffer
	output = &buffer
	useStructuredLogging = true

	log("another_level", "format %s", "string")

	output = os.Stdout
	useStructuredLogging = false

	assert.Equal(t, buffer.String(),
		"{\"service\":\"conductor\",\"level\":\"another_level\",\"message\":\"format string\"}\n")
}

func TestDebugLog(t *testing.T) {
	var buffer bytes.Buffer
	output = &buffer
	debugLoggingEnabled = true

	Debug("format %s", "string")

	output = os.Stdout
	debugLoggingEnabled = false

	assert.Equal(t, buffer.String(), "DEBUG: format string\n")

	buffer.Reset()
	output = &buffer

	Debug("format %s", "string")

	output = os.Stdout

	assert.Equal(t, buffer.String(), "")
}
