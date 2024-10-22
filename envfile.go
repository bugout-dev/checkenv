package main

import (
	"bufio"
	"os"
	"strings"
)

func EnvFileProvider(filter string) (map[string]string, error) {
	// TODO(zomglings): For now, we expect the filter to contain the name of the env file. However,
	// we will later want to add other things, like prefixes that we should strip off of lines (e.g. "export" or "set")
	// and prefixes that we should ignore (e.g. "#" or "//")
	vars := map[string]string{}
	filename := filter
	ifp, fileErr := os.Open(filename)
	if fileErr != nil {
		return vars, fileErr
	}
	defer ifp.Close()

	scanner := bufio.NewScanner(ifp)
	for scanner.Scan() {
		line := scanner.Text()
		components := strings.Split(line, "=")
		name := components[0]
		var value string
		if len(components) > 1 {
			value = strings.Join(components[1:], "=")
		}
		vars[name] = value
	}

	scanErr := scanner.Err()
	return vars, scanErr
}

func init() {
	helpString := "Provides the environment variables defined in the env file with the given path."
	RegisterPlugin("file", helpString, EnvFileProvider)
}
