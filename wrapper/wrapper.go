package wrapper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Wrapper struct {
	Image string
	Tag   string
}

type digestResponse struct {
	Config struct {
		Digest string `json:"digest"`
	} `json:"config"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

func New(image string, tag string) *Wrapper {
	return &Wrapper{Image: image, Tag: tag}
}

func (w *Wrapper) GetDigest() string {
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
	decoder := json.NewDecoder(resp.Body)
	var d digestResponse
	err = decoder.Decode(&d)
	if err != nil {
		panic(err)
	}
	return d.Config.Digest
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

//func Example() {
//	w := New("library/ubuntu", "18.04")
//	fmt.Println(w.GetDigest())
//}
