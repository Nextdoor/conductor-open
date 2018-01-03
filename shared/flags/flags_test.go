package flags

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvString(t *testing.T) {
	os.Setenv("test_env", "value")
	assert.Equal(t, "value", EnvString("test_env", "default"))

	os.Unsetenv("test_env")
	assert.Equal(t, "default", EnvString("test_env", "default"))
}

func TestRequiredEnvStringSuccess(t *testing.T) {
	os.Setenv("test_env", "value")
	assert.Equal(t, "value", RequiredEnvString("test_env"))
}

func TestRequiredEnvStringPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "RequiredEnvString did not panic.")
		}
	}()

	os.Unsetenv("test_env")
	RequiredEnvString("test_env")
}

func TestEnvBool(t *testing.T) {
	os.Setenv("test_env", "FaLSE")
	assert.Equal(t, false, EnvBool("test_env", true))

	os.Setenv("test_env", "true")
	assert.Equal(t, true, EnvBool("test_env", true))

	os.Unsetenv("test_env")
	assert.Equal(t, true, EnvBool("test_env", true))
}

func TestRequiredEnvBoolPasses(t *testing.T) {
	os.Setenv("test_env", "false")
	assert.Equal(t, false, RequiredEnvBool("test_env"))

	os.Setenv("test_env", "True")
	assert.Equal(t, true, RequiredEnvBool("test_env"))
}

func TestRequiredEnvBoolPanicsUnset(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "RequiredEnvBool did not panic.")
		}
	}()

	os.Unsetenv("test_env")
	RequiredEnvBool("test_env")
}

func TestRequiredEnvBoolPanicsInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "RequiredEnvBool did not panic.")
		}
	}()

	os.Setenv("test_env", "value")
	RequiredEnvBool("test_env")
}
