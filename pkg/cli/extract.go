package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"

	"github.com/anotherjesse/r8im/pkg/auth"
	"github.com/anotherjesse/r8im/pkg/images"
	r8Layers "github.com/anotherjesse/r8im/pkg/layers"
)

func newExtractCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "extract <image> [--output file]",
		Short:  "extract weights from image",
		Hidden: false,

		RunE: extractCommand,
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&sToken, "token", "t", "", "replicate cog token")
	cmd.Flags().StringVarP(&dest, "output", "o", "", "destination tar file")

	return cmd
}

func extractCommand(cmd *cobra.Command, args []string) error {
	if sToken == "" {
		sToken = os.Getenv("COG_TOKEN")
	}

	if len(args) == 0 {
		return nil
	}

	u, err := auth.VerifyCogToken(sRegistry, sToken)
	if err != nil {
		fmt.Fprintln(os.Stderr, "authentication error, invalid token or registry host error")
		return err
	}
	auth := authn.FromConfig(authn.AuthConfig{Username: u, Password: sToken})

	imageName := args[0]
	layers, err := images.Layers(imageName, auth)
	if err != nil {
		return err
	}

	weightsFound := false

	for _, layer := range layers {
		if strings.HasSuffix(layer.Command, " # weights") {
			l := layer.Raw
			rc, err := l.Uncompressed()
			if err != nil {
				return err
			}
			weightsFound, err = r8Layers.ExtractTarWithoutPrefixAndIgnoreWhiteout(rc, dest)
			if err != nil {
				return err
			}
			if weightsFound {
				return nil
			}
		}
	}

	for _, layer := range layers {
		if strings.HasPrefix(layer.Command, "COPY . /src") {
			l := layer.Raw
			rc, err := l.Uncompressed()
			if err != nil {
				return err
			}
			weightsFound, err = r8Layers.ExtractTarWithoutPrefixAndIgnoreWhiteout(rc, dest)
			if err != nil {
				return err
			}
			if weightsFound {
				return nil
			}
		}
	}

	if !weightsFound {
		return fmt.Errorf("no weights found")
	}

	return nil
}
