package cli

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"

	"github.com/replicate/r8/pkg/auth"
	"github.com/replicate/r8/pkg/images"
)

// var (
// 	sToken string
// )

func newRemixCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "remix",
		Short:  "remix existing layers of an existing image",
		Hidden: false,

		RunE: remixCommmand,
	}

	cmd.Flags().StringVarP(&sToken, "token", "t", "", "replicate cog token")

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
	err = images.Remix(auth)
	if err != nil {
		return err
	}

	return nil
}
