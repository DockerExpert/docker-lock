package generator

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return fmt.Sprintf("%v", []string(*s))
}

func (s *stringSliceFlag) Set(filePath string) error {
	*s = append(*s, filePath)
	return nil
}

type Flags struct {
	Dockerfiles []string
	Globs       []string
	Recursive   bool
	Lockfile    string
	ConfigFile  string
}

func NewFlags(cmdLineArgs []string) (*Flags, error) {
	var dockerfiles stringSliceFlag
	var globs stringSliceFlag
	var recursive bool
	var lockfile string
	var configFile string
	command := flag.NewFlagSet("generate", flag.ExitOnError)
	command.Var(&dockerfiles, "f", "Path to Dockerfile from current directory.")
	command.Var(&globs, "g", "Glob pattern to select Dockerfiles from current directory.")
	command.BoolVar(&recursive, "r", false, "recursively collect Dockerfiles from current directory.")
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
	return &Flags{Dockerfiles: []string(dockerfiles), Globs: []string(globs), Recursive: recursive, Lockfile: lockfile, ConfigFile: configFile}, nil
}
