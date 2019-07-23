package generate

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/michaelperel/docker-lock/wrapper"
    "github.com/michaelperel/docker-lock/options"
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
    images := dockerfilesToImages(dockerfiles)
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

func imageLineToImage(imageLine string) (Image, error) {
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
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
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
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return Image{Name: name, Tag: tag, Digest: digest}, nil
	}
	return Image{}, errors.New("Malformed base image: " + imageLine)
}

func dockerfilesToImages(dockerfiles []string) []Image {
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
				imageLine := strings.TrimPrefix(line, "from ")
				image, err := imageLineToImage(imageLine)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				images = append(images, image)
			}
		}
	}
	return images
}
