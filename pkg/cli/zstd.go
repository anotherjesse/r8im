package cli

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"

	"github.com/anotherjesse/r8im/pkg/auth"
	"github.com/anotherjesse/r8im/pkg/images"
)

func newZstdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "zstd <image> <dest>",
		Short:  "recompress layers of an existing image using zstd, pushing result to dest",
		Hidden: false,

		RunE: zstdCommmand,
		Args: cobra.ExactArgs(2),
	}

	cmd.Flags().StringVarP(&sToken, "token", "t", "", "replicate cog token")

	return cmd
}

func zstdCommmand(cmd *cobra.Command, args []string) error {
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

	digest, err := images.Zstd(imageName, dest, auth)
	if err != nil {
		return err
	}

	fmt.Println(digest)

	return nil
}
