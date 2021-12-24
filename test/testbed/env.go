package testbed

import (
	"os"
)

func FromEnv() *Testbed {
	result := &Testbed{
		RelayServerUrl: "http://localhost:8080",
		Username:       "smoketest-user",
		Password:       "smoketest-password",
	}

	if val, ok := os.LookupEnv("RELAY_SMOKETEST_SERVER_URL"); ok {
		result.RelayServerUrl = val
	}

	if val, ok := os.LookupEnv("RELAY_SMOKETEST_USERNAME"); ok {
		result.Username = val
	}

	if val, ok := os.LookupEnv("RELAY_SMOKETEST_PASSWORD"); ok {
		result.Password = val
	}

	return result
}
