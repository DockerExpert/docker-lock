package generator

import (
	"errors"
	"flag"
	"fmt"
)

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return fmt.Sprintf("%v", []string(*s))
}

func (s *stringSliceFlag) Set(filePath string) error {
	*s = append(*s, filePath)
	return nil
}

type flags struct {
	dockerfiles []string
	globs       []string
	recursive   bool
	lockfile    string
}

func newFlags(cmdLineArgs []string) (*flags, error) {
	var dockerfiles stringSliceFlag
	var globs stringSliceFlag
	var recursive bool
	var lockfile string
	command := flag.NewFlagSet("generate", flag.ExitOnError)
	command.Var(&dockerfiles, "f", "Path to Dockerfile from current directory.")
	command.Var(&globs, "g", "Glob pattern to select Dockerfiles from current directory.")
	command.BoolVar(&recursive, "r", false, "recursively collect Dockerfiles from current directory.")
	command.StringVar(&lockfile, "o", "docker-lock.json", "Path to Lockfile from current directory.")
	command.Parse(cmdLineArgs)
	if lockfile == "" {
		return nil, errors.New("Lockfile cannot be empty.")
	}
	return &flags{dockerfiles: []string(dockerfiles), globs: []string(globs), recursive: recursive, lockfile: lockfile}, nil
}
