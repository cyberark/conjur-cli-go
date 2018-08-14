package main

import (
	"fmt"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

// ResourceCommands contains the definitions of commands related to Conjur resources.
var ResourceCommands = func(client cmd.ConjurClient, fs afero.Fs) []cli.Command {
	var listOptions cmd.ResourceListOptions

	return []cli.Command{
		{
			Name:  "list",
			Usage: "List objects",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "inspect, i",
					Usage:       "Show full object information",
					Destination: &listOptions.Inspect,
				},
				cli.StringFlag{
					Name:        "kind, k",
					Usage:       "Filter by kind",
					Destination: &listOptions.Kind,
				},
				cli.UintFlag{
					Name:  "limit, l",
					Usage: "Maximum number of records to return",
				},
				cli.UintFlag{
					Name:  "offset, o",
					Usage: "Offset to start from",
				},
				cli.BoolFlag{
					Name:  "raw, r",
					Usage: "Show annotations in 'raw' format",
				},
				cli.StringFlag{
					Name:  "role",
					Usage: "Role to act as. By default, the current logged-in role is used",
				},
				cli.StringFlag{
					Name:  "search, s",
					Usage: "Full-text search on resource id and annotation values",
				},
			},
			Action: func(c *cli.Context) error {
				lister := cmd.NewResourceLister(client.(cmd.ResourceClient))

				resp, err := lister.Do(listOptions)
				if err != nil {
					return err
				}

				fmt.Println(string(resp))
				return nil
			},
		},
		{
			Name:  "show",
			Usage: "Show an object",
			Action: func(c *cli.Context) error {
				showOptions := cmd.ResourceShowOptions{
					ResourceID: c.Args().Get(0),
				}

				shower := cmd.NewResourceShower(client.(cmd.ResourceClient))

				resp, err := shower.Do(showOptions)
				if err != nil {
					return err
				}

				fmt.Println(string(resp))
				return nil
			},
		},
	}
}
