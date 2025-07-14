package cmd

import (
	"context"
	"errors"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/cmd/style"
	"github.com/cyberark/conjur-cli-go/pkg/version"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	disableCompletion := false
	config := clients.LoadConfigOrDefault()
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
	err := style.Execute(rootCmd)
	if err != nil {
		// check if error is about x509 certificate
		if strings.Contains(err.Error(), "x509:") {
			rootCmd.PrintErrln(
				"Your Conjur server's certificate is not trusted. Consider running 'conjur init' to initialize the CLI with your Conjur server's certificate.")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			rootCmd.PrintErrln(
				"Your request has timed out. If your operation is expected to be long-running, please consider increasing the HTTP timeout. For details, please refer to the command help.")
		}
		os.Exit(1)
	}
}

var rootCmd = newRootCommand()
