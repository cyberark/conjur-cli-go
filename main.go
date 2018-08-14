package main

import (
	"os"
	"sort"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-cli-go/internal/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

type CommandFactory func(api cmd.ConjurClient, fs afero.Fs) []cli.Command

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Usage = "A CLI for Conjur"

	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true, DisableLevelTruncation: true})

	config := conjurapi.LoadConfig()

	client, err := conjurapi.NewClientFromEnvironment(config)
	if err != nil {
		log.Errorf("Failed creating a Conjur client: %s\n", err.Error())
		os.Exit(1)
	}

	api := cmd.ConjurClient(client)
	fs := afero.NewOsFs()
	commandFactories := []CommandFactory{
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

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
