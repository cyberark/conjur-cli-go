package main

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/urfave/cli"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-api-go/conjurapi/authn"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
)

// AuthnCommands contains the definitions of the commands related to authentication.
var AuthnCommands = func(initFunc cli.BeforeFunc, fs afero.Fs) []cli.Command {
	var (
		authenticateOptions cmd.AuthnAuthenticatorOptions
		loginOptions        cmd.AuthnLoginOptions
	)
	getClient := func(ctx *cli.Context) cmd.AuthnClient {
		return ctx.App.Metadata["client"].(cmd.AuthnClient)
	}

	return []cli.Command{
		{
			Name:  "authn",
			Usage: "Login and logout",
			Subcommands: []cli.Command{
				{
					Name:  "authenticate",
					Usage: "Obtains an authentication token using the current logged-in user",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:        "header, H",
							Usage:       "Base64 encode the result and format as an HTTP Authorization header",
							Destination: &authenticateOptions.AsHeader,
						},
					},
					Before: initFunc,
					Action: func(c *cli.Context) (err error) {
						authenticator := cmd.NewAuthnAuthenticator(getClient(c))

						resp, err := authenticator.Do(authenticateOptions)
						if err != nil {
							return
						}
						fmt.Println(string(resp))
						return nil
					},
				},
				{
					Name:  "login",
					Usage: "Logs in and caches credentials to netrc",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "username, u",
							Usage:       "Username",
							Destination: &loginOptions.Username,
						},
						cli.StringFlag{
							Name:        "password, p",
							Usage:       "Password",
							Destination: &loginOptions.Password,
						},
					},
					Action: func(c *cli.Context) error {
						// XXX this configuration needs to be reworked
						config, err := conjurapi.LoadConfig()
						if err != nil {
							return err
						}

						client, _ := conjurapi.NewClientFromKey(config, authn.LoginPair{})
						loginer := cmd.NewAuthnLoginer(client, fs)

						return loginer.Do(loginOptions)
					},
				},
				{
					Name:   "whoami",
					Usage:  "Prints out the current logged in username",
					Before: initFunc,
					Action: func(c *cli.Context) (err error) {
						whoamier := cmd.NewWhoAmIer(getClient(c))

						resp, err := whoamier.Do()
						if err != nil {
							return err
						}
						fmt.Println(string(resp))
						return nil
					},
				},
			},
		},
	}
}
