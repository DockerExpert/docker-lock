package main

import (
	"fmt"
	"github.com/michaelperel/docker-lock/lock"
	"github.com/michaelperel/docker-lock/registry"
	"os"
)

func main() {
	// Boilerplate required by every cli-plugin to show up in the 'docker' command.
	arg := os.Args[1]
	if arg == "docker-cli-plugin-metadata" {
		fmt.Println(getMetadata())
		os.Exit(0)
	}

	if len(os.Args) <= 2 {
		fmt.Fprintln(os.Stderr, "Expected 'generate' or 'verify' subcommands.")
		os.Exit(1)
	}
	subCommand := os.Args[2]
	switch subCommand {
	case "generate":
		options := lock.NewOptions(subCommand, os.Args[3:])
		generator := lock.Generator{Options: *options}
		wrapper := new(registry.DockerWrapper)
		generator.GenerateLockfile(wrapper)
	case "verify":
		options := lock.NewOptions(subCommand, os.Args[3:])
		generator := lock.Generator{Options: *options}
		verifier := lock.Verifier{generator}
		wrapper := new(registry.DockerWrapper)
		equal, reason := verifier.VerifyLockfile(wrapper)
		if !equal {
			fmt.Fprintln(os.Stderr, reason)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "Expected 'generate' or 'verify' subcommands.")
		os.Exit(1)
	}
}
