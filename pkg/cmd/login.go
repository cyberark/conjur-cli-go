package cmd

import (
	"fmt"
	"time"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"
	"github.com/cyberark/conjur-cli-go/pkg/clients"

	"github.com/spf13/cobra"
)

type loginCmdFuncs struct {
	LoadAndValidateConjurConfig func(timeout time.Duration) (conjurapi.Config, error)
	LoginWithPromptFallback     func(client clients.ConjurClient, username string, password string) (*authn.LoginPair, error)
	OidcLogin                   func(conjurClient clients.ConjurClient, username string, password string) (clients.ConjurClient, error)
	JWTAuthenticate             func(conjurClient clients.ConjurClient) error
}

var defaultLoginCmdFuncs = loginCmdFuncs{
	LoadAndValidateConjurConfig: clients.LoadAndValidateConjurConfig,
	LoginWithPromptFallback:     clients.LoginWithPromptFallback,
	OidcLogin:                   clients.OidcLogin,
	JWTAuthenticate:             clients.JWTAuthenticate,
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

	config, _ := clients.LoadAndValidateConjurConfig(0)
	var debug bool
	if config.IsConjurCE() || config.IsConjurOSS() {
		debug, err = cmd.Flags().GetBool("debug")
	}
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

			timeout, err := clients.GetTimeout(cmd)
			if err != nil {
				return err
			}

			config, err := funcs.LoadAndValidateConjurConfig(timeout)
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
			} else if config.AuthnType == "jwt" {
				// We have to recreate the client with the JWT method so it
				// attaches a JWTAuthenticator to the client otherwise
				// conjurClient.GetAuthenticator() will return nil
				conjurClient, err = conjurapi.NewClientFromJwt(config)
				if err != nil {
					return err
				}
				// Just run authenticate to validate the jwt. This isn't
				// necessary (since the JWT path is set in the `init` command)
				// but is provided as a convenience to the user, to allow them
				// to validate their jwt file before running other commands.
				err = funcs.JWTAuthenticate(conjurClient)
				if err != nil {
					err = fmt.Errorf("Unable to authenticate with Conjur using the provided JWT file: %s", err)
				}
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

	cmd.Flags().StringP("id", "i", "", "The identity to authenticate with. For hosts: 'host/<full path>'.")
	cmd.Flags().StringP("password", "p", "", "Password or API key for the specified identity.")

	return cmd
}

func init() {
	loginCmd := newLoginCmd(defaultLoginCmdFuncs)
	rootCmd.AddCommand(loginCmd)
}
