package main

import (
	"fmt"
)

func main() {
	for name, plugin := range RegisteredPlugins {
		pluginInitErr := plugin.Init()
		pluginEnv, pluginErr := plugin.Provider("")
		fmt.Printf("Plugin: %s\n", name)
		fmt.Printf("Help: %s\n", plugin.Help)
		if pluginInitErr != nil {
			fmt.Printf("Init error: %s\n", pluginInitErr.Error())
		}
		if pluginErr != nil {
			fmt.Printf("Plugin error: %s\n", pluginErr.Error())
		} else {
			fmt.Printf("Env:\n%v\n", pluginEnv)
		}
	}
}
