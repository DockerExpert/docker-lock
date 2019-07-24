package main

import (
	"fmt"
	"github.com/michaelperel/docker-lock/lock"
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
		generator := lock.Generator{*options}
		generator.GenerateLockfile()
	case "verify":
		options := lock.NewOptions(subCommand, os.Args[3:])
		generator := lock.Generator{*options}
		verifier := lock.Verifier{generator}
		equal, reason := verifier.VerifyLockfile()
		if !equal {
			fmt.Fprintln(os.Stderr, reason)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "Expected 'generate' or 'verify' subcommands.")
		os.Exit(1)
	}
}
