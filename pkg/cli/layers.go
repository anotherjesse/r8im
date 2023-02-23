package cli

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"

	"github.com/anotherjesse/r8im/pkg/auth"
	"github.com/anotherjesse/r8im/pkg/images"
)

func newLayerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "layers [image]",
		Short:  "list layers of an existing image",
		Hidden: false,

		RunE: layersCommmand,
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&sToken, "token", "t", "", "replicate cog token")

	return cmd
}

func layersCommmand(cmd *cobra.Command, args []string) error {
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

	for _, layer := range layers {
		fmt.Println(layer)
	}

	return nil
}
