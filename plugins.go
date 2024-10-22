package main

import (
	"fmt"
)

// A CheckenvProvider is a function mapping an input string to a map[string]string representing environment
// variables and their values.
// The input string represents a filter specific to the provider in question. Different providers may
// implement filters using different syntaxes. Each provider has to provide information about its filtering
// syntax when it registers itself.
// It can also return an error if there was any kind of issue retrieving the requested environment variables.
type CheckenvProvider func(string) (map[string]string, error)

// CheckenvPlugin represents all the metadata that checkenv requires about a registered plugin:
// 1. A Help string to display to users explaining the plugins filter syntax, how it is initialized, and possibly more.
// 2. An Init function which initializes the plugin when checkenv starts up.
// 3. The Provider function responsible for providing environment configuration from the source that the plugin represents.
type CheckenvPlugin struct {
	Help     string
	Provider CheckenvProvider
}

// RegisteredPlugins maps plugin names to the actual plugins. checkenv uses this map at runtime to
// understand what environments a user wants to check/compare.
var RegisteredPlugins map[string]CheckenvPlugin = make(map[string]CheckenvPlugin)

// RegisterPlugin is the function that each plugin must call to provide its functionality to checkenv users.
func RegisterPlugin(name string, help string, provider CheckenvProvider) {
	if _, ok := RegisteredPlugins[name]; ok {
		panic(fmt.Sprintf("A plugin already exists with name: %s", name))
	}
	plugin := CheckenvPlugin{
		Help:     help,
		Provider: provider,
	}
	RegisteredPlugins[name] = plugin
}
