package main

import (
	"github.com/spf13/afero"
	"github.com/urfave/cli"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
)

// InitCommands contains the definition of the command for initialization.
var InitCommands = func(initFunc cli.BeforeFunc, fs afero.Fs) []cli.Command {
	var options cmd.InitOptions

	return []cli.Command{
		{
			Name:  "init",
			Usage: "Initialize the Conjur configuration",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "account, a",
					Usage:       "Conjur organization account name",
					Destination: &options.Account,
				},
				cli.StringFlag{
					Name:        "certificate, c",
					Usage:       "Conjur SSL certificate (will be obtained from host unless provided by this option)",
					Destination: &options.Certificate,
				},
				cli.StringFlag{
					Name:        "file, f",
					Usage:       "File to write the configuration to",
					Destination: &options.File,
				},
				cli.BoolFlag{
					Name:        "force",
					Usage:       "Force overwrite of existing files",
					Destination: &options.Force,
				},
				cli.StringFlag{
					Name:        "url, u",
					Usage:       "URL of the Conjur service",
					Destination: &options.URL,
				},
			},
			// don't set Before: here, we're writing the API configuration
			Action: func(c *cli.Context) error {
				initer := cmd.NewIniter(fs)

				return initer.Do(options)
			},
		},
	}
}
