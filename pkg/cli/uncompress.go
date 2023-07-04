package cli

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"

	"github.com/anotherjesse/r8im/pkg/auth"
	"github.com/anotherjesse/r8im/pkg/images"
)

func newUncompressCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "uncompress <image> <dest>",
		Short:  "uncompress layers of an existing image, pushing result to dest",
		Hidden: false,

		RunE: uncompressCommmand,
		Args: cobra.ExactArgs(2),
	}

	cmd.Flags().StringVarP(&sToken, "token", "t", "", "replicate cog token")

	return cmd
}

func uncompressCommmand(cmd *cobra.Command, args []string) error {
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
	dest := args[1]

	digest, err := images.Uncompress(imageName, dest, auth)
	if err != nil {
		return err
	}

	fmt.Println(digest)

	return nil
}
