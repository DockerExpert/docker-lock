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

func parseFlags(subCommand string, args []string) ([]string, string, error) {
	var dockerfileFlag stringSliceFlag
	var globFlag stringSliceFlag
	var recursiveFlag bool
	var lockfileFlag string

	command := flag.NewFlagSet(subCommand, flag.ExitOnError)
	command.Var(&dockerfileFlag, "f", "Path to Dockerfile from current directory.")
	command.Var(&globFlag, "g", "Glob pattern (surrounded in quotes) to select Dockerfiles from current directory.")
	command.BoolVar(&recursiveFlag, "r", false, "recursively collect Dockerfiles from current directory.")
	command.StringVar(&lockfileFlag, "o", "docker-lock.json", "Path to Lockfile from current directory.")
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
			return nil, "", err
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
	return dockerfiles, lockfileFlag, nil
}
