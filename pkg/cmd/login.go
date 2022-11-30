package cmd

import (
	"fmt"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/pkg/clients"
	"github.com/cyberark/conjur-cli-go/pkg/prompts"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Conjur using the provided username and password.",
	Long: `Authenticate with Conjur using the provided username and password.

The command will prompt for username and password if they are not provided via flag.

On successful login, the password is exchanged for the user's API key, which is cached in the operating system user's .netrc file. Subsequent commands will authenticate using the cached credentials. To switch users, login again using new credentials. To erase credentials, use the 'logout' command.

Examples:

- conjur authn login -u alice -p My$ecretPass
- conjur authn login`,
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
