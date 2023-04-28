package images

import (
	"fmt"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func ReallyRemix(baseRef string, weightsRef string, dest string, auth authn.Authenticator) (string, error) {
	fmt.Fprintln(os.Stderr, "fetching metadata for", weightsRef)
	start := time.Now()
	weightsImage, err := crane.Pull(weightsRef, crane.WithAuth(auth))
	if err != nil {
		return "", fmt.Errorf("pulling %w", err)
	}
	fmt.Fprintln(os.Stderr, "pulling took", time.Since(start))

	fmt.Fprintln(os.Stderr, "fetching metadata for", baseRef)
	start = time.Now()
	baseImage, err := crane.Pull(baseRef, crane.WithAuth(auth))
	if err != nil {
		return "", fmt.Errorf("pulling %w", err)
	}
	fmt.Fprintln(os.Stderr, "pulling took", time.Since(start))

	fmt.Fprintln(os.Stderr, "finding weights layer")

	start = time.Now()
	weightsLayer, err := findWeightsLayer(weightsImage)
	if err != nil {
		return "", fmt.Errorf("getting layers %w", err)
	}
	fmt.Fprintln(os.Stderr, "finding weights layer took", time.Since(start))

	start = time.Now()
	mutant, err := appendLayers(baseImage, weightsLayer)
	if err != nil {
		return "", fmt.Errorf("appending layers %w", err)
	}
	fmt.Fprintln(os.Stderr, "appending layers took", time.Since(start))

	fmt.Fprintln(os.Stderr, "mutant image:", mutant)

	// --- pushing image

	start = time.Now()

	err = crane.Push(mutant, dest, crane.WithAuth(auth))
	if err != nil {
		return "", fmt.Errorf("pushing %s: %w", dest, err)
	}

	fmt.Fprintln(os.Stderr, "pushing took", time.Since(start))

	return "mutant.hexdigest", nil
}

func findWeightsLayer(image v1.Image) (v1.Layer, error) {
	cfg, err := image.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("getting config %w", err)
	}
	idx := 0
	for _, h := range cfg.History {
		if h.EmptyLayer {
			continue
		}

		if h.Comment == "weights" {
			layers, err := image.Layers()
			if err != nil {
				return nil, fmt.Errorf("getting layers %w", err)
			}
			return layers[idx], nil
		}
		idx++
	}
	return nil, fmt.Errorf("no weights layer found")
}
