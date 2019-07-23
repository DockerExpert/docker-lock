package main

import (
	"fmt"
	"github.com/michaelperel/docker-lock/generate"
	"github.com/michaelperel/docker-lock/metadata"
	"os"
)

func main() {
	arg := os.Args[1]
	if arg == "docker-cli-plugin-metadata" {
		fmt.Println(metadata.JSONMetadata())
		os.Exit(0)
	}
	if len(os.Args) <= 2 {
		fmt.Fprintln(os.Stderr, "Expected 'generate' or 'verify' subcommands.\n")
		os.Exit(1)
	}
	switch subCmd := os.Args[2]; subCmd {
	case "generate":
		generate.LockFile()
	case "verify":
		// verify.LockFile()
	default:
		fmt.Fprintln(os.Stderr, "Expected 'generate' or 'verify' subcommands.\n")
		os.Exit(1)
	}
}
