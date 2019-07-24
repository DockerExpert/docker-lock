package lock

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type Options struct {
	Dockerfiles []string
	Recursive   bool
	Lockfile    string
}

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return fmt.Sprintf("%v", []string(*s))
}

func (s *stringSliceFlag) Set(filePath string) error {
	*s = append(*s, filePath)
	return nil
}

func NewOptions(subCommand string, args []string) *Options {
	var dockerfiles stringSliceFlag
	var recursive bool
	var lockfile string
	command := flag.NewFlagSet(subCommand, flag.ExitOnError)
	command.Var(&dockerfiles, "f", "Path to Dockerfile from current directory.")
	command.BoolVar(&recursive, "r", false, "Recursively collect Dockerfiles from current directory.")
	command.StringVar(&lockfile, "o", "docker-lock.json", "Path to Lockfile from current directory.")
	command.Parse(args)
	options := Options{Dockerfiles: []string(dockerfiles), Recursive: recursive, Lockfile: lockfile}
	err := options.validate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return &options
}

func (options *Options) validate() error {
	if options.Recursive && len(options.Dockerfiles) > 0 {
		return errors.New("Cannot specify both -r and -f.")
	}
	return nil
}
