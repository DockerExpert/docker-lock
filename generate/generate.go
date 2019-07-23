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

type Cmd struct {
	Dockerfiles []string
	Recursive   bool
}

func parseCmd() Cmd {
	var cmd Cmd
	var dockerfiles stringSliceFlag
	var recursive bool

	gFlag := flag.NewFlagSet("generate", flag.ExitOnError)
	gFlag.Var(&dockerfiles, "f", "Path to Dockerfile from current directory.")
	gFlag.BoolVar(&recursive, "r", false, "Recursively collect Dockerfiles from current directory.")
	gFlag.Parse(os.Args[3:])

	// Handle invalid cases
	dockerfilesAndRecursiveSpecified := recursive && len(dockerfiles) > 0
	if dockerfilesAndRecursiveSpecified {
		fmt.Fprintf(os.Stderr, "Cannot specify both -r and -f\n")
		os.Exit(1)
	}

	// Handle valid Cases
	noFlags := !recursive && len(dockerfiles) == 0
	dockerfilesSpecified := len(dockerfiles) > 0
	recursiveSpecified := recursive
	if noFlags {
		cmd.Dockerfiles = []string{"Dockerfile"}
		return cmd
	}
	if dockerfilesSpecified {
		cmd.Dockerfiles = []string(dockerfiles)
		return cmd
	}
	if recursiveSpecified {
		cmd.Recursive = recursive
		cmd.Dockerfiles = findDockerfilesInAllDirectories()
		return cmd
	}
	return cmd
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

func LockFile() {
	cmd := parseCmd()
	fmt.Println(cmd)
}

func parseFrom(line string) (Image, error) {
	// 4 cases: - node:8.10.0@sha256:06ebd9b1879057e24c1e87db508ba9fd0dd7f766bbf55665652d31487ca194eb
	//          - node:8.10.0
	//          - node@sha256:06ebd9b1879057e24c1e87db508ba9fd0dd7f766bbf55665652d31487ca194eb
	//          - node
	var tagExists, digestExists bool
	tagIndex := strings.IndexByte(line, ':')
	if tagIndex != -1 {
		tagExists = true
	}
	digestIndex := strings.IndexByte(line, '@')
	if digestIndex != -1 {
		digestExists = true
	}

	if tagExists && digestExists {
		name := line[:tagIndex]
		tag := line[tagIndex+1 : digestIndex]
		digest := line[digestIndex+1:]
		return Image{Name: name, Tag: tag, Digest: digest}, nil
	}
	if tagExists && !digestExists {
		name := line[:tagIndex]
		tag := line[tagIndex+1:]
		// TODO: http call for digest
		return Image{Name: name, Tag: tag}, nil
	}
	if !tagExists && digestExists {
		name := line[:digestIndex]
		// TODO: http call for tag
		digest := line[digestIndex:]
		return Image{Name: name, Digest: digest}, nil
	}
	if !tagExists && !digestExists {
		name := line
		// TODO: http call for tag? Does it always equal latest??
		// TODO: http call for digest
		return Image{Name: name}, nil
	}
	return Image{}, errors.New("Malformed Dockerfile")
}

func extractImages(df string) []string {
	openDf, err := os.Open(df)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer openDf.Close()
	dfScanner := bufio.NewScanner(openDf)
	dfScanner.Split(bufio.ScanLines)
	fromLines := make([]string, 0)
	for dfScanner.Scan() {
		line := strings.ToLower(dfScanner.Text())
		if strings.HasPrefix(line, "from") {
			fromLines = append(fromLines, dfScanner.Text())
		}
	}
	return fromLines
}
