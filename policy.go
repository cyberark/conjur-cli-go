package main

import (
	"fmt"

	"github.com/cyberark/conjur-cli-go/action"
	"github.com/urfave/cli"
)

var loadContext struct {
	delete  bool
	replace bool
}

// PolicyCommands contains the definitions of the commands related to policies.
var PolicyCommands = []cli.Command{
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
						Destination: &loadContext.delete,
					},
					cli.BoolFlag{
						Name:        "replace",
						Usage:       "Fully replace the existing policy, deleting any data that is not declared in the new policy.",
						Destination: &loadContext.replace,
					},
				},
				Action: func(c *cli.Context) error {
					policyID := c.Args().Get(0)
					filename := c.Args().Get(1)
					resp, err := action.Policy{PolicyID: policyID}.Load(AppClient(c.App), filename, loadContext.delete, loadContext.delete)
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
