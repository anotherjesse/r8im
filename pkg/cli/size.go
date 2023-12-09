package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"

	"github.com/anotherjesse/r8im/pkg/auth"
	"github.com/anotherjesse/r8im/pkg/images"
)

func newSizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "size",
		Short:  "calculate size of existing image",
		Hidden: false,
		Args:   cobra.ExactArgs(1),
		RunE:   sizeCommand,
	}

	cmd.Flags().StringVarP(&sToken, "token", "t", "", "replicate cog token")
	cmd.Flags().StringVarP(&sRegistry, "registry", "r", "r8.im", "registry host")

	return cmd
}

func sizeCommand(cmd *cobra.Command, args []string) error {
	if sToken == "" {
		sToken = os.Getenv("REPLICATE_API_TOKEN")
	}
	if sToken == "" {
		sToken = os.Getenv("COG_TOKEN")
	}

	u, err := auth.VerifyCogToken(sRegistry, sToken)
	if err != nil {
		fmt.Fprintln(os.Stderr, "authentication error, invalid token or registry host error")
		return err
	}
	auth := authn.FromConfig(authn.AuthConfig{Username: u, Password: sToken})

	all_images, err := getR8Urls(args[0], auth)
	if err != nil {
		return err
	}

	uniqLayers := make(map[string]int64)
	totalSize := 0

	for _, image := range all_images {
		imageSize := 0
		imageUniqSize := 0
		layers, err := images.Layers(image, auth)
		if err != nil {
			return err
		}

		for _, layer := range layers {
			imageSize += int(layer.Size)
			if _, ok := uniqLayers[layer.Digest]; ok {
				continue
			}
			uniqLayers[layer.Digest] = layer.Size
			imageUniqSize += int(layer.Size)
		}

		totalSize += imageUniqSize
		version := strings.Split(image, "@sha256:")[1]
		fmt.Printf("%s\t%s\t%s\n", version, humanize.Bytes(uint64(imageSize)), humanize.Bytes(uint64(imageUniqSize)))
	}

	fmt.Printf("Total Size: %s\n", humanize.Bytes(uint64(totalSize)))

	return nil
}

func ensureRegistry(baseRef string) string {
	if !strings.Contains(baseRef, sRegistry) {
		return sRegistry + "/" + baseRef
	}
	return baseRef
}

func getReplicateClient(session authn.Authenticator) (*replicate.Client, error) {
	a, err := session.Authorization()
	if err != nil {
		return nil, err
	}
	token := a.Password
	client, err := replicate.NewClient(replicate.WithToken(token))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getVersions(owner string, model string, session authn.Authenticator) ([]string, error) {
	client, err := getReplicateClient(session)
	if err != nil {
		return nil, err
	}

	resp, err := client.ListModelVersions(context.Background(), owner, model)
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, version := range resp.Results {
		versions = append(versions, version.ID)
	}

	return versions, nil
}

func getR8Urls(baseRef string, session authn.Authenticator) ([]string, error) {
	baseRef = ensureRegistry(baseRef)
	parts := strings.Split(baseRef, "/")
	owner := parts[len(parts)-2]
	model := parts[len(parts)-1]
	version := ""
	if strings.Contains(model, ":") {
		version = strings.Split(model, ":")[1]
		model = strings.Split(model, ":")[0]
	}
	if strings.Contains(model, "@") {
		model = strings.Split(model, "@")[0]
	}

	if version != "" {
		image := fmt.Sprintf("%s/%s/%s@sha256:%s", sRegistry, owner, model, version)
		return []string{image}, nil
	}

	versions, err := getVersions(owner, model, session)
	if err != nil {
		return nil, err
	}

	var all_images []string
	for _, version := range versions {
		image := fmt.Sprintf("%s/%s/%s@sha256:%s", sRegistry, owner, model, version)
		all_images = append(all_images, image)
	}

	return all_images, nil
}
