package main

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/urfave/cli"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
)

// PolicyCommands contains the definitions of the commands related to policies.
var PolicyCommands = func(client cmd.ConjurClient, fs afero.Fs) []cli.Command {
	var options cmd.PolicyLoadOptions

	return []cli.Command{
		{
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
							Destination: &options.Delete,
						},
						cli.BoolFlag{
							Name:        "replace",
							Usage:       "Fully replace the existing policy, deleting any data that is not declared in the new policy.",
							Destination: &options.Replace,
						},
					},
					Action: func(c *cli.Context) error {
						options.PolicyID = c.Args().Get(0)
						options.Filename = c.Args().Get(1)

						loader := cmd.NewPolicyLoader(client.(cmd.PolicyClient), fs)

						resp, err := loader.Do(options)
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
