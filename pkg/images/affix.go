package images

import (
	"fmt"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func Affix(baseRef string, dest string, newLayer string, auth authn.Authenticator) (string, error) {

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
