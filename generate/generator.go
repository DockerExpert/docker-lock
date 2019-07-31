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
	outfile      string
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
	dockerfiles, err := getDockerfiles(flags)
	if err != nil {
		return nil, err
	}
	composefiles, err := getComposefiles(flags)
	if err != nil {
		return nil, err
	}
	if len(dockerfiles) == 0 && len(composefiles) == 0 {
		fi, err := os.Stat("Dockerfile")
		if err == nil {
			if mode := fi.Mode(); mode.IsRegular() {
				dockerfiles = []string{"Dockerfile"}
			}
		}
		for _, defaultComposefile := range []string{"docker-compose.yml", "docker-compose.yaml"} {
			fi, err := os.Stat(defaultComposefile)
			if err == nil {
				if mode := fi.Mode(); mode.IsRegular() {
					composefiles = append(composefiles, defaultComposefile)
				}
			}
		}
	}
	return &Generator{Dockerfiles: dockerfiles, Composefiles: composefiles, outfile: flags.Outfile}, nil
}

func getDockerfiles(flags *Flags) ([]string, error) {
	isDefaultDockerfile := func(fpath string) bool {
		return filepath.Base(fpath) == "Dockerfile"
	}
	return getFiles(flags.Dockerfiles, flags.Recursive, isDefaultDockerfile, flags.Globs)
}

func getComposefiles(flags *Flags) ([]string, error) {
	isDefaultComposefile := func(fpath string) bool {
		return filepath.Base(fpath) == "docker-compose.yml" || filepath.Base(fpath) == "docker-compose.yaml"
	}
	return getFiles(flags.Composefiles, flags.ComposeRecursive, isDefaultComposefile, flags.ComposeGlobs)
}

func getFiles(files []string, recursive bool, isDefaultName func(string) bool, globs []string) ([]string, error) {
	fileSet := make(map[string]bool)
	for _, fileName := range files {
		fileSet[fileName] = true
	}
	if recursive {
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if isDefaultName(filepath.Base(path)) {
				fileSet[path] = true
			}
			return nil
		})
	}
	for _, pattern := range globs {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			fileSet[match] = true
		}
	}
	if len(fileSet) == 0 {
		return []string{}, nil
	}
	collectedFiles := make([]string, len(fileSet))
	i := 0
	for file := range fileSet {
		collectedFiles[i] = file
		i++
	}
	return collectedFiles, nil
}

func (g *Generator) GenerateLockfile(wrapper registry.Wrapper) error {
	lockfileBytes, err := g.GenerateLockfileBytes(wrapper)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(g.outfile, lockfileBytes, 0644)
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

func (g *Generator) getImages(wrapper registry.Wrapper) ([]Image, error) {
	imageLineResults := make(chan imageLineResult)
	wg := new(sync.WaitGroup)
	for _, fileName := range g.Dockerfiles {
		wg.Add(1)
		go g.parseDockerfile(imageLineResults, fileName, wg)
	}

	for _, fileName := range g.Composefiles {
		wg.Add(1)
		go g.parseComposefile(imageLineResults, fileName, wg)
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
		go g.getImage(imLine, wrapper, imageResults)
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

func (g *Generator) getImage(imLine imageLineResult, wrapper registry.Wrapper, imageResults chan<- imageResult) {
	line := imLine.line
	tagSeparator := -1
	digestSeparator := -1
	for i, c := range line {
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
		name := line[:tagSeparator]
		tag := line[tagSeparator+1 : digestSeparator]
		digest := line[digestSeparator+1+len("sha256:"):]
		imageResults <- imageResult{image: Image{Name: name, Tag: tag, Digest: digest}, err: nil}
		return
	}
	// ubuntu:18.04
	if tagSeparator != -1 && digestSeparator == -1 {
		name := line[:tagSeparator]
		tag := line[tagSeparator+1:]
		digest, err := wrapper.GetDigest(name, tag)
		if err != nil {
			err := fmt.Errorf("%s. From line: '%s'. From file: '%s'.", err, line, imLine.fileName)
			imageResults <- imageResult{image: Image{}, err: err}
			return
		}
		imageResults <- imageResult{image: Image{Name: name, Tag: tag, Digest: digest}, err: nil}
		return
	}
	// ubuntu@sha256:9b1702dcfe32c873a770a32cfd306dd7fc1c4fd134adfb783db68defc8894b3c
	if tagSeparator == -1 && digestSeparator != -1 {
		name := line[:digestSeparator]
		digest := line[digestSeparator+1+len("sha256:"):]
		imageResults <- imageResult{image: Image{Name: name, Digest: digest}, err: nil}
		return
	}
	// ubuntu
	if tagSeparator == -1 && digestSeparator == -1 {
		name := line
		tag := "latest"
		digest, err := wrapper.GetDigest(name, tag)
		if err != nil {
			err := fmt.Errorf("%s. From line: '%s'. From file: '%s'.", err, line, imLine.fileName)
			imageResults <- imageResult{image: Image{}, err: err}
			return
		}
		imageResults <- imageResult{image: Image{Name: name, Tag: tag, Digest: digest}, err: nil}
		return
	}
}

func (g *Generator) parseDockerfile(imageLineResults chan<- imageLineResult, fileName string, wg *sync.WaitGroup) {
	defer wg.Done()
	dockerfile, err := os.Open(fileName)
	if err != nil {
		imageLineResults <- imageLineResult{fileName: fileName, err: err}
		return
	}
	defer dockerfile.Close()
	stageNames := make(map[string]bool)
	buildVars := make(map[string]string)
	scanner := bufio.NewScanner(dockerfile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 0 {
			switch instruction := fields[0]; instruction {
			case "ARG", "ENV", "arg", "env":
				switch {
				//INSTRUCTION VAR1=VAL1 VAR2=VAL2 ...
				case strings.Contains(fields[1], "="):
					for _, pair := range fields[1:] {
						splitPair := strings.Split(pair, "=")
						key, val := splitPair[0], splitPair[1]
						buildVars[key] = val
					}
					//INSTUCTION VAR1 VAL1
				case len(fields) == 3:
					key, val := fields[1], fields[2]
					buildVars[key] = val
				}
			case "FROM", "from":
				line := expandBuildVars(fields[1], buildVars)
				// guarding against the case where the line is the name of a previous build stage
				// rather than a base image.
				// For instance, FROM <previous-stage> AS <name>
				if !stageNames[line] {
					imageLineResults <- imageLineResult{line: line, fileName: fileName, err: nil}
				}
				// multistage build
				// FROM <image> AS <name>
				// FROM <previous-stage> as <name>
				if len(fields) == 4 {
					stageName := expandBuildVars(fields[3], buildVars)
					stageNames[stageName] = true
				}
			}
		}
	}
}

func expandBuildVars(line string, buildVars map[string]string) string {
	mapper := func(buildVar string) string {
		return buildVars[buildVar]
	}
	return os.Expand(line, mapper)
}

func (g *Generator) parseComposefile(imageLineResults chan<- imageLineResult, fileName string, wg *sync.WaitGroup) {
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
			line := os.ExpandEnv(svc.ImageName)
			imageLineResults <- imageLineResult{line: line, fileName: fileName}
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
				go g.parseDockerfile(imageLineResults, dockerfile, wg)
			case mode.IsRegular():
				wg.Add(1)
				go g.parseDockerfile(imageLineResults, svc.Build, wg)
			}
		}
	}
}
