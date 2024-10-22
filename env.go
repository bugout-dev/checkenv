// checkenv plugin that provides the environment variables defined in the checkenv process.

package main

import (
	"os"
	"strings"
)

func CurrentEnvProvider(filter string) (map[string]string, error) {
	rawEnvs := os.Environ()
	environment := make(map[string]string)

	for _, envspec := range rawEnvs {
		components := strings.Split(envspec, "=")
		key := components[0]
		value := strings.Join(components[1:], "=")
		environment[key] = value
	}

	return environment, nil
}

func init() {
	helpString := "Provides the environment variables defined in the checkenv process."
	RegisterPlugin("env", helpString, CurrentEnvProvider)
}
