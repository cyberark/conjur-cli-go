package main

import (
	"github.com/urfave/cli"
)

// InitCommand contains the definition of the command for initialization.
var InitCommand = cli.Command{
	Name:  "init",
	Usage: "Initialize the Conjur configuration",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "account, a",
			Usage: "Conjur organization account name",
		},
		cli.StringFlag{
			Name:  "certificate, c",
			Usage: "Conjur SSL certificate (will be obtained from host unless provided by this option)",
		},
		cli.StringFlag{
			Name:  "file, f",
			Usage: "File to write the configuration to",
		},
		cli.BoolFlag{
			Name:  "force",
			Usage: "Force overwrite of existing files",
		},
		cli.StringFlag{
			Name:  "url, u",
			Usage: "URL of the Conjur service",
		},
	},
	Action: func(c *cli.Context) error {
		return nil
	},
}
