package cmd

import (
	"fmt"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/utils"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:          "login",
	Short:        "A brief description of your command",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		setCommandStreamsOnPrompt := func(prompt *promptui.Prompt) *promptui.Prompt {
			prompt.Stdin = utils.NoopReadCloser(cmd.InOrStdin())
			prompt.Stdout = utils.NoopWriteCloser(cmd.OutOrStdout())

			return prompt
		}

		// TODO: extract this common code for gathering configuring into a seperate package
		// Some of the code is in conjur-api-go and needs to be made configurable so that you can pass a custom path to .conjurrc
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return err
		}

		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return err
		}

		config, err := conjurapi.LoadConfig()
		if err != nil {
			return err
		}

		if config.ApplianceURL == "" {
			return fmt.Errorf("%s", "Missing required configuration for Conjur API URL")
		}

		if config.Account == "" {
			return fmt.Errorf("%s", "Missing required configuation for Conjur account")
		}

		// TODO: I should be able to create a client and unauthenticated client
		conjurClient, err := conjurapi.NewClient(config)
		if err != nil {
			return err
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		if verbose {
			clients.VerboseLoggingForClient(cmd, conjurClient)
		}

		_, err = clients.LoginWithOptionalPrompt(setCommandStreamsOnPrompt, conjurClient, username, password)
		if err != nil {
			return err
		}

		cmd.Println("Logged in")

		// TODO: should we be storing this in .netrc for future use ?
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringP("username", "u", "", "")
	loginCmd.Flags().StringP("password", "p", "", "")
}
