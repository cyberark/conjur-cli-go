package main

import (
	"github.com/urfave/cli"
)

// AuthnCommands contains the definitions of the commands related to authentication.
var AuthnCommands = []cli.Command{
	{
		Name:  "authn",
		Usage: "Login and logout",
		Subcommands: []cli.Command{
			{
				Name:  "authenticate",
				Usage: "Obtains an authentication token using the current logged-in user",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:  "login",
				Usage: "Logs in and caches credentials to netrc",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:  "whoami",
				Usage: "Prints out the current logged in username",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
		},
	},
}
