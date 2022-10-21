package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cyberark/conjur-cli-go/pkg/conjurrc"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"

	"github.com/spf13/cobra"
)

func getInitCmdFlagValues(cmd *cobra.Command) (string, string, string, bool, error) {
	account, err := cmd.Flags().GetString("account")
	if err != nil {
		return "", "", "", false, err
	}
	applianceURL, err := cmd.Flags().GetString("url")
	if err != nil {
		return "", "", "", false, err
	}
	conjurrcFilePath, err := cmd.Flags().GetString("file")
	if err != nil {
		return "", "", "", false, err
	}
	forceFileOverwrite, err := cmd.Flags().GetBool("force")
	if err != nil {
		return "", "", "", false, err
	}

	return account, applianceURL, conjurrcFilePath, forceFileOverwrite, nil
}

func runInitCommand(cmd *cobra.Command, args []string) error {
	var err error

	account, applianceURL, conjurrcFilePath, forceFileOverwrite, err := getInitCmdFlagValues(cmd)
	if err != nil {
		return err
	}

	setCommandStreamsOnPrompt := prompts.PromptDecoratorForCommand(cmd)

	account, applianceURL, err = prompts.MaybeAskForConnectionDetails(
		setCommandStreamsOnPrompt,
		account,
		applianceURL,
	)
	if err != nil {
		return err
	}

	err = conjurrc.WriteConjurrc(
		account,
		applianceURL,
		conjurrcFilePath,
		func(filePath string) error {
			if forceFileOverwrite {
				return nil
			}

			err := prompts.AskToOverwriteFile(setCommandStreamsOnPrompt, filePath)
			if err != nil {
				// TODO: make all the errors lowercase to make Go static check happy, then have something higher up that capitalizes the first letter
				// of errors from commands
				return fmt.Errorf("Not overwriting %s", filePath)
			}

			return nil
		})
	if err != nil {
		return err
	}

	cmd.Printf("Wrote configuration to %s\n", conjurrcFilePath)
	return nil
}

// NewInitCommand initializes and configures the 'conjur init' command.
func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Use the init command to initialize the Conjur CLI with a Conjur endpoint.",
		Long: `Use the init command to initialize the Conjur CLI with a Conjur endpoint.

The init command creates a configuration file (.conjurrc) that contains the details for connecting to Conjur. This file is located under the user's root directory.`,
		SilenceUsage: true,
		RunE:         runInitCommand,
	}

	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cmd.PersistentFlags().StringP("account", "a", "", "Conjur organization account name")
	cmd.PersistentFlags().StringP("url", "u", "", "URL of the Conjur service")
	cmd.PersistentFlags().StringP("certificate", "c", "", "Conjur SSL certificate (will be obtained from host unless provided by this option)")
	cmd.PersistentFlags().StringP("file", "f", filepath.Join(userHomeDir, ".conjurrc"), "File to write the configuration to")
	cmd.PersistentFlags().Bool("force", false, "Force overwrite of existing file")

	return cmd
}

func init() {
	initCmd := NewInitCommand()
	rootCmd.AddCommand(initCmd)
}
