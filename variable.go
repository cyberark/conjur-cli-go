package main

import (
	"os"

	"github.com/cyberark/conjur-cli-go/action"
	"github.com/urfave/cli"
)

// VariableCommands contains the definitions of the commands related to variables.
var VariableCommands = []cli.Command{
	{
		Name:  "variable",
		Usage: "Manage variables",
		Subcommands: []cli.Command{
			{
				Name:  "value",
				Usage: "get value for variable",
				Action: func(c *cli.Context) error {
					name := c.Args().Get(0)
					value, err := action.Variable{Name: name}.Value(AppClient(c.App))
					if err != nil {
						return err
					}
					os.Stdout.Write(value)
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
							name := c.Args().Get(0)
							value := c.Args().Get(1)
							return action.Variable{Name: name}.ValuesAdd(AppClient(c.App), value)
						},
					},
				},
			},
		},
	},
}
