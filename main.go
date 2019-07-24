package main

import (
	"fmt"
	"github.com/michaelperel/docker-lock/lock"
	"github.com/michaelperel/docker-lock/options"
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
		options := options.Parse(subCommand, os.Args[3:])
		lock.GenerateLockfile(options)
	case "verify":
		options := options.Parse(subCommand, os.Args[3:])
		equal, reason := lock.VerifyLockfile(options)
		if !equal {
			fmt.Fprintln(os.Stderr, reason)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "Expected 'generate' or 'verify' subcommands.")
		os.Exit(1)
	}
}
