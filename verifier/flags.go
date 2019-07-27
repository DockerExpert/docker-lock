package verifier

import (
	"errors"
	"flag"
	"os"
)

type Flags struct {
	Lockfile   string
	ConfigFile string
}

func NewFlags(cmdLineArgs []string) (*Flags, error) {
	var lockfile string
	var configFile string
	command := flag.NewFlagSet("verify", flag.ExitOnError)
	command.StringVar(&lockfile, "o", "docker-lock.json", "Path to Lockfile from current directory.")
	command.StringVar(&configFile, "c", "", "Path to config file for auth credentials.")
	command.Parse(cmdLineArgs)
	if lockfile == "" {
		return nil, errors.New("Lockfile cannot be empty.")
	}
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
	return &Flags{Lockfile: lockfile, ConfigFile: configFile}, nil
}
