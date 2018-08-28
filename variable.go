package main

import (
	"os"

	"github.com/spf13/afero"
	"github.com/urfave/cli"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
)

// VariableCommands contains the definitions of the commands related to variables.
var VariableCommands = func(initFunc cli.BeforeFunc, fs afero.Fs) []cli.Command {
	getClient := func(ctx *cli.Context) cmd.VariableClient {
		return ctx.App.Metadata["client"].(cmd.VariableClient)
	}

	return []cli.Command{
		{
			Name:  "variable",
			Usage: "Manage variables",
			Subcommands: []cli.Command{
				{
					Name:   "value",
					Usage:  "get value for variable",
					Before: initFunc,
					Action: func(c *cli.Context) error {
						options := cmd.VariableValueOptions{
							ID: c.Args().Get(0),
						}
						valuer := cmd.NewVariableValuer(getClient(c))

						value, err := valuer.Do(options)
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
							Name:   "add",
							Usage:  "add value for a variable",
							Before: initFunc,
							Action: func(c *cli.Context) error {
								options := cmd.VariableValuesAddOptions{
									ID:    c.Args().Get(0),
									Value: c.Args().Get(1),
								}

								adder := cmd.NewVariableValuesAdder(getClient(c))

								return adder.Do(options)
							},
						},
					},
				},
			},
		},
	}
}
