package main

import (
	"github.com/urfave/cli"
)

var valueContext struct {
	name  string
	value string
}

var VariableCommands = cli.Command{
	Name:  "variable",
	Usage: "Manage variables",
	Subcommands: []cli.Command{
		{
			Name:  "value",
			Usage: "get value for variable",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name:  "values",
			Usage: "operate on variable values",
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "add value for a variable",
					Action: func(c *cli.Context) error {
						return nil
					},
				},
			},
		},
	},
}
