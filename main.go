package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Metadata struct {
	SchemaVersion    string
	Vendor           string
	Version          string
	ShortDescription string
}

func PrintMetadata() {
	m := Metadata{
		SchemaVersion:    "0.1.0",
		Vendor:           "https://github.com/michaelperel/docker-lock",
		Version:          "v0.1.0",
		ShortDescription: "Generate and validate lock files for Docker",
	}
	var jsonData []byte
	jsonData, err := json.Marshal(m)
	if err != nil {
		panic("Malformed metadata")
	}
	fmt.Print(string(jsonData))

}

func main() {
	arg := os.Args[1]
	if arg == "docker-cli-plugin-metadata" {
		PrintMetadata()
        return
	}
    fmt.Println(os.Args)
}
