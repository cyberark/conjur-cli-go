package main

import (
	"os"
	"sort"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/internal/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

type commandFactory func(api cmd.ConjurClient, fs afero.Fs) []cli.Command

func run() (err error) {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Usage = "A CLI for Conjur"

	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true, DisableLevelTruncation: true})

	config := conjurapi.LoadConfig()

	client, err := conjurapi.NewClientFromEnvironment(config)
	if err != nil {
		return
	}

	api := cmd.ConjurClient(client)
	fs := afero.NewOsFs()
	commandFactories := []commandFactory{
		AuthnCommands,
		InitCommands,
		PolicyCommands,
		ResourceCommands,
		VariableCommands,
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
