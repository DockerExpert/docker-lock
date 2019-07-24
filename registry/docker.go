package registry

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
)

type DockerWrapper struct{}

type tokenResponse struct {
	Token string `json:"token"`
}

func (w *DockerWrapper) GetDigest(name string, tag string) (string, error) {
	// Docker-Content-Digest is the root of the hash chain
	// https://github.com/docker/distribution/issues/1662
	token, err := w.getToken(name)
	if err != nil {
		return "", err
	}
	registryUrl := "https://registry-1.docker.io/v2/" + name + "/manifests/" + tag
	req, err := http.NewRequest("GET", registryUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" && !strings.HasPrefix(name, "library/") {
		name = "library/" + name
		return w.GetDigest(name, tag)
	}
	if digest == "" {
		return "", errors.New("No digest found")
	}
	return strings.TrimPrefix(digest, "sha256:"), nil
}

func (w *DockerWrapper) getToken(name string) (string, error) {
	client := &http.Client{}
	url := "https://auth.docker.io/token?scope=repository:" + name + ":pull&service=registry.docker.io"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	username := os.Getenv("DOCKER_USERNAME")
	password := os.Getenv("DOCKER_PASSWORD")
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var t tokenResponse
	if err = decoder.Decode(&t); err != nil {
		return "", err
	}
	return t.Token, nil
}
