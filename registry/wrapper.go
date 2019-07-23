package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Wrapper struct {
	Image string
	Tag   string
}

type tokenResponse struct {
	Token string `json:"token"`
}

func NewWrapper(image string, tag string) *Wrapper {
	return &Wrapper{Image: image, Tag: tag}
}

func (w *Wrapper) GetDigest() (string, error) {
	// Docker-Content-Digest is the root of the hash chain
	// https://github.com/docker/distribution/issues/1662
	token := w.getToken()
	registryUrl := "https://registry-1.docker.io/v2/" + w.Image + "/manifests/" + w.Tag
	req, err := http.NewRequest("GET", registryUrl, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" && !strings.HasPrefix(w.Image, "library/") {
		w.Image = "library/" + w.Image
		return w.GetDigest()
	}
	if digest == "" {
		return "", errors.New("No digest found")
	}
	return strings.TrimPrefix(digest, "sha256:"), nil
}

func (w *Wrapper) getToken() string {
	client := &http.Client{}
	url := "https://auth.docker.io/token?scope=repository:" + w.Image + ":pull&service=registry.docker.io"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var t tokenResponse
	err = decoder.Decode(&t)
	if err != nil {
		panic(err)
	}
	return t.Token
}
