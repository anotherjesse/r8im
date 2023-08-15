package images

import (
	"fmt"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// FIXME(ja): the mediatypes of layers are tar.gzip? does that mean we should create weights as tar.gzip to go faster?

func Affix(baseRef string, dest string, newLayer string, auth authn.Authenticator) (string, error) {

	var base v1.Image
	var err error

	fmt.Fprintln(os.Stderr, "fetching metadata for", baseRef)

	start := time.Now()
	base, err = crane.Pull(baseRef, crane.WithAuth(auth))
	if err != nil {
		return "", fmt.Errorf("pulling %w", err)
	}
	fmt.Fprintln(os.Stderr, "pulling took", time.Since(start))

	// --- adding new layer ontop of existing image

	var img v1.Image

	if newLayer != "" {
		fmt.Fprintln(os.Stderr, "appending as new layer", newLayer)

		start = time.Now()
		img, err = appendLayer(base, newLayer)
		if err != nil {
			return "", fmt.Errorf("appending %v: %w", newLayer, err)
		}
		fmt.Fprintln(os.Stderr, "appending took", time.Since(start))
	} else {
		cfg, err := base.ConfigFile()
		if err != nil {
			return "", fmt.Errorf("getting config file: %w", err)
		}

		cfg.Config.Labels["cloned"] = "true"

		img, err = mutate.ConfigFile(base, cfg)
		if err != nil {
			return "", fmt.Errorf("mutating config file: %w", err)
		}
	}
	// --- pushing image

	start = time.Now()

	err = crane.Push(img, dest, crane.WithAuth(auth))
	if err != nil {
		return "", fmt.Errorf("pushing %s: %w", dest, err)
	}

	fmt.Fprintln(os.Stderr, "pushing took", time.Since(start))

	d, err := img.Digest()
	if err != nil {
		return "", err
	}
	image_id := fmt.Sprintf("%s@%s", dest, d)
	return image_id, nil
}

// All of this code is from pkg/v1/mutate - so we can add history

func appendLayer(base v1.Image, path string) (v1.Image, error) {
	baseMediaType, err := base.MediaType()
	if err != nil {
		return nil, fmt.Errorf("getting base image media type: %w", err)
	}
	layerType := types.DockerLayer

	if baseMediaType == types.OCIManifestSchema1 {
		layerType = types.OCILayer
	}

	layers := make([]v1.Layer, 0, 1)
	layer, err := getLayer(path, layerType)
	if err != nil {
		return nil, fmt.Errorf("reading layer %q: %w", path, err)
	}
	layers = append(layers, layer)

	return appendLayers(base, layers...)
}

func getLayer(path string, layerType types.MediaType) (v1.Layer, error) {
	f, err := streamFile(path)
	if err != nil {
		return nil, err
	}
	if f != nil {
		return stream.NewLayer(f, stream.WithMediaType(layerType)), nil
	}

	return tarball.LayerFromFile(path, tarball.WithMediaType(layerType))
}

func streamFile(path string) (*os.File, error) {
	if path == "-" {
		return os.Stdin, nil
	}
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !fi.Mode().IsRegular() {
		return os.Open(path)
	}

	return nil, nil
}

func appendLayers(base v1.Image, layers ...v1.Layer) (v1.Image, error) {
	additions := make([]mutate.Addendum, 0, len(layers))
	history := v1.History{
		CreatedBy: "cp . /src/weights # weights",
		Created:   v1.Time{Time: time.Now()},
		Author:    "r8im",
		Comment:   "weights",
	}

	for _, layer := range layers {
		additions = append(additions, mutate.Addendum{Layer: layer, History: history})
	}

	return mutate.Append(base, additions...)
}
