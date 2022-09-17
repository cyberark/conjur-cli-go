package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/cyberark/conjur-cli/pkg/conjurrc"
	"github.com/cyberark/conjur-cli/pkg/utils"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func runInitCommand(cmd *cobra.Command, args []string) error {
	var err error

	account := cmd.Flag("account").Value.String()
	applianceURL := cmd.Flag("url").Value.String()
	filePath := cmd.Flag("file").Value.String()

	setCommandStreamsOnPrompt := func(prompt *promptui.Prompt) {
		prompt.Stdin = utils.NopReadCloser(cmd.InOrStdin())
		prompt.Stdout = utils.NopWriteCloser(cmd.OutOrStdout())
	}

	if len(applianceURL) == 0 {
		prompt := promptui.Prompt{
			Label: "Enter the URL of your Conjur service",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("URL is required")
				}

				_, err := url.ParseRequestURI(input)
				return err
			},
		}
		setCommandStreamsOnPrompt(&prompt)

		applianceURL, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if len(account) == 0 {
		prompt := promptui.Prompt{
			Label: "Enter your organization account name",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("Account is required")
				}
				return nil
			},
		}
		setCommandStreamsOnPrompt(&prompt)

		account, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	err = conjurrc.WriteConjurrc(account, applianceURL, filePath, func(filePath string) error {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("File %s exists. Overwrite", filePath),
			IsConfirm: true,
		}
		setCommandStreamsOnPrompt(&prompt)

		_, err := prompt.Run()
		if err != nil {
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

	// TODO: figure out what to do when getting the user home directory returns an error
	userHomeDir, _ := os.UserHomeDir()

	cmd.PersistentFlags().StringP("account", "a", "", "Conjur organization account name")
	cmd.PersistentFlags().StringP("url", "u", "", "URL of the Conjur service")
	cmd.PersistentFlags().BoolP("help", "h", false, "Help for init command") // TODO: maybe change this for everything
	cmd.PersistentFlags().StringP("certificate", "c", "", "Conjur SSL certificate (will be obtained from host unless provided by this option)")
	cmd.PersistentFlags().StringP("file", "f", filepath.Join(userHomeDir, ".conjurrc"), "File to write the configuration to")
	cmd.PersistentFlags().Bool("force", false, "Force overwrite of existing file")

	return cmd
}

func init() {
	initCmd := NewInitCommand()
	rootCmd.AddCommand(initCmd)
}
