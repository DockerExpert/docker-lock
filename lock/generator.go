package lock

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/michaelperel/docker-lock/registry"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

type Generator struct {
	Options
}

type image struct {
	Name   string `json:"name"`
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
}

func (g *Generator) GenerateLockfile(wrapper registry.Wrapper) {
	lockfileBytes := g.generateLockfileBytes(wrapper)
	g.writeFile(lockfileBytes)
}

func (g *Generator) generateLockfileBytes(wrapper registry.Wrapper) []byte {
	images := g.getImages(wrapper)
	lockfileBytes, err := json.MarshalIndent(images, "", "\t")
	if err != nil {
		panic(err)
	}
	return lockfileBytes
}

func (g *Generator) writeFile(lockfileBytes []byte) {
	err := ioutil.WriteFile(g.Lockfile, lockfileBytes, 0644)
	if err != nil {
		panic(err)
	}
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

func (g *Generator) getImages(wrapper registry.Wrapper) []image {
	images := make([]image, 0)
	for _, dockerfile := range g.Dockerfiles {
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
				image, err := g.getImage(line, wrapper)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					fmt.Fprintf(os.Stderr, "File: '%s'.", dockerfile)
					os.Exit(1)
				}
				images = append(images, image)
			}
		}
	}
	sort.Slice(images, func(i, j int) bool { return images[i].Name < images[j].Name })
	return images
}
