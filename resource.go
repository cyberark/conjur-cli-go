package main

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/urfave/cli"

	"github.com/cyberark/conjur-cli-go/internal/cmd"
)

// ResourceCommands contains the definitions of commands related to Conjur resources.
var ResourceCommands = func(initFunc cli.BeforeFunc, fs afero.Fs) []cli.Command {
	var listOptions cmd.ResourceListOptions
	getClient := func(ctx *cli.Context) cmd.ResourceClient {
		return ctx.App.Metadata["client"].(cmd.ResourceClient)
	}

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
			Before: initFunc,
			Action: func(c *cli.Context) error {
				lister := cmd.NewResourceLister(getClient(c))

				resp, err := lister.Do(listOptions)
				if err != nil {
					return err
				}

				fmt.Println(string(resp))
				return nil
			},
		},
		{
			Name:   "show",
			Usage:  "Show an object",
			Before: initFunc,
			Action: func(c *cli.Context) error {
				showOptions := cmd.ResourceShowOptions{
					ResourceID: c.Args().Get(0),
				}

				shower := cmd.NewResourceShower(getClient(c))

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
