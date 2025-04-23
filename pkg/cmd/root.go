package cmd

import (
	"context"
	"errors"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"os"
	"time"

	"github.com/cyberark/conjur-cli-go/pkg/version"
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	disableCompletion := false
	config, _ := clients.LoadAndValidateConjurConfig(0)
	if config.IsConjurCloud() {
		disableCompletion = true
	}

	rootCmd := &cobra.Command{
		Use:               "conjur",
		Short:             "Conjur CLI",
		Long:              "Command-line toolkit for managing Conjur resources and performing common tasks.",
		Version:           version.FullVersionName,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: disableCompletion},
	}

	if config.IsConjurCE() || config.IsConjurOSS() {
		rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug logging enabled")
		rootCmd.PersistentFlags().Duration("timeout", time.Minute, "HTTP timeout duration, between 1s and 10m")
	}
	rootCmd.SetVersionTemplate("Conjur CLI version {{.Version}}\n")
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	err := rootCmd.Execute()
	if errors.Is(err, context.DeadlineExceeded) {
		rootCmd.PrintErrln(
			"Your request has timed out. If your operation is expected to be long-running, please consider increasing the HTTP timeout. For details, please refer to the command help.")
	}
	if err != nil {
		os.Exit(1)
	}
}

var rootCmd = newRootCommand()
