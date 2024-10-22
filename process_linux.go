// Gets process environment by PID in Linux operating systems.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func dropNullByte(data []byte) []byte {
	size := len(data)
	if size > 0 && data[size-1] == '\000' {
		return data[0 : size-1]
	}
	return data
}

func SplitProcEnviron(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF {
		if len(data) == 0 {
			return 0, nil, nil
		} else {
			return len(data), dropNullByte(data), nil
		}
	}

	if data[0] == '\000' {
		return 1, nil, nil
	}

	if i := bytes.IndexByte(data, '\000'); i > 0 {
		return i + 1, dropNullByte(data[0:i]), nil
	}

	return 0, nil, nil
}

func ProcessEnvProvider(filter string) (map[string]string, error) {
	vars := map[string]string{}

	// TODO(zomglings): For now, we expect the filter to contain the process PID but we will later
	// want to add other things, like whitelists or blacklists of environment variables to read.
	pid, pidErr := strconv.ParseUint(filter, 10, 64)
	if pidErr != nil {
		return vars, pidErr
	}
	filename := fmt.Sprintf("/proc/%d/environ", pid)

	ifp, fileErr := os.Open(filename)
	if fileErr != nil {
		return vars, fileErr
	}
	defer ifp.Close()

	scanner := bufio.NewScanner(ifp)
	scanner.Split(SplitProcEnviron)
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
	helpString := "Provides the environment variables set for the process with the given pid."
	RegisterPlugin("proc", helpString, ProcessEnvProvider)
}
