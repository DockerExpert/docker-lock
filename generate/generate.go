package generate

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Image struct {
	Name   string
	Tag    string
	Digest string
}

func LockFile() {
    dockerfiles := flagsToDockerfiles()
    images := dockerfilesToImages(dockerfiles)
    fmt.Println(images)
}

func flagsToDockerfiles() []string {
	var dockerfiles stringSliceFlag
	var recursive bool

	gFlag := flag.NewFlagSet("generate", flag.ExitOnError)
	gFlag.Var(&dockerfiles, "f", "Path to Dockerfile from current directory.")
	gFlag.BoolVar(&recursive, "r", false, "Recursively collect Dockerfiles from current directory.")
	gFlag.Parse(os.Args[3:])

	if recursive && len(dockerfiles) > 0 {
		fmt.Fprintf(os.Stderr, "Cannot specify both -r and -f\n")
		os.Exit(1)
	}
	if len(dockerfiles) > 0 {
        return []string(dockerfiles)
	}
	if recursive {
        return findDockerfilesInAllDirectories()
	}
    return []string{"Dockerfile"}
}

func findDockerfilesInAllDirectories() []string {
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
	tagIndex := strings.IndexByte(imageLine, ':')
	digestIndex := strings.IndexByte(imageLine, '@')
	// 4 valid cases
	// FROM ubuntu:18.04@sha256:9b1702dcfe32c873a770a32cfd306dd7fc1c4fd134adfb783db68defc8894b3c
	if tagIndex != -1 && digestIndex != -1 {
		name := imageLine[:tagIndex]
		tag := imageLine[tagIndex+1 : digestIndex]
		digest := imageLine[digestIndex+1:]
		return Image{Name: name, Tag: tag, Digest: digest}, nil
	}

	// FROM ubuntu:18.04
	if tagIndex != -1 && digestIndex == -1 {
		name := imageLine[:tagIndex]
		tag := imageLine[tagIndex+1:]
		// TODO: http call for digest
		return Image{Name: name, Tag: tag}, nil
	}

	// FROM ubuntu@sha256:9b1702dcfe32c873a770a32cfd306dd7fc1c4fd134adfb783db68defc8894b3c
	if tagIndex == -1 && digestIndex != -1 {
		name := imageLine[:digestIndex]
		// TODO: http call for tag
		digest := imageLine[digestIndex:]
		return Image{Name: name, Digest: digest}, nil
	}
	// FROM ubuntu
	if tagIndex == -1 && digestIndex == -1 {
		name := imageLine
		// TODO: http call for tag? Does it always equal latest??
		// TODO: http call for digest
		return Image{Name: name}, nil
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
                }
                images = append(images, image)
            }
        }
    }
	return images
}
