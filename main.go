package main

import (
	"fmt"
	"github.com/michaelperel/docker-lock/generate"
	"github.com/michaelperel/docker-lock/metadata"
    "github.com/michaelperel/docker-lock/options"
	"os"
)


func main() {
    // Boilerplate required by every cli-plugin to show up in the 'docker' command.
	arg := os.Args[1]
	if arg == "docker-cli-plugin-metadata" {
		fmt.Println(metadata.JSONMetadata())
		os.Exit(0)
	}

	if len(os.Args) <= 2 {
		fmt.Fprintln(os.Stderr, "Expected 'generate' or 'verify' subcommands.")
		os.Exit(1)
	}
    subCommand := os.Args[2]
    options := options.Parse(subCommand, os.Args[3:])
	switch subCommand {
	case "generate":
		generate.LockFile(options)
	case "verify":
		// verify.LockFile(options)
	default:
		fmt.Fprintln(os.Stderr, "Expected 'generate' or 'verify' subcommands.")
		os.Exit(1)
	}
}
