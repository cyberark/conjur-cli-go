package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "conjur",
		Short: "Conjur CLI",
		Long:  "Command-line toolkit for managing Conjur resources and performing common tasks.",
	}

	rootCmd.PersistentFlags().Bool("debug", false, "Debug logging enabled")
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var rootCmd = newRootCommand()
