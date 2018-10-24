package main

import (
	"github.com/spf13/afero"
	"github.com/urfave/cli"

	"github.com/cyberark/conjur-api-go/conjurapi"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
)

// AuthnCommands contains the definitions of the commands related to authentication.
var AuthnCommands = func(initFunc cli.BeforeFunc, fs afero.Fs) []cli.Command {
	var loginOptions cmd.AuthnLoginOptions
	getClient := func(ctx *cli.Context) cmd.AuthnClient {
		return ctx.App.Metadata["client"].(cmd.AuthnClient)
	}

	return []cli.Command{
		{
			Name:  "authn",
			Usage: "Login and logout",
			Subcommands: []cli.Command{
				{
					Name:   "authenticate",
					Usage:  "Obtains an authentication token using the current logged-in user",
					Before: initFunc,
					Action: func(c *cli.Context) error {
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
					Before: initFunc,
					Action: func(c *cli.Context) error {
						loginer := cmd.NewAuthnLoginer(getClient(c), fs)

						loginOptions.Config = c.App.Metadata["config"].(conjurapi.Config)
						return loginer.Do(loginOptions)
					},
				},
				{
					Name:   "whoami",
					Usage:  "Prints out the current logged in username",
					Before: initFunc,
					Action: func(c *cli.Context) error {
						return nil
					},
				},
			},
		},
	}
}
