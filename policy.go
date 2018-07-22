package main

import (
	"github.com/urfave/cli"
)

var loadContext struct {
	delete   bool
	replace  bool
	policy   string
	filename string
}

var PolicyCommands = cli.Command{
	Name:  "policy",
	Usage: "Manage policies",
	Subcommands: []cli.Command{
		{
			Name:      "load",
			Usage:     "Load a policy",
			ArgsUsage: "POLICY FILENAME",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "delete",
					Usage:       "Allow explicit deletion statements in the policy.",
					Destination: &loadContext.delete,
				},
				cli.BoolFlag{
					Name:        "replace",
					Usage:       "Fully replace the existing policy, deleting any data that is not declared in the new policy.",
					Destination: &loadContext.replace,
				},
			},
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	},
}
