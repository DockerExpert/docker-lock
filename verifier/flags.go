package verifier

import (
	"errors"
	"flag"
)

type flags struct {
	lockfile string
}

func newFlags(cmdLineArgs []string) (*flags, error) {
	var lockfile string
	command := flag.NewFlagSet("verify", flag.ExitOnError)
	command.StringVar(&lockfile, "o", "docker-lock.json", "Path to Lockfile from current directory.")
	command.Parse(cmdLineArgs)
	if lockfile == "" {
		return nil, errors.New("Lockfile cannot be empty.")
	}
	return &flags{lockfile: lockfile}, nil
}
