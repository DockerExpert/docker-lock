package main

import (
	"errors"
	"fmt"
	"github.com/michaelperel/docker-lock/lock"
	"github.com/michaelperel/docker-lock/registry"
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
		options, err := lock.NewOptions(subCommand, os.Args[3:])
		handleError(err)
		generator := lock.Generator{Options: *options}
		wrapper := new(registry.DockerWrapper)
		err = generator.GenerateLockfile(wrapper)
		handleError(err)
	case "verify":
		options, err := lock.NewOptions(subCommand, os.Args[3:])
		generator := lock.Generator{Options: *options}
		verifier := lock.Verifier{Generator: generator}
		wrapper := new(registry.DockerWrapper)
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
