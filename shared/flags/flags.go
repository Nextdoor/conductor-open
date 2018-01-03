package flags

import (
	"fmt"
	"os"
	"strings"
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

	result, valid := parseBoolString(value)
	if !valid {
		panic(fmt.Sprintf("Env bool %s can be either true or false, not %s", name, value))
	}
	return result
}

func RequiredEnvBool(name string) bool {
	value, present := os.LookupEnv(name)
	if !present {
		panic(fmt.Sprintf("Env bool %s must be set", name))
	}

	result, valid := parseBoolString(value)
	if !valid {
		panic(fmt.Sprintf("Env bool %s must be either true or false, not %s", name, value))
	}
	return result
}

// Returns two bools: the parsed bool value, and whether the string was a valid bool.
// Can be "true" or "false", case insensitive.
func parseBoolString(value string) (bool, bool) {
	value = strings.ToLower(value)
	if value == "true" {
		return true, true
	} else if value == "false" {
		return false, true
	}
	return false, false
}
