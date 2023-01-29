package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func builder(baseRef string, dest string, newLayer string, auth authn.Authenticator) (string, error) {

	var base v1.Image
	var err error

	fmt.Println("fetching metadata for", baseRef)

	start := time.Now()
	base, err = crane.Pull(baseRef, crane.WithAuth(auth))
	if err != nil {
		return "", fmt.Errorf("pulling %w", err)
	}
	fmt.Println("pulling took", time.Since(start))

	// --- adding new layer ontop of existing image

	fmt.Println("appending as new layer", newLayer)

	start = time.Now()
	img, err := crane.Append(base, newLayer)
	if err != nil {
		return "", fmt.Errorf("appending %v: %w", newLayer, err)
	}
	fmt.Println("appending took", time.Since(start))

	// --- pushing image

	start = time.Now()

	err = crane.Push(img, dest, crane.WithAuth(auth))
	if err != nil {
		return "", fmt.Errorf("pushing %s: %w", dest, err)
	}

	fmt.Println("pushing took", time.Since(start))

	d, err := img.Digest()
	if err != nil {
		return "", err
	}
	image_id := fmt.Sprintf("%s@%s", dest, d)
	return image_id, nil
}

func addressWithScheme(address string) string {
	if strings.Contains(address, "://") {
		return address
	}
	return "https://" + address
}

func verifyToken(registryHost string, token string) (username string, err error) {
	if token == "" {
		return "", fmt.Errorf("token is required")
	}

	resp, err := http.PostForm(addressWithScheme(registryHost)+"/cog/v1/verify-token", url.Values{
		"token": []string{token},
	})
	if err != nil {
		return "", fmt.Errorf("failed to verify token: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("user does not exist")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to verify token, got status %d", resp.StatusCode)
	}
	body := &struct {
		Username string `json:"username"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(body); err != nil {
		return "", err
	}
	return body.Username, nil
}

func main() {
	var sToken string
	var sRegistry string
	var baseRef string
	var dest string
	var tar string
	flag.StringVar(&sToken, "token", "", "replicate cog token")
	flag.StringVar(&sRegistry, "registry", "r8.im", "registry host")
	flag.StringVar(&baseRef, "base", "", "base image reference - include tag: r8.im/username/modelname@sha256:hexdigest")
	flag.StringVar(&dest, "dest", "", "destination image reference: r8.im/username/modelname")
	flag.StringVar(&tar, "tar", "", "tar file to append as new layer")
	flag.Parse()

	u, err := verifyToken(sRegistry, sToken)
	if err != nil {
		fmt.Println("authentication error, invalid token or registry host error")
		panic(err)
	}
	auth := authn.FromConfig(authn.AuthConfig{Username: u, Password: sToken})

	if baseRef == "" {
		panic("--base missing! base image reference is required")
	}

	if dest == "" {
		panic("--dest missing! destination image reference is required")
	}

	if tar == "" {
		panic("--tar missing! tar file is required")
	}

	image_id, err := builder(baseRef, dest, tar, auth)
	if err != nil {
		panic(err)
	}

	fmt.Println(image_id)
}
