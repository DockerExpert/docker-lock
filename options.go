package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return fmt.Sprintf("%v", []string(*s))
}

func (s *stringSliceFlag) Set(filePath string) error {
	*s = append(*s, filePath)
	return nil
}

// TODO: make return value an object once all options figured out
func parseFlags(subCommand string, args []string) ([]string, string, string, error) {
	var dockerfileFlag stringSliceFlag
	var globFlag stringSliceFlag
	var recursiveFlag bool
	var lockfileFlag string
	var configFileFlag string

	command := flag.NewFlagSet(subCommand, flag.ExitOnError)
	command.Var(&dockerfileFlag, "f", "Path to Dockerfile from current directory.")
	command.Var(&globFlag, "g", "Glob pattern (surrounded in quotes) to select Dockerfiles from current directory.")
	command.BoolVar(&recursiveFlag, "r", false, "recursively collect Dockerfiles from current directory.")
	command.StringVar(&lockfileFlag, "o", "docker-lock.json", "Path to Lockfile from current directory.")
	command.StringVar(&configFileFlag, "c", "", "Path to config file for auth credentials.")
	command.Parse(args)

	dockerfileSet := make(map[string]bool)
	for _, dockerfile := range dockerfileFlag {
		dockerfileSet[dockerfile] = true
	}
	if recursiveFlag {
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Base(path) == "Dockerfile" {
				dockerfileSet[path] = true
			}
			return nil
		})
	}
	for _, pattern := range globFlag {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, "", "", err
		}
		for _, match := range matches {
			dockerfileSet[match] = true
		}
	}
	dockerfiles := make([]string, 0, len(dockerfileSet))
	for dockerfile := range dockerfileSet {
		dockerfiles = append(dockerfiles, dockerfile)
	}
	if len(dockerfiles) == 0 {
		dockerfiles = []string{"Dockerfile"}
	}
	if configFileFlag != "" {
		if _, err := os.Stat(configFileFlag); os.IsNotExist(err) {
			return nil, "", "", err
		}
	}
	if configFileFlag == "" {
		defaultConfig := os.ExpandEnv("$HOME") + "/.docker/config.json"
		if _, err := os.Stat(defaultConfig); err == nil {
			configFileFlag = defaultConfig
		}
	}
	return dockerfiles, lockfileFlag, configFileFlag, nil
}
