package metadata

import (
	"encoding/json"
)

type metadata struct {
	SchemaVersion    string
	Vendor           string
	Version          string
	ShortDescription string
}

func JSONMetadata() string {
	m := metadata{
		SchemaVersion:    "0.1.0",
		Vendor:           "https://github.com/michaelperel/docker-lock",
		Version:          "v0.1.0",
		ShortDescription: "Generate and validate lock files for Docker",
	}
	var jsonData []byte
	jsonData, err := json.Marshal(m)
	if err != nil {
		panic("Malformed metadata.")
	}
	return string(jsonData)
}
