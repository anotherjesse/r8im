package images

import (
	"fmt"
	"io"
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

type uncompressedLayer struct {
	wrapped v1.Layer
	digest  v1.Hash
	size    int64
}

var _ v1.Layer = &uncompressedLayer{}

// Digest implements Layer.Digest()
func (u *uncompressedLayer) Digest() (v1.Hash, error) {
	return u.digest, nil
}

// DiffID returns the Hash of the uncompressed layer.
func (u *uncompressedLayer) DiffID() (v1.Hash, error) {
	return u.wrapped.DiffID()
}

// Compressed returns an io.ReadCloser for the compressed layer contents.
func (u *uncompressedLayer) Compressed() (io.ReadCloser, error) {
	// this is the trick. we return Uncompressed() because there is no compression
	return u.wrapped.Uncompressed()
}

// Uncompressed returns an io.ReadCloser for the uncompressed layer contents.
func (u *uncompressedLayer) Uncompressed() (io.ReadCloser, error) {
	return u.wrapped.Uncompressed()
}

// Size returns the compressed size of the Layer.
func (u *uncompressedLayer) Size() (int64, error) {
	return u.size, nil
}

// MediaType returns the media type of the Layer.
func (_ *uncompressedLayer) MediaType() (types.MediaType, error) {
	return types.DockerUncompressedLayer, nil
}

func newUncompressedLayer(orig v1.Layer) (v1.Layer, error) {
	layer := &uncompressedLayer{
		wrapped: orig,
	}

	var err error
	if layer.digest, layer.size, err = computeDigest(orig.Uncompressed); err != nil {
		return nil, err
	}

	return layer, nil
}

func computeDigest(opener tarball.Opener) (v1.Hash, int64, error) {
	rc, err := opener()
	if err != nil {
		return v1.Hash{}, 0, err
	}
	defer rc.Close()

	return v1.SHA256(rc)
}

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

	// we want nonEmptyHistory to be the v1.History entries that are not EmptyLayer
	// it should have the same indexes as layers
	nonEmptyHistory := make([]v1.History, 0, len(layers))
	for _, h := range ocf.History {
		if !h.EmptyLayer {
			nonEmptyHistory = append(nonEmptyHistory, h)
		}
	}
	if len(nonEmptyHistory) != len(layers) {
		return nil, fmt.Errorf("number of non-empty history entries (%d) is different from number of layers (%d)", len(nonEmptyHistory), len(layers))
	}

	var historyIdx, addendumIdx int
	for layerIdx := 0; layerIdx < len(layers); addendumIdx, layerIdx = addendumIdx+1, layerIdx+1 {
		startLayer := time.Now()
		compressedSize, err := layers[layerIdx].Size()
		if err != nil {
			return nil, fmt.Errorf("getting compressed size: %w", err)
		}
		fmt.Fprintln(os.Stderr, "uncompressing layer", layerIdx, compressedSize, "created by", nonEmptyHistory[layerIdx].CreatedBy, "with size")
		newLayer, err := decompressLayer(layers[layerIdx])
		if err != nil {
			return nil, fmt.Errorf("setting uncompressed layer: %w", err)
		}
		// truncate time to 3 decimal places
		truncTime := time.Duration(int64(time.Since(startLayer).Seconds()*1000)) * time.Millisecond
		fmt.Fprintln(os.Stderr, "uncompressing layer", layerIdx, "took", truncTime)
		// uncompressedSize, err := newLayer.Size()
		if err != nil {
			return nil, fmt.Errorf("getting uncompressed size: %w", err)
		}
		// compressionRatio := int(float64(compressedSize)/float64(uncompressedSize)*100) / 100.0
		// fmt.Fprintln(os.Stderr,
		// 	"compression ratio for layer",
		// 	layerIdx,
		// 	"is", compressionRatio, "(", compressedSize, "/", uncompressedSize,
		// 	"created by", nonEmptyHistory[layerIdx].CreatedBy, ")",
		// )

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

func getSize(layer v1.Layer) int64 {
	size, err := layer.Size()
	if err != nil {
		panic(err)
	}
	return size
}

func decompressLayer(layer v1.Layer) (v1.Layer, error) {
	uncompLayer, err := newUncompressedLayer(layer)
	if err != nil {
		return nil, fmt.Errorf("creating new layer: %w", err)
	}
	if os.Getenv("NO_COMPRESSION") != "" {
		return uncompLayer, nil
	}
	zstdLayer, err := tarball.LayerFromOpener(uncompLayer.Uncompressed,
		tarball.WithCompression(compression.ZStd),
		tarball.WithMediaType(types.OCILayerZStd),
		// compression levels:
		// https://github.com/klauspost/compress/blob/master/zstd/encoder_options.go#L196
		// zstd technically goes up to 22 though
		tarball.WithCompressionLevel(11),
	)
	if err != nil {
		return nil, fmt.Errorf("creating new layer: %w", err)
	}
	prevRatio := float64(getSize(layer)) / float64(getSize(uncompLayer))
	compRatio := float64(getSize(zstdLayer)) / float64(getSize(uncompLayer))
	fmt.Fprintln(os.Stderr, "compression ratio for layer is", compRatio, "(", prevRatio, "->", compRatio, ")")
	if compRatio < 0.9 {
		fmt.Fprintln(os.Stderr, "using zstd compression")
		return zstdLayer, nil
	}
	return uncompLayer, nil
}
