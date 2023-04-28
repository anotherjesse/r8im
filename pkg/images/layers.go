package images

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type Layer struct {
	Digest    string
	MediaType string
	Size      int64
	Command   string
	Raw       v1.Layer
}

func Layers(imageName string, auth authn.Authenticator) ([]Layer, error) {
	results := make([]Layer, 0)

	var base v1.Image
	var err error

	fmt.Fprintln(os.Stderr, "fetching metadata for", imageName)

	start := time.Now()
	base, err = crane.Pull(imageName, crane.WithAuth(auth))
	if err != nil {
		return nil, fmt.Errorf("pulling %w", err)
	}
	fmt.Fprintln(os.Stderr, "pulling took", time.Since(start))

	layers, err := base.Layers()
	if err != nil {
		return nil, fmt.Errorf("getting layers %w", err)
	}

	for _, layer := range layers {
		digest, err := layer.Digest()
		if err != nil {
			return nil, fmt.Errorf("getting digest %w", err)
		}

		size, err := layer.Size()
		if err != nil {
			return nil, fmt.Errorf("getting size %w", err)
		}
		mediatype, err := layer.MediaType()
		if err != nil {
			return nil, fmt.Errorf("getting mediatype %w", err)
		}

		results = append(results, Layer{
			Digest:    digest.String(),
			Size:      size,
			MediaType: string(mediatype),
			Raw:       layer,
		})
	}

	// Grab the commands from the history
	cfg, nil := base.ConfigFile()
	if err != nil {
		return results, fmt.Errorf("getting config %w", err)
	}
	idx := 0
	for _, h := range cfg.History {
		if h.EmptyLayer {
			continue
		}

		s := strings.TrimPrefix(h.CreatedBy, "/bin/sh -c ")
		s = strings.TrimPrefix(s, "#(nop) ")
		if len(s) > 40 {
			s = s[:40]
		}
		results[idx].Command = s
		idx++
	}

	return results, nil
}
