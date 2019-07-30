package generate

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/michaelperel/docker-lock/registry"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Generator struct {
	Dockerfiles  []string
	Composefiles []string
	Outfile      string
}

type Image struct {
	Name   string `json:"name"`
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
}

type compose struct {
	Services map[string]struct {
		ImageName string `yaml:"image"`
		Build     string `yaml:"build"`
	} `yaml:"services"`
}

type imageResult struct {
	image Image
	err   error
}

type imageLineResult struct {
	line     string
	fileName string
	err      error
}

type Lockfile struct {
	Generator *Generator
	Images    []Image
}

func NewGenerator(flags *Flags) (*Generator, error) {
	//composefiles
	composefiles := []string{"docker-compose.yml"}

	//dockerfiles
	dockerfileSet := make(map[string]bool)
	for _, dockerfile := range flags.Dockerfiles {
		dockerfileSet[dockerfile] = true
	}
	if flags.Recursive {
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Base(path) == "Dockerfile" {
				dockerfileSet[path] = true
			}
			return nil
		})
	}
	for _, pattern := range flags.Globs {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			dockerfileSet[match] = true
		}
	}
	if len(dockerfileSet) == 0 {
		return &Generator{Dockerfiles: []string{"Dockerfile"}, Composefiles: composefiles, Outfile: flags.Outfile}, nil
	}
	dockerfiles := make([]string, len(dockerfileSet))
	i := 0
	for dockerfile := range dockerfileSet {
		dockerfiles[i] = dockerfile
		i++
	}
	return &Generator{Dockerfiles: dockerfiles, Composefiles: composefiles, Outfile: flags.Outfile}, nil
}

func (g *Generator) GenerateLockfile(wrapper registry.Wrapper) error {
	lockfileBytes, err := g.GenerateLockfileBytes(wrapper)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(g.Outfile, lockfileBytes, 0644)
}

func (g *Generator) GenerateLockfileBytes(wrapper registry.Wrapper) ([]byte, error) {
	images, err := g.getImages(wrapper)
	if err != nil {
		return nil, err
	}
	lockfile := Lockfile{Generator: g, Images: images}
	lockfileBytes, err := json.MarshalIndent(lockfile, "", "\t")
	if err != nil {
		return nil, err
	}
	return lockfileBytes, nil
}

func (g *Generator) requestImage(imLine imageLineResult, wrapper registry.Wrapper, imageResults chan<- imageResult) {
	tagSeparator := -1
	digestSeparator := -1
	for i, c := range imLine.line {
		if c == ':' {
			tagSeparator = i
		}
		if c == '@' {
			digestSeparator = i
			break
		}
	}
	// 4 valid cases
	// ubuntu:18.04@sha256:9b1702dcfe32c873a770a32cfd306dd7fc1c4fd134adfb783db68defc8894b3c
	if tagSeparator != -1 && digestSeparator != -1 {
		name := imLine.line[:tagSeparator]
		tag := imLine.line[tagSeparator+1 : digestSeparator]
		digest := imLine.line[digestSeparator+1+len("sha256:"):]
		imageResults <- imageResult{image: Image{Name: name, Tag: tag, Digest: digest}, err: nil}
		return
	}
	// ubuntu:18.04
	if tagSeparator != -1 && digestSeparator == -1 {
		name := imLine.line[:tagSeparator]
		tag := imLine.line[tagSeparator+1:]
		digest, err := wrapper.GetDigest(name, tag)
		if err != nil {
			err := fmt.Errorf("%s. From line: '%s'. From file: '%s'.", err, imLine.line, imLine.fileName)
			imageResults <- imageResult{image: Image{}, err: err}
			return
		}
		imageResults <- imageResult{image: Image{Name: name, Tag: tag, Digest: digest}, err: nil}
		return
	}
	// ubuntu@sha256:9b1702dcfe32c873a770a32cfd306dd7fc1c4fd134adfb783db68defc8894b3c
	if tagSeparator == -1 && digestSeparator != -1 {
		name := imLine.line[:digestSeparator]
		digest := imLine.line[digestSeparator+1+len("sha256:"):]
		imageResults <- imageResult{image: Image{Name: name, Digest: digest}, err: nil}
		return
	}
	// ubuntu
	if tagSeparator == -1 && digestSeparator == -1 {
		name := imLine.line
		tag := "latest"
		digest, err := wrapper.GetDigest(name, tag)
		if err != nil {
			err := fmt.Errorf("%s. From line: '%s'. From file: '%s'.", err, imLine.line, imLine.fileName)
			imageResults <- imageResult{image: Image{}, err: err}
			return
		}
		imageResults <- imageResult{image: Image{Name: name, Tag: tag, Digest: digest}, err: nil}
		return
	}
}

func (g *Generator) getDockerfileImageLines(imageLineResults chan<- imageLineResult, fileName string, wg *sync.WaitGroup) {
	defer wg.Done()
	dockerfile, err := os.Open(fileName)
	if err != nil {
		imageLineResults <- imageLineResult{fileName: fileName, err: err}
		return
	}
	defer dockerfile.Close()
	scanner := bufio.NewScanner(dockerfile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())
		if strings.HasPrefix(line, "from ") {
			line = strings.TrimPrefix(line, "from ")
			imageLineResults <- imageLineResult{line: line, fileName: fileName, err: nil}
		}
	}
}

func (g *Generator) getComposefileImageLines(imageLineResults chan<- imageLineResult, fileName string, wg *sync.WaitGroup) {
	defer wg.Done()
	yamlByt, err := ioutil.ReadFile(fileName)
	if err != nil {
		imageLineResults <- imageLineResult{fileName: fileName, err: err}
		return
	}
	var comp compose
	if err := yaml.Unmarshal(yamlByt, &comp); err != nil {
		imageLineResults <- imageLineResult{fileName: fileName, err: err}
		return
	}
	for _, svc := range comp.Services {
		if svc.Build == "" && svc.ImageName != "" {
			imageLineResults <- imageLineResult{line: svc.ImageName, fileName: fileName}
		}
		if svc.Build != "" {
			fi, err := os.Stat(svc.Build)
			if err != nil {
				imageLineResults <- imageLineResult{fileName: fileName, err: err}
				return
			}
			switch mode := fi.Mode(); {
			case mode.IsDir():
				dockerfile := path.Join(svc.Build, "Dockerfile")
				wg.Add(1)
				go g.getDockerfileImageLines(imageLineResults, dockerfile, wg)
			case mode.IsRegular():
				wg.Add(1)
				go g.getDockerfileImageLines(imageLineResults, svc.Build, wg)
			}
		}
	}
}

func (g *Generator) getImages(wrapper registry.Wrapper) ([]Image, error) {
	imageLineResults := make(chan imageLineResult)
	wg := new(sync.WaitGroup)
	for _, fileName := range g.Dockerfiles {
		wg.Add(1)
		go g.getDockerfileImageLines(imageLineResults, fileName, wg)
	}

	for _, fileName := range g.Composefiles {
		wg.Add(1)
		go g.getComposefileImageLines(imageLineResults, fileName, wg)
	}

	go func() {
		wg.Wait()
		close(imageLineResults)
	}()

	imageResults := make(chan imageResult)
	var numImages int
	for imLine := range imageLineResults {
		numImages++
		if imLine.err != nil {
			return nil, imLine.err
		}
		go g.requestImage(imLine, wrapper, imageResults)
	}

	var images []Image
	for i := 0; i < numImages; i++ {
		result := <-imageResults
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
