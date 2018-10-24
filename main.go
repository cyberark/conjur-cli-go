package main

import (
	"os"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/urfave/cli"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

type commandFactory func(initFunc cli.BeforeFunc, fs afero.Fs) []cli.Command

func run() (err error) {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Usage = "A CLI for Conjur"

	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true, DisableLevelTruncation: true})

	commandFactories := []commandFactory{
		AuthnCommands,
		InitCommands,
		PolicyCommands,
		ResourceCommands,
		VariableCommands,
	}

	initFunc := func(c *cli.Context) error {
		config, err := conjurapi.LoadConfig()
		if err != nil {
			return err
		}
		c.App.Metadata["config"] = config
		logrus.Infof("%v", c.App.Metadata["config"])

		if client, err := conjurapi.NewClientFromEnvironment(config); err == nil {
			c.App.Metadata["client"] = client
		} else {
			// No defer funcs should be set to run yet, so it's ok to call
			// logrus.Fatal here. Also, returning an error from a BeforeFunc
			// causes urfave/cli to print the help text, which doesn't make
			// sense if initialization fails.
			logrus.Fatal(err)
		}

		return nil
	}
	fs := afero.NewOsFs()
	for _, factory := range commandFactories {
		app.Commands = append(app.Commands, factory(initFunc, fs)...)
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	return app.Run(os.Args)
}

func main() {
	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}
