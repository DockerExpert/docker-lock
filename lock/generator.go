package lock

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/michaelperel/docker-lock/registry"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

type Generator struct {
	Dockerfiles []string
	Lockfile    string
}

type image struct {
	Name   string `json:"name"`
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
}

func NewGenerator(dockerfiles []string, lockfile string) (*Generator, error) {
	if lockfile == "" {
		return nil, errors.New("Lockfile cannot be empty.")
	}
	return &Generator{Dockerfiles: dockerfiles, Lockfile: lockfile}, nil
}

func (g *Generator) GenerateLockfile(wrapper registry.Wrapper) error {
	lockfileBytes, err := g.generateLockfileBytes(wrapper)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(g.Lockfile, lockfileBytes, 0644)
}

func (g *Generator) generateLockfileBytes(wrapper registry.Wrapper) ([]byte, error) {
	images, err := g.getImages(wrapper)
	if err != nil {
		return nil, err
	}
	lockfileBytes, err := json.MarshalIndent(images, "", "\t")
	if err != nil {
		return nil, err
	}
	return lockfileBytes, nil
}

func (g *Generator) getImage(fromLine string, wrapper registry.Wrapper) (image, error) {
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
		digest, err := wrapper.GetDigest(name, tag)
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
		digest, err := wrapper.GetDigest(name, tag)
		if err != nil {
			return image{}, fmt.Errorf("Unable to retrieve digest from line '%s'.", fromLine)
		}
		return image{Name: name, Tag: tag, Digest: digest}, nil
	}
	return image{}, fmt.Errorf("Malformed from line: '%s'.", fromLine)
}

func (g *Generator) getImages(wrapper registry.Wrapper) ([]image, error) {
	images := make([]image, 0)
	for _, dockerfile := range g.Dockerfiles {
		openDockerfile, err := os.Open(dockerfile)
		if err != nil {
			return nil, err
		}
		defer openDockerfile.Close()
		scanner := bufio.NewScanner(openDockerfile)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := strings.ToLower(scanner.Text())
			if strings.HasPrefix(line, "from ") {
				image, lineErr := g.getImage(line, wrapper)
				if lineErr != nil {
					fileErr := fmt.Errorf("File: '%s'.", dockerfile)
					return nil, fmt.Errorf("%s %s", err, fileErr)
				}
				images = append(images, image)
			}
		}
	}
	sort.Slice(images, func(i, j int) bool { return images[i].Name < images[j].Name })
	return images, nil
}
