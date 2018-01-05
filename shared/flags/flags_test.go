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
	os.Setenv("test_env", "0")
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

func TestEnvInt(t *testing.T) {
	os.Setenv("test_env", "10")
	assert.Equal(t, 10, EnvInt("test_env", 5))

	os.Setenv("test_env", "50")
	assert.Equal(t, 50, EnvInt("test_env", 5))

	os.Unsetenv("test_env")
	assert.Equal(t, 5, EnvInt("test_env", 5))
}

func TestRequiredEnvIntPasses(t *testing.T) {
	os.Setenv("test_env", "0")
	assert.Equal(t, 0, RequiredEnvInt("test_env"))

	os.Setenv("test_env", "100")
	assert.Equal(t, 100, RequiredEnvInt("test_env"))
}

func TestRequiredEnvIntPanicsUnset(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "RequiredEnvInt did not panic.")
		}
	}()

	os.Unsetenv("test_env")
	RequiredEnvInt("test_env")
}

func TestRequiredEnvIntPanicsInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "RequiredEnvInt did not panic.")
		}
	}()

	os.Setenv("test_env", "value")
	RequiredEnvInt("test_env")
}
