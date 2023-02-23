package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() (*cobra.Command, error) {
	rootCmd := cobra.Command{
		Use:           "r8im",
		Short:         "replicate.com helpers",
		Version:       "0.0.1",
		SilenceErrors: true,
	}

	rootCmd.AddCommand(
		newAffixCommand(),
		newLayerCommand(),
		newRemixCommand(),
	)

	return &rootCmd, nil
}
