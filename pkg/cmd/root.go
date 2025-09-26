package cmd

import (
	"context"
	"crypto/x509"
	"errors"
	"os"
	"time"

	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/cmd/style"
	"github.com/cyberark/conjur-cli-go/pkg/version"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	disableCompletion := false
	config := clients.LoadConfigOrDefault()
	if config.IsSaaS() {
		disableCompletion = true
	}

	rootCmd := &cobra.Command{
		Use:               "conjur",
		Short:             "Secrets Manager CLI",
		Long:              "Command-line toolkit for managing Secrets Manager resources and performing common tasks.",
		Version:           version.FullVersionName,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: disableCompletion},
	}

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug logging enabled")
	rootCmd.PersistentFlags().Duration("timeout", time.Minute, "HTTP timeout duration, between 1s and 10m")
	rootCmd.SetVersionTemplate("Secrets Manager CLI version {{.Version}}\n")
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	err := style.Execute(rootCmd)
	if err != nil {
		// check if error is about x509 unknown certificate signing authority
		if errors.As(err, &x509.UnknownAuthorityError{}) {
			rootCmd.PrintErrln(
				"WARNING: The server’s TLS certificate has changed.\n\n" +
					"This means one of the following:\n\n" +
					"    The server's certificate has been rotated\n" +
					"    The server or your device may have been compromised\n\n" +
					"If needed, verify the change with your administrator before reinitializing the CLI and accepting the new certificate.")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			rootCmd.PrintErrln(
				"Your request has timed out. If your operation is expected to be long-running, please consider increasing the HTTP timeout. For details, please refer to the command help.")
		}
		os.Exit(1)
	}
}

var rootCmd = newRootCommand()
