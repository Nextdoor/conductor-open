package flags

import (
	"fmt"
	"os"
	"strconv"
)

func EnvString(name string, defaultValue string) string {
	value, present := os.LookupEnv(name)
	if !present {
		return defaultValue
	}
	return value
}

func RequiredEnvString(name string) string {
	value, present := os.LookupEnv(name)
	if !present {
		panic(fmt.Sprintf("Env string %s must be set", name))
	}
	return value
}

func EnvBool(name string, defaultValue bool) bool {
	value, present := os.LookupEnv(name)
	if !present {
		return defaultValue
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		panic(fmt.Sprintf(
			"Env bool %s must be /[0-1]|t(rue)?|f(alse)?/i, not %s", name, value))
	}
	return result
}

func RequiredEnvBool(name string) bool {
	value, present := os.LookupEnv(name)
	if !present {
		panic(fmt.Sprintf("Env bool %s must be set", name))
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		panic(fmt.Sprintf(
			"Env bool %s must be /[0-1]|t(rue)?|f(alse)?/i, not %s", name, value))
	}
	return result
}

func EnvInt(name string, defaultValue int) int {
	value, present := os.LookupEnv(name)
	if !present {
		return defaultValue
	}

	result, err := strconv.ParseInt(value, 10, 32)
	if err != nil || result < 0 {
		panic(fmt.Sprintf("Env int %s must be a valid positive integer, not %s", name, value))
	}
	return int(result)
}

func RequiredEnvInt(name string) int {
	value, present := os.LookupEnv(name)
	if !present {
		panic(fmt.Sprintf("Env int %s must be set", name))
	}

	result, err := strconv.ParseInt(value, 10, 32)
	if err != nil || result < 0 {
		panic(fmt.Sprintf("Env int %s must be a valid positive integer, not %s", name, value))
	}
	return int(result)
}
