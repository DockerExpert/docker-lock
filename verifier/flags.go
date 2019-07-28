package verifier

import (
	"errors"
	"flag"
	"os"
)

type Flags struct {
	Outfile    string
	ConfigFile string
}

func NewFlags(cmdLineArgs []string) (*Flags, error) {
	var outfile string
	var configFile string
	command := flag.NewFlagSet("verify", flag.ExitOnError)
	command.StringVar(&outfile, "o", "docker-lock.json", "Path to save Lockfile from current directory.")
	command.StringVar(&configFile, "c", "", "Path to config file for auth credentials.")
	command.Parse(cmdLineArgs)
	if outfile == "" {
		return nil, errors.New("Outfile cannot be empty.")
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
	return &Flags{Outfile: outfile, ConfigFile: configFile}, nil
}
