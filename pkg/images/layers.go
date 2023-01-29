package images

import (
	"fmt"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type Layer struct {
	Digest    string
	MediaType string
	Size      int64
	// FIXME(ja): can we get the command or annotations?
}

func Layers(imageName string, auth authn.Authenticator) ([]Layer, error) {
	var base v1.Image
	var err error

	fmt.Println("fetching metadata for", imageName)

	start := time.Now()
	base, err = crane.Pull(imageName, crane.WithAuth(auth))
	if err != nil {
		return nil, fmt.Errorf("pulling %w", err)
	}
	fmt.Println("pulling took", time.Since(start))

	layers, err := base.Layers()
	if err != nil {
		return nil, fmt.Errorf("getting layers %w", err)
	}

	results := make([]Layer, 0)

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
		})
	}

	return results, nil
}

// Tags:   (unavailable)                                                                                             -rw-r--r--         0:0     3.0 kB  │   ├── adduser.conf
// Id:     ca3ced397349afdb5201203e65af600910f0af4b3d20c9d73e9e322e7ec77bfa                                          drwxr-xr-x         0:0      100 B  │   ├── alternatives
// Digest: sha256:3e271fb25447b8677bb29b9eef44d0e0a83bfe31c05c92b3b4f6121d58e8448c                                   -rw-r--r--         0:0      100 B  │   │   ├── README
// Command:                                                                                                          -rwxrwxrwx         0:0        0 B  │   │   ├── awk → /usr/bin/mawk
