package images

import (
	"fmt"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
)

func Remix(auth authn.Authenticator) error {
	// results := make([]Layer, 0)

	var base v1.Image
	var err error

	weights_image := "r8.im/anotherjesse/faster@sha256:2922cfb4febba1a72cacc9d407a726efe5a87ce32e2be5b4e5817209db87b7d1"

	fmt.Fprintln(os.Stderr, "fetching metadata for", weights_image)
	start := time.Now()
	weights, err := crane.Pull(weights_image, crane.WithAuth(auth))
	if err != nil {
		return fmt.Errorf("pulling %w", err)
	}
	fmt.Fprintln(os.Stderr, "pulling took", time.Since(start))

	base_image := "r8.im/anotherjesse/find@sha256:ef93356c06503ad651b7efe1ed705c58633826a0acfd664f50094eaac9829b79"
	fmt.Fprintln(os.Stderr, "fetching metadata for", base_image)
	start = time.Now()
	base, err = crane.Pull(base_image, crane.WithAuth(auth))
	if err != nil {
		return fmt.Errorf("pulling %w", err)
	}
	fmt.Fprintln(os.Stderr, "pulling took", time.Since(start))

	weights_layer_id := "sha256:23a377230e377792bdfb5321e5d470140c405366daa9bd5aa7d1c6ff3bc6f772"

	fmt.Fprintln(os.Stderr, "finding layer", weights_layer_id)
	layers, err := weights.Layers()
	if err != nil {
		return fmt.Errorf("getting layers %w", err)
	}
	var weights_layer v1.Layer
	for _, layer := range layers {
		digest, err := layer.Digest()
		if err != nil {
			return fmt.Errorf("getting digest %w", err)
		}
		if digest.String() == weights_layer_id {
			weights_layer = layer
			break
		}
	}

	start = time.Now()
	mutant, err := mutate.AppendLayers(base, weights_layer)
	if err != nil {
		return fmt.Errorf("appending layers %w", err)
	}
	fmt.Fprintln(os.Stderr, "appending layers took", time.Since(start))
	fmt.Fprintln(os.Stderr, "mutant image:", mutant)

	// --- pushing image

	dest := "r8.im/anotherjesse/faster"

	start = time.Now()

	err = crane.Push(mutant, dest, crane.WithAuth(auth))
	if err != nil {
		return fmt.Errorf("pushing %s: %w", dest, err)
	}

	fmt.Fprintln(os.Stderr, "pushing took", time.Since(start))

	// layers, err := base.Layers()
	// if err != nil {
	// 	return nil, fmt.Errorf("getting layers %w", err)
	// }

	// for _, layer := range layers {
	// 	digest, err := layer.Digest()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("getting digest %w", err)
	// 	}

	// 	size, err := layer.Size()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("getting size %w", err)
	// 	}
	// 	mediatype, err := layer.MediaType()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("getting mediatype %w", err)
	// 	}

	// 	results = append(results, Layer{
	// 		Digest:    digest.String(),
	// 		Size:      size,
	// 		MediaType: string(mediatype),
	// 	})
	// }

	// // Grab the commands from the history
	// cfg, nil := base.ConfigFile()
	// if err != nil {
	// 	return results, fmt.Errorf("getting config %w", err)
	// }
	// idx := 0
	// for _, h := range cfg.History {
	// 	if h.EmptyLayer {
	// 		continue
	// 	}

	// 	s := strings.TrimPrefix(h.CreatedBy, "/bin/sh -c ")
	// 	s = strings.TrimPrefix(s, "#(nop) ")
	// 	if len(s) > 40 {
	// 		s = s[:40]
	// 	}
	// 	results[idx].Command = s
	// 	idx++
	// }

	// return results, nil
	return nil
}
