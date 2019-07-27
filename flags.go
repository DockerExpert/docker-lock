package main

import (
	"flag"
	"os"
)

type flags struct {
	configFile string
}

func newFlags(cmdLineArgs []string) (*flags, error) {
	var configFile string
	command := flag.NewFlagSet("lock", flag.ExitOnError)
	command.StringVar(&configFile, "c", "", "Path to config file for auth credentials.")
	command.Parse(cmdLineArgs)
	if configFile != "" {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return nil, err
		}
	}
	if configFile == "" {
		defaultConfig := os.ExpandEnv("$HOME") + "/.docker/config.json"
		if _, err := os.Stat(defaultConfig); err == nil {
			configFile = defaultConfig
		}
	}
	return &flags{configFile: configFile}, nil
}
