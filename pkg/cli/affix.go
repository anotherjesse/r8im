package cli

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"

	"github.com/replicate/r8/pkg/auth"
	"github.com/replicate/r8/pkg/images"
)

var (
	sToken    string
	sRegistry string
	baseRef   string
	dest      string
	tar       string
)

func newAffixCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "affix",
		Short:  "add a new layer to an existing image",
		Hidden: false,
		RunE:   affixCommmand,
	}

	cmd.Flags().StringVarP(&sToken, "token", "t", "", "replicate cog token")
	cmd.Flags().StringVarP(&sRegistry, "registry", "r", "r8.im", "registry host")
	cmd.Flags().StringVarP(&baseRef, "base", "b", "", "base image reference - include tag: r8.im/username/modelname@sha256:hexdigest")
	cmd.MarkFlagRequired("base")
	cmd.Flags().StringVarP(&dest, "dest", "d", "", "destination image reference: r8.im/username/modelname")
	cmd.MarkFlagRequired("dest")
	cmd.Flags().StringVarP(&tar, "tar", "f", "", "tar file to append as new layer")
	cmd.MarkFlagRequired("tar")
	cmd.MarkFlagFilename("tar", "tar", "tar.gz", "tgz")

	return cmd
}

func affixCommmand(cmd *cobra.Command, args []string) error {
	if sToken == "" {
		sToken = os.Getenv("COG_TOKEN")
	}

	u, err := auth.VerifyCogToken(sRegistry, sToken)
	if err != nil {
		fmt.Fprintln(os.Stderr, "authentication error, invalid token or registry host error")
		return err
	}
	auth := authn.FromConfig(authn.AuthConfig{Username: u, Password: sToken})

	image_id, err := images.Affix(baseRef, dest, tar, auth)
	if err != nil {
		return err
	}

	fmt.Println(image_id)

	return nil
}
