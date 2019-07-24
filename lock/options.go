package lock

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	options.setDefaults()
	return &options
}

func (o *Options) validate() error {
	if o.Recursive && len(o.Dockerfiles) > 0 {
		return errors.New("Cannot specify both -r and -f.")
	}
	return nil
}

func (o *Options) setDefaults() {
	if len(o.Dockerfiles) != 0 {
		return
	}
	if o.Recursive {
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Base(path) == "Dockerfile" {
				o.Dockerfiles = append(o.Dockerfiles, path)
			}
			return nil
		})
		return
	}
	o.Dockerfiles = []string{"Dockerfile"}
}
