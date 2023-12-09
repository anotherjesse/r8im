package cli

import (
	"log"
	"os"

	"github.com/google/go-containerregistry/pkg/logs"
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
		newCloneCommand(),
		newExtractCommand(),
		newLayerCommand(),
		newRemixCommand(),
		newSizeCommand(),
		newZstdCommand(),
	)
	logs.Warn = log.New(os.Stderr, "gcr WARN: ", log.LstdFlags)
	logs.Progress = log.New(os.Stderr, "gcr: ", log.LstdFlags)

	return &rootCmd, nil
}
