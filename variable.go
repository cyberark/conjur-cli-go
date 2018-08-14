package main

import (
	"os"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

// VariableCommands contains the definitions of the commands related to variables.
var VariableCommands = func(client cmd.ConjurClient, fs afero.Fs) []cli.Command {
	variableClient := client.(cmd.VariableClient)
	return []cli.Command{
		{
			Name:  "variable",
			Usage: "Manage variables",
			Subcommands: []cli.Command{
				{
					Name:  "value",
					Usage: "get value for variable",
					Action: func(c *cli.Context) error {
						options := cmd.VariableValueOptions{
							ID: c.Args().Get(0),
						}
						valuer := cmd.NewVariableValuer(variableClient)

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
							Name:  "add",
							Usage: "add value for a variable",
							Action: func(c *cli.Context) error {
								options := cmd.VariableValuesAddOptions{
									ID:    c.Args().Get(0),
									Value: c.Args().Get(1),
								}

								adder := cmd.NewVariableValuesAdder(variableClient)

								return adder.Do(options)
							},
						},
					},
				},
			},
		},
	}
}
