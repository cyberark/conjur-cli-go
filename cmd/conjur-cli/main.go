package main

import (
	"os"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/urfave/cli"

	"github.com/cyberark/conjur-api-go/conjurapi"

	cliCommands "github.com/cyberark/conjur-cli-go/internal/cli"
	"github.com/cyberark/conjur-cli-go/internal/cmd"
)

type commandFactory func(api cmd.ConjurClient, fs afero.Fs) []cli.Command

func run() error {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Usage = "A CLI for Conjur"

	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true, DisableLevelTruncation: true})

	config, err := conjurapi.LoadConfig()
	if err != nil {
		return err
	}

	client, err := conjurapi.NewClientFromEnvironment(config)
	if err != nil {
		return err
	}

	api := cmd.ConjurClient(client)
	fs := afero.NewOsFs()
	commandFactories := []commandFactory{
		cliCommands.AuthnCommands,
		cliCommands.InitCommands,
		cliCommands.PolicyCommands,
		cliCommands.ResourceCommands,
		cliCommands.VariableCommands,
	}

	for _, factory := range commandFactories {
		app.Commands = append(app.Commands, factory(api, fs)...)
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	return app.Run(os.Args)
}

func main() {
	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}
