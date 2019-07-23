package options

import (
    "os"
    "flag"
    "fmt"
    "errors"
)

type Options struct {
    Dockerfiles []string
    Recursive bool
}

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return fmt.Sprintf("%v", []string(*s))
}

func (s *stringSliceFlag) Set(filePath string) error {
	*s = append(*s, filePath)
	return nil
}

func Parse(subCommand string, args []string) Options {
    var dockerfiles stringSliceFlag
    var recursive bool
    command := flag.NewFlagSet(subCommand, flag.ExitOnError)
    command.Var(&dockerfiles, "f", "Path to Dockerfile from current directory.")
	command.BoolVar(&recursive, "r", false, "Recursively collect Dockerfiles from current directory.")
	command.Parse(args)
    options := Options{Dockerfiles: []string(dockerfiles), Recursive: recursive}
    err := options.validate()
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
    }
    return options
}

func (options Options) validate() error {
    if options.Recursive && len(options.Dockerfiles) > 0 {
        return errors.New("Cannot specify both -r and -f.")
    }
    return nil
}
