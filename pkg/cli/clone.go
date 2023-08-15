package cli

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"

	"github.com/anotherjesse/r8im/pkg/auth"
	"github.com/anotherjesse/r8im/pkg/images"
)

func newCloneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "clone",
		Short:  "copy existing image to a new image",
		Hidden: false,
		RunE:   cloneCommmand,
	}

	cmd.Flags().StringVarP(&sToken, "token", "t", "", "replicate cog token")
	cmd.Flags().StringVarP(&sRegistry, "registry", "r", "r8.im", "registry host")
	cmd.Flags().StringVarP(&baseRef, "base", "b", "", "base image reference - include tag: r8.im/username/modelname@sha256:hexdigest")
	cmd.MarkFlagRequired("base")
	cmd.Flags().StringVarP(&dest, "dest", "d", "", "destination image reference: r8.im/username/modelname")
	cmd.MarkFlagRequired("dest")

	return cmd
}

func cloneCommmand(cmd *cobra.Command, args []string) error {
	if sToken == "" {
		sToken = os.Getenv("COG_TOKEN")
	}

	u, err := auth.VerifyCogToken(sRegistry, sToken)
	if err != nil {
		fmt.Fprintln(os.Stderr, "authentication error, invalid token or registry host error")
		return err
	}
	auth := authn.FromConfig(authn.AuthConfig{Username: u, Password: sToken})

	image_id, err := images.Affix(baseRef, dest, "", auth)
	if err != nil {
		return err
	}

	fmt.Println(image_id)

	return nil
}
