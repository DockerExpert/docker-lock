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

	var subCommand string
	subCommandIndex := -1
	for i, arg := range os.Args {
		if arg == "generate" {
			subCommand = "generate"
			subCommandIndex = i
			break
		}
		if arg == "verify" {
			subCommand = "verify"
			subCommandIndex = i
			break
		}
	}
	if subCommandIndex == -1 {
		handleError(errors.New("Expected 'generate' or 'verify' subcommands."))
	}

	// cli form = [binary, lock, lockargs, subcmd, subcmdargs...]
	lockIndex := 2
	flags, err := newFlags(os.Args[lockIndex:subCommandIndex])
	handleError(err)
	switch subCommand {
	case "generate":
		generator, err := generator.New(os.Args[subCommandIndex+1:])
		handleError(err)
		wrapper := &registry.DockerWrapper{ConfigFile: flags.configFile}
		err = generator.GenerateLockfile(wrapper)
		handleError(err)
	case "verify":
		verifier, err := verifier.New(os.Args[subCommandIndex+1:])
		handleError(err)
		wrapper := &registry.DockerWrapper{ConfigFile: flags.configFile}
		err = verifier.VerifyLockfile(wrapper)
		handleError(err)
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
