package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() (*cobra.Command, error) {
	rootCmd := cobra.Command{
		Use:   "r8",
		Short: "replicate.com helpers",
		Example: `   To run a command inside a Docker environment defined with Cog:
      $ cog run echo hello world`,
		Version:       "0.0.1",
		SilenceErrors: true,
	}

	rootCmd.AddCommand(
		newAffixCommand(),
	)

	return &rootCmd, nil
}
