package lock

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/michaelperel/docker-lock/options"
	"github.com/michaelperel/docker-lock/wrapper"
	"os"
	"path/filepath"
	"strings"
)

type image struct {
	Name   string `json:"name"`
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
}

func Generate(options options.Options) []byte {
	dockerfiles := getDockerfiles(options)
	images := getimages(dockerfiles)
	lockfileBytes, err := json.MarshalIndent(images, "", "\t")
	if err != nil {
		panic(err)
	}
	return lockfileBytes
}

func getDockerfiles(options options.Options) []string {
	if len(options.Dockerfiles) != 0 {
		return options.Dockerfiles
	}
	if options.Recursive {
		return getDockerfilesRecursively()
	}
	return []string{"Dockerfile"}
}

func getDockerfilesRecursively() []string {
	dockerfiles := make([]string, 0)
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) == "Dockerfile" {
			dockerfiles = append(dockerfiles, path)
		}
		return nil
	})
	return dockerfiles
}

func getimage(fromLine string) (image, error) {
	imageLine := strings.TrimPrefix(fromLine, "from ")
	tagSeparator := -1
	digestSeparator := -1
	for i, c := range imageLine {
		if c == ':' {
			tagSeparator = i
		}
		if c == '@' {
			digestSeparator = i
			break
		}
	}
	// 4 valid cases
	// FROM ubuntu:18.04@sha256:9b1702dcfe32c873a770a32cfd306dd7fc1c4fd134adfb783db68defc8894b3c
	if tagSeparator != -1 && digestSeparator != -1 {
		name := imageLine[:tagSeparator]
		tag := imageLine[tagSeparator+1 : digestSeparator]
		digest := imageLine[digestSeparator+1+len("sha256:"):]
		return image{Name: name, Tag: tag, Digest: digest}, nil
	}
	// FROM ubuntu:18.04
	if tagSeparator != -1 && digestSeparator == -1 {
		name := imageLine[:tagSeparator]
		tag := imageLine[tagSeparator+1:]
		w := wrapper.New(name, tag)
		digest, err := w.GetDigest()
		if err != nil {
			return image{}, fmt.Errorf("Unable to retrieve digest from line '%s'.", fromLine)
		}
		return image{Name: name, Tag: tag, Digest: digest}, nil
	}
	// FROM ubuntu@sha256:9b1702dcfe32c873a770a32cfd306dd7fc1c4fd134adfb783db68defc8894b3c
	if tagSeparator == -1 && digestSeparator != -1 {
		name := imageLine[:digestSeparator]
		digest := imageLine[digestSeparator+1+len("sha256:"):]
		return image{Name: name, Digest: digest}, nil
	}
	// FROM ubuntu
	if tagSeparator == -1 && digestSeparator == -1 {
		name := imageLine
		tag := "latest"
		w := wrapper.New(name, tag)
		digest, err := w.GetDigest()
		if err != nil {
			return image{}, fmt.Errorf("Unable to retrieve digest from line '%s'.", fromLine)
		}
		return image{Name: name, Tag: tag, Digest: digest}, nil
	}
	return image{}, fmt.Errorf("Malformed from line: '%s'.", fromLine)
}

func getimages(dockerfiles []string) []image {
	images := make([]image, 0)
	for _, dockerfile := range dockerfiles {
		openDockerfile, err := os.Open(dockerfile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer openDockerfile.Close()
		scanner := bufio.NewScanner(openDockerfile)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := strings.ToLower(scanner.Text())
			if strings.HasPrefix(line, "from ") {
				image, err := getimage(line)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					fmt.Fprintf(os.Stderr, "File: '%s'.", dockerfile)
					os.Exit(1)
				}
				images = append(images, image)
			}
		}
	}
	return images
}
