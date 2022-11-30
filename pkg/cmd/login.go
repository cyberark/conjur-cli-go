package cmd

import (
	"fmt"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:          "login",
	Short:        "A brief description of your command",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

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

		config, err := clients.LoadAndValidateConjurConfig()
		if err != nil {
			return err
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
			clients.MaybeVerboseLoggingForClient(verbose, cmd, conjurClient)
		}

		if config.AuthnType == "" || config.AuthnType == "authn" || config.AuthnType == "ldap" {
			decoratePrompt := prompts.PromptDecoratorForCommand(cmd)
			_, err = clients.LoginWithPromptFallback(decoratePrompt, conjurClient, username, password)
		} else if config.AuthnType == "oidc" {
			_, err = clients.OidcLogin(conjurClient)
		} else {
			return fmt.Errorf("unsupported authentication type: %s", config.AuthnType)
		}
		if err != nil {
			return err
		}

		cmd.Println("Logged in")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringP("username", "u", "", "")
	loginCmd.Flags().StringP("password", "p", "", "")
}
