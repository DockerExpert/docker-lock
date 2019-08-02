package generate

import (
	"bufio"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type parsedImageLine struct {
	line     string
	fileName string
}

func parseComposefiles(fileNames []string) ([]parsedImageLine, error) {
	parsedImageLines := make([]parsedImageLine, 0)
	for _, fileName := range fileNames {
		fileName = "docker-compose.yml"
		yamlByt, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, err
		}
		var comp map[string]interface{}
		if err := yaml.Unmarshal(yamlByt, &comp); err != nil {
			return nil, err
		}
		services := comp["services"].(map[interface{}]interface{})
		for _, serviceConfig := range services {
			var result parsedImageLine
			result.fileName = fileName
			config := serviceConfig.(map[interface{}]interface{})
			if _, ok := config["build"]; !ok {
				imageName, _ := config["image"].(string)
				result.line = os.ExpandEnv(imageName)
				parsedImageLines = append(parsedImageLines, result)
				continue
			}
			switch build := config["build"].(type) {
			case string:
				if build != "" {
					build = os.ExpandEnv(build)
					fi, err := os.Stat(build)
					if err != nil {
						return nil, err
					}
					mode := fi.Mode()
					if mode.IsDir() {
						results, err := parseDockerfiles([]string{path.Join(build, "Dockerfile")}, nil)
						if err != nil {
							return nil, err
						}
						parsedImageLines = append(parsedImageLines, results...)
						continue
					} else {
						results, err := parseDockerfiles([]string{build}, nil)
						if err != nil {
							return nil, err
						}
						parsedImageLines = append(parsedImageLines, results...)
						continue
					}
				}
			case map[interface{}]interface{}:
				context := build["context"].(string)
				context = os.ExpandEnv(context)
				dockerfileName, _ := build["dockerfile"].(string)
				dockerfileName = os.ExpandEnv(dockerfileName)
				var dockerfile string
				if dockerfileName == "" {
					dockerfile = path.Join(context, "Dockerfile")
				} else {
					dockerfile = path.Join(context, dockerfileName)
				}
				args, _ := build["args"].([]interface{})
				argsMap := make(map[string]string)
				if len(args) == 0 {
					results, err := parseDockerfiles([]string{dockerfile}, nil)
					if err != nil {
						return nil, err
					}
					parsedImageLines = append(parsedImageLines, results...)
					continue
				} else {
					for _, arg := range args {
						argString := os.ExpandEnv(arg.(string))
						argsSlice := strings.Split(argString, "=")
						argsMap[argsSlice[0]] = argsSlice[1]
					}
					results, err := parseDockerfiles([]string{dockerfile}, argsMap)
					if err != nil {
						return nil, err
					}
					parsedImageLines = append(parsedImageLines, results...)
					continue
				}
			}
		}
	}
	return parsedImageLines, nil
}

func parseDockerfiles(fileNames []string, buildArgs map[string]string) ([]parsedImageLine, error) {
	parsedImageLines := make([]parsedImageLine, 0)
	for _, fileName := range fileNames {
		dockerfile, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}
		defer dockerfile.Close()
		stageNames := make(map[string]bool)
		buildVars := make(map[string]string)
		scanner := bufio.NewScanner(dockerfile)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) > 0 {
				switch instruction := strings.ToLower(fields[0]); instruction {
				case "arg", "env":
					//INSTRUCTION VAR1=VAL1 VAR2=VAL2 ...
					if strings.Contains(fields[1], "=") {
						for _, pair := range fields[1:] {
							splitPair := strings.Split(pair, "=")
							key, val := splitPair[0], splitPair[1]
							buildVars[key] = val
						}
					} else if len(fields) == 3 {
						//INSTUCTION VAR1 VAL1
						key, val := fields[1], fields[2]
						buildVars[key] = val
					} else if instruction == "arg" && len(fields) == 2 {
						// ARG VAR1
						argName := fields[1]
						if argVal, ok := buildArgs[argName]; ok {
							buildVars[argName] = argVal
						}
					}
				case "from":
					line := expandBuildVars(fields[1], buildVars)
					// each from resets buildvars
					buildVars = make(map[string]string)
					// guarding against the case where the line is the name of a previous build stage
					// rather than a base image.
					// For instance, FROM <previous-stage> AS <name>
					if !stageNames[line] {
						result := parsedImageLine{line: line, fileName: fileName}
						parsedImageLines = append(parsedImageLines, result)
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
	return parsedImageLines, nil
}

func expandBuildVars(line string, buildVars map[string]string) string {
	mapper := func(buildVar string) string {
		val, ok := buildVars[buildVar]
		if !ok {
			return val
		}
		// Remove excess quotes, for instance ARG="val" should be equivalent to ARG=val
		if len(val) > 0 && val[0] == '"' {
			val = val[1:]
		}
		if len(val) > 0 && val[len(val)-1] == '"' {
			val = val[:len(val)-1]
		}
		return val
	}
	return os.Expand(line, mapper)
}
