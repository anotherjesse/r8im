package cli

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"

	"github.com/anotherjesse/r8im/pkg/auth"
	"github.com/anotherjesse/r8im/pkg/images"
)

var (
	weightsRef string
)

func newRemixCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "remix",
		Short:  "remix existing layers of an existing image",
		Hidden: false,

		RunE: remixCommmand,
	}

	cmd.Flags().StringVarP(&sToken, "token", "t", "", "replicate cog token")
	cmd.Flags().StringVarP(&sRegistry, "registry", "r", "r8.im", "registry host")

	cmd.Flags().StringVarP(&baseRef, "base", "b", "", "base image reference - include tag: r8.im/username/modelname@sha256:hexdigest")
	cmd.MarkFlagRequired("base")
	cmd.Flags().StringVarP(&weightsRef, "weights", "w", "", "weights image reference - include tag: r8.im/username/weights@sha256:hexdigest")
	cmd.MarkFlagRequired("weights")
	cmd.Flags().StringVarP(&dest, "dest", "d", "", "destination image reference: r8.im/username/modelname")
	cmd.MarkFlagRequired("dest")

	return cmd
}

func remixCommmand(cmd *cobra.Command, args []string) error {
	if sToken == "" {
		sToken = os.Getenv("COG_TOKEN")
	}

	u, err := auth.VerifyCogToken(sRegistry, sToken)
	if err != nil {
		fmt.Fprintln(os.Stderr, "authentication error, invalid token or registry host error")
		return err
	}
	auth := authn.FromConfig(authn.AuthConfig{Username: u, Password: sToken})

	fmt.Println("remix time")
	url, err := images.ReallyRemix(baseRef, weightsRef, dest, auth)

	fmt.Println(url)

	if err != nil {
		return err
	}

	return nil
}
