package main

import (
	"errors"
	"fmt"
	"github.com/michaelperel/docker-lock/generator"
	"github.com/michaelperel/docker-lock/registry"
	"github.com/michaelperel/docker-lock/verifier"
	"os"
)

func main() {
	// Boilerplate required by every cli-plugin to show up in the 'docker' command.
	if os.Args[1] == "docker-cli-plugin-metadata" {
		metadata, err := getMetadata()
		handleError(err)
		fmt.Println(metadata)
		os.Exit(0)
	}
	if len(os.Args) <= 2 {
		handleError(errors.New("Expected 'generate' or 'verify' subcommands."))
	}
	switch subCommand := os.Args[2]; subCommand {
	case "generate":
		dockerfiles, lockfile, configFile, err := parseFlags(subCommand, os.Args[3:])
		handleError(err)
		generator, err := generator.New(dockerfiles, lockfile)
		handleError(err)
		wrapper := &registry.DockerWrapper{ConfigFile: configFile}
		err = generator.GenerateLockfile(wrapper)
		handleError(err)
	case "verify":
		dockerfiles, lockfile, configFile, err := parseFlags(subCommand, os.Args[3:])
		handleError(err)
		generator, err := generator.New(dockerfiles, lockfile)
		handleError(err)
		verifier, err := verifier.New(generator)
		handleError(err)
		wrapper := &registry.DockerWrapper{ConfigFile: configFile}
		err = verifier.VerifyLockfile(wrapper)
		handleError(err)
	default:
		handleError(errors.New("Expected 'generate' or 'verify' subcommands."))
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
