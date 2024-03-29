package images

import (
	"fmt"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/compression"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Uncompress(imageName string, dest string, auth authn.Authenticator) (string, error) {
	var base v1.Image
	var err error

	fmt.Fprintln(os.Stderr, "fetching metadata for", imageName)

	start := time.Now()
	base, err = crane.Pull(imageName, crane.WithAuth(auth))
	if err != nil {
		return "", fmt.Errorf("pulling %w", err)
	}
	fmt.Fprintln(os.Stderr, "pulling took", time.Since(start))

	img, err := uncompress(base)
	if err != nil {
		return "", err
	}

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

func uncompress(base v1.Image) (v1.Image, error) {

	// inspired by https://github.com/google/go-containerregistry/blob/v0.15.2/pkg/v1/mutate/mutate.go#L371
	newImage := empty.Image

	layers, err := base.Layers()
	if err != nil {
		return nil, fmt.Errorf("getting layers %w", err)
	}

	ocf, err := base.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("getting original config file %w", err)
	}

	addendums := make([]mutate.Addendum, max(len(ocf.History), len(layers)))

	startUncompressing := time.Now()

	var historyIdx, addendumIdx int
	for layerIdx := 0; layerIdx < len(layers); addendumIdx, layerIdx = addendumIdx+1, layerIdx+1 {
		startLayer := time.Now()
		newLayer, err := uncompressedLayer(layers[layerIdx])
		if err != nil {
			return nil, fmt.Errorf("setting uncompressed layer: %w", err)
		}
		fmt.Fprintln(os.Stderr, "uncompressing layer", layerIdx, "took", time.Since(startLayer))

		// try to search for the history entry that corresponds to this layer
		for ; historyIdx < len(ocf.History); historyIdx++ {
			addendums[addendumIdx].History = ocf.History[historyIdx]
			// if it's an EmptyLayer, do not set the Layer and have the Addendum with just the History
			// and move on to the next History entry
			if ocf.History[historyIdx].EmptyLayer {
				addendumIdx++
				continue
			}
			// otherwise, we can exit from the cycle
			historyIdx++
			break
		}
		addendums[addendumIdx].Layer = newLayer
	}
	fmt.Fprintln(os.Stderr, "total uncompressing took", time.Since(startUncompressing))

	// add all leftover History entries
	for ; historyIdx < len(ocf.History); historyIdx, addendumIdx = historyIdx+1, addendumIdx+1 {
		addendums[addendumIdx].History = ocf.History[historyIdx]
	}

	newImage, err = mutate.Append(newImage, addendums...)
	if err != nil {
		return nil, fmt.Errorf("Appending: %w", err)
	}

	cf, err := newImage.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("setting config file: %w", err)
	}

	cfg := cf.DeepCopy()

	// Copy basic config over
	cfg.Architecture = ocf.Architecture
	cfg.OS = ocf.OS
	cfg.OSVersion = ocf.OSVersion
	cfg.Config = ocf.Config

	// Strip away timestamps from the config file
	// cfg.Created = v1.Time{Time: t}

	for i, h := range cfg.History {
		// h.Created = v1.Time{Time: t}
		h.CreatedBy = ocf.History[i].CreatedBy
		h.Comment = ocf.History[i].Comment
		h.EmptyLayer = ocf.History[i].EmptyLayer
		// Explicitly ignore Author field; which hinders reproducibility
		h.Author = ""
		cfg.History[i] = h
	}

	return mutate.ConfigFile(newImage, cfg)
}

func uncompressedLayer(layer v1.Layer) (v1.Layer, error) {
	newLayer, err := tarball.LayerFromOpener(layer.Compressed,
		tarball.WithCompression(compression.None),
		tarball.WithMediaType(types.DockerUncompressedLayer),
	)
	if err != nil {
		return nil, fmt.Errorf("creating new layer: %w", err)
	}
	return newLayer, nil
}
