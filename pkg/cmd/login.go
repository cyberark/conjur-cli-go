package cmd

import (
	"fmt"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-cli-go/pkg/clients"

	"github.com/spf13/cobra"
)

type loginCmdFuncs struct {
	LoadAndValidateConjurConfig func() (conjurapi.Config, error)
	LoginWithPromptFallback     func(client clients.ConjurClient, username string, password string) (*authn.LoginPair, error)
	OidcLogin                   func(conjurClient clients.ConjurClient, username string, password string) (clients.ConjurClient, error)
}

var defaultLoginCmdFuncs = loginCmdFuncs{
	LoadAndValidateConjurConfig: clients.LoadAndValidateConjurConfig,
	LoginWithPromptFallback:     clients.LoginWithPromptFallback,
	OidcLogin:                   clients.OidcLogin,
}

type loginCmdFlagValues struct {
	identity string
	password string
	debug    bool
}

func getLoginCmdFlagValues(cmd *cobra.Command) (loginCmdFlagValues, error) {
	// TODO: extract this common code for gathering configuring into a seperate package
	// Some of the code is in conjur-api-go and needs to be made configurable so that you can pass a custom path to .conjurrc
	identity, err := cmd.Flags().GetString("id")
	if err != nil {
		return loginCmdFlagValues{}, err
	}

	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return loginCmdFlagValues{}, err
	}

	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return loginCmdFlagValues{}, err
	}

	return loginCmdFlagValues{
		identity: identity,
		password: password,
		debug:    debug,
	}, nil
}

func newLoginCmd(funcs loginCmdFuncs) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Conjur using the provided identity and password",
		Long: `Authenticate with Conjur using the provided identity and password.

The command will prompt for identity and password if they are not provided via flags.

On successful login, the password is exchanged for the user's API key, which is cached in the operating system user's credential storage or .netrc file. Subsequent commands will authenticate using the cached credentials. To switch users, login again using new credentials. To erase credentials, use the 'logout' command.

Examples:

- conjur login -i alice -p My$ecretPass
- conjur login`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			cmdFlagVals, err := getLoginCmdFlagValues(cmd)
			if err != nil {
				return err
			}

			config, err := funcs.LoadAndValidateConjurConfig()
			if err != nil {
				return err
			}

			// TODO: I should be able to create a client and unauthenticated client
			conjurClient, err := conjurapi.NewClient(config)
			if err != nil {
				return err
			}

			if cmdFlagVals.debug {
				clients.MaybeDebugLoggingForClient(cmdFlagVals.debug, cmd, conjurClient)
			}

			if config.AuthnType == "" || config.AuthnType == "authn" || config.AuthnType == "ldap" {
				_, err = funcs.LoginWithPromptFallback(conjurClient, cmdFlagVals.identity, cmdFlagVals.password)
			} else if config.AuthnType == "oidc" {
				_, err = funcs.OidcLogin(conjurClient, cmdFlagVals.identity, cmdFlagVals.password)
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

	cmd.Flags().StringP("id", "i", "", "")
	cmd.Flags().StringP("password", "p", "", "")

	return cmd
}

func init() {
	loginCmd := newLoginCmd(defaultLoginCmdFuncs)
	rootCmd.AddCommand(loginCmd)
}
