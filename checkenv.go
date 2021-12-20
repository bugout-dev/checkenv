package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type showSpec struct {
	loadFrom      map[string]interface{}
	providersFull map[string]interface{}
	providerVars  map[string]map[string]interface{}
}

func parseShowSpec(args []string) *showSpec {
	spec := showSpec{loadFrom: map[string]interface{}{}, providersFull: map[string]interface{}{}, providerVars: map[string]map[string]interface{}{}}
	for _, arg := range args {
		components := strings.Split(arg, "://")
		// TODO(zomglings): Can environment variable names contain the characters "://"? I don't think so.
		// However, it is possible that the provider filter arguments *could* contain those characters. That
		// means that this logic is wrong. We should probably use a different separator.
		providerSpec := components[0]
		spec.loadFrom[providerSpec] = nil
		if len(components) == 1 {
			spec.providersFull[providerSpec] = nil
		} else {
			if _, ok := spec.providerVars[providerSpec]; !ok {
				spec.providerVars[providerSpec] = map[string]interface{}{}
			}
			varSpec := strings.Join(components[1:], "://")
			varNames := strings.Split(varSpec, ",")
			for _, varName := range varNames {
				spec.providerVars[providerSpec][varName] = nil
			}
		}
	}

	return &spec
}

func VariablesFromProviderSpec(providerSpec string) (map[string]string, error) {
	components := strings.Split(providerSpec, "+")
	provider := components[0]
	var providerArgs string
	if len(components) > 1 {
		providerArgs = strings.Join(components[1:], "+")
	}
	plugin, pluginExists := RegisteredPlugins[provider]
	if !pluginExists {
		return map[string]string{}, fmt.Errorf("unregistered provider: %s", provider)
	}
	return plugin.Provider(providerArgs)
}

func main() {
	pluginsCommand := "plugins"
	pluginsFlags := flag.NewFlagSet("plugins", flag.ExitOnError)
	pluginsHelp := pluginsFlags.Bool("h", false, "Use this flag if you want help with this command")
	pluginsFlags.BoolVar(pluginsHelp, "help", false, "Use this flag if you want help with this command")

	showCommand := "show"
	showFlags := flag.NewFlagSet("show", flag.ExitOnError)
	showHelp := showFlags.Bool("h", false, "Use this flag if you want help with this command")
	showFlags.BoolVar(showHelp, "help", false, "Use this flag if you want help with this command")
	showExport := showFlags.Bool("export", false, "Use this flag to prepend and \"export \" before every environment variable definition")

	availableCommands := fmt.Sprintf("%s,%s", pluginsCommand, showCommand)

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Please use one of the subcommands: %s\n", availableCommands)
		os.Exit(2)
	}

	command := os.Args[1]

	switch command {
	case pluginsCommand:
		pluginsFlags.Parse(os.Args[2:])
		if *pluginsHelp {
			fmt.Fprintf(os.Stderr, "Usage: %s %s\nTakes no arguments.\nLists available plugins with a brief description of each one.\n", os.Args[0], os.Args[1])
			os.Exit(2)
		}
		fmt.Println("Available plugins:")
		for name, plugin := range RegisteredPlugins {
			fmt.Printf("%s\n\t%s\n", name, plugin.Help)
		}
	case showCommand:
		showFlags.Parse(os.Args[2:])
		if *showHelp || showFlags.NArg() == 0 {
			fmt.Fprintf(os.Stderr, "Usage: %s %s [<provider_name>[+<provider_args>] ...] [<provider_name>[+<provider_args>]://<var_name_1>,<var_name_2>,...,<var_name_n> ...]\nShows the environment variables defined by the given providers.\n", os.Args[0], os.Args[1])
			os.Exit(2)
		}
		spec := parseShowSpec(showFlags.Args())
		providedVars := make(map[string]map[string]string)
		for providerSpec := range spec.loadFrom {
			vars, providerErr := VariablesFromProviderSpec(providerSpec)
			if providerErr != nil {
				log.Fatalf(providerErr.Error())
			}
			providedVars[providerSpec] = vars
		}

		exportPrefix := ""
		if *showExport {
			exportPrefix = "export "
		}

		for providerSpec := range spec.providersFull {
			fmt.Printf("# Generated with %s - all variables:\n", providerSpec)
			for k, v := range providedVars[providerSpec] {
				fmt.Printf("%s%s=%s\n", exportPrefix, k, v)
			}
		}
		for providerSpec, queriedVars := range spec.providerVars {
			fmt.Printf("# Generated with %s - specific variables:\n", providerSpec)
			definedVars := providedVars[providerSpec]
			for k := range queriedVars {
				v, ok := definedVars[k]
				if !ok {
					fmt.Printf("# UNDEFINED: %s\n", k)
				} else {
					fmt.Printf("%s%s=%s\n", exportPrefix, k, v)
				}
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s. Please use one of the subcommands: %s.\n", command, availableCommands)
	}
}
