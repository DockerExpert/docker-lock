package generate

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/michaelperel/docker-lock/options"
	"github.com/michaelperel/docker-lock/wrapper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Image struct {
	Name   string `json:"name"`
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
}

func LockFile(options options.Options) {
	dockerfiles := getDockerfiles(options)
	images := getImages(dockerfiles)
	lockFile, err := json.MarshalIndent(images, "", "\t")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("docker-lock.json", lockFile, 0644)
	if err != nil {
		panic(err)
	}
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

func getImage(fromLine string) (Image, error) {
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
		return Image{Name: name, Tag: tag, Digest: digest}, nil
	}
	// FROM ubuntu:18.04
	if tagSeparator != -1 && digestSeparator == -1 {
		name := imageLine[:tagSeparator]
		tag := imageLine[tagSeparator+1:]
		w := wrapper.New(name, tag)
		digest, err := w.GetDigest()
		if err != nil {
			return Image{}, fmt.Errorf("Unable to retrieve digest from line '%s'.", fromLine)
		}
		return Image{Name: name, Tag: tag, Digest: digest}, nil
	}
	// FROM ubuntu@sha256:9b1702dcfe32c873a770a32cfd306dd7fc1c4fd134adfb783db68defc8894b3c
	if tagSeparator == -1 && digestSeparator != -1 {
		name := imageLine[:digestSeparator]
		digest := imageLine[digestSeparator+1+len("sha256:"):]
		return Image{Name: name, Digest: digest}, nil
	}
	// FROM ubuntu
	if tagSeparator == -1 && digestSeparator == -1 {
		name := imageLine
		tag := "latest"
		w := wrapper.New(name, tag)
		digest, err := w.GetDigest()
		if err != nil {
			return Image{}, fmt.Errorf("Unable to retrieve digest from line '%s'.", fromLine)
		}
		return Image{Name: name, Tag: tag, Digest: digest}, nil
	}
	return Image{}, fmt.Errorf("Malformed from line: '%s'.", fromLine)
}

func getImages(dockerfiles []string) []Image {
	images := make([]Image, 0)
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
				image, err := getImage(line)
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
