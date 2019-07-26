package generator

import (
	"bufio"
	"encoding/json"
	"errors"
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

type Image struct {
	Name   string `json:"name"`
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
}

type imageResult struct {
	image Image
	err   error
}

func New(dockerfiles []string, lockfile string) (*Generator, error) {
	if lockfile == "" {
		return nil, errors.New("Lockfile cannot be empty.")
	}
	return &Generator{Dockerfiles: dockerfiles, Lockfile: lockfile}, nil
}

func (g *Generator) GenerateLockfile(wrapper registry.Wrapper) error {
	lockfileBytes, err := g.GenerateLockfileBytes(wrapper)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(g.Lockfile, lockfileBytes, 0644)
}

func (g *Generator) GenerateLockfileBytes(wrapper registry.Wrapper) ([]byte, error) {
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

func (g *Generator) getImage(fromLine string, wrapper registry.Wrapper, imageCh chan<- imageResult) {
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
		imageCh <- imageResult{image: Image{Name: name, Tag: tag, Digest: digest}, err: nil}
		return
	}
	// FROM ubuntu:18.04
	if tagSeparator != -1 && digestSeparator == -1 {
		name := imageLine[:tagSeparator]
		tag := imageLine[tagSeparator+1:]
		digest, err := wrapper.GetDigest(name, tag)
		if err != nil {
			imageCh <- imageResult{image: Image{}, err: err}
			return
		}
		imageCh <- imageResult{image: Image{Name: name, Tag: tag, Digest: digest}, err: nil}
		return
	}
	// FROM ubuntu@sha256:9b1702dcfe32c873a770a32cfd306dd7fc1c4fd134adfb783db68defc8894b3c
	if tagSeparator == -1 && digestSeparator != -1 {
		name := imageLine[:digestSeparator]
		digest := imageLine[digestSeparator+1+len("sha256:"):]
		imageCh <- imageResult{image: Image{Name: name, Digest: digest}, err: nil}
		return
	}
	// FROM ubuntu
	if tagSeparator == -1 && digestSeparator == -1 {
		name := imageLine
		tag := "latest"
		digest, err := wrapper.GetDigest(name, tag)
		if err != nil {
			imageCh <- imageResult{image: Image{}, err: err}
			return
		}
		imageCh <- imageResult{image: Image{Name: name, Tag: tag, Digest: digest}, err: nil}
		return
	}
}

func (g *Generator) getImages(wrapper registry.Wrapper) ([]Image, error) {
	var images []Image
	var numImages int
	imageCh := make(chan imageResult)

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
				numImages++
				go g.getImage(line, wrapper, imageCh)
			}
		}
	}
	for i := 0; i < numImages; i++ {
		result := <-imageCh
		if result.err != nil {
			return nil, result.err
		}
		images = append(images, result.image)
	}
	sort.Slice(images, func(i, j int) bool {
		if images[i].Name != images[j].Name {
			return images[i].Name < images[j].Name
		}
		if images[i].Tag != images[j].Tag {
			return images[i].Tag < images[j].Tag
		}
		if images[i].Digest != images[j].Digest {
			return images[i].Digest < images[j].Digest
		}
		return true
	})
	return images, nil
}
