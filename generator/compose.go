package generator

import (
	"fmt"

	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Compose struct {
	Services map[string]struct {
		Image string `yaml:"image"`
		Build string `yaml:"build"`
	} `yaml:"services"`
}

func main() {
	yamlByt, err := ioutil.ReadFile("docker-compose.yaml")
	if err != nil {
		panic(err)
	}
	var compose Compose
	if err := yaml.Unmarshal(yamlByt, &compose); err != nil {
		panic(err)
	}
	fmt.Println(compose.Services["elasticsearch"])
}
