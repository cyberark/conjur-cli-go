package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cyberark/conjur-cli-go/pkg/conjurrc"
	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// TODO: whenever this is called we should store to .conjurrc
func askForConnectionDetails(decoratePrompt decoratePromptFunc, account string, applianceURL string) (string, string, error) {
	var err error

	if len(applianceURL) == 0 {
		prompt := decoratePrompt(newApplianceURLPrompt())
		applianceURL, err = runPrompt(prompt)

		if err != nil {
			return "", "", err
		}
	}

	if len(account) == 0 {
		prompt := decoratePrompt(newAccountPrompt())
		account, err = runPrompt(prompt)

		if err != nil {
			return "", "", err
		}
	}

	return account, applianceURL, err
}

func runInitCommand(cmd *cobra.Command, args []string) error {
	var err error

	setCommandStreamsOnPrompt := func(prompt *promptui.Prompt) *promptui.Prompt {
		prompt.Stdin = utils.NoopReadCloser(cmd.InOrStdin())
		prompt.Stdout = utils.NoopWriteCloser(cmd.OutOrStdout())

		return prompt
	}

	account := cmd.Flag("account").Value.String()
	applianceURL := cmd.Flag("url").Value.String()
	filePath := cmd.Flag("file").Value.String()

	account, applianceURL, err = askForConnectionDetails(setCommandStreamsOnPrompt, account, applianceURL)
	if err != nil {
		return err
	}

	err = conjurrc.WriteConjurrc(account, applianceURL, filePath, func(filePath string) error {
		prompt := setCommandStreamsOnPrompt(newFileExistsPrompt(filePath))
		_, err = runPrompt(prompt)

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

	cmd.Printf("Wrote configuration to %s\n", filePath)
	return nil
}

// NewInitCommand initializes and configures the 'conjur init' command.
func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "init",
		Short:        "Initialize the Conjur configuration",
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
