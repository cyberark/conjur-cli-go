package main

import (
	"github.com/urfave/cli"
)

var ResourceCommands = []cli.Command{
	{
		Name:  "list",
		Usage: "List objects",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "inspect",
				Usage: "Show full object information",
			},
			cli.StringFlag{
				Name:  "kind, k",
				Usage: "Filter by kind",
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
			return nil
		},
	},
	{
		Name:  "show",
		Usage: "Show an object",
	},
}
